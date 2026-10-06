package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSearchAndFetch(t *testing.T) {
	var methods, names []string
	var sessionID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != userAgent || r.Header.Get("Authorization") != "" {
			t.Error("expected project User-Agent and anonymous requests")
		}
		var request struct {
			ID     int                    `json:"id"`
			Method string                 `json:"method"`
			Params map[string]interface{} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		methods = append(methods, request.Method)
		var result interface{}
		switch request.Method {
		case "initialize":
			result = map[string]interface{}{"protocolVersion": "2024-11-05", "capabilities": map[string]interface{}{}, "serverInfo": map[string]string{"name": "fixture", "version": "1"}}
		case "notifications/initialized":
			w.WriteHeader(http.StatusAccepted)
			return
		case "tools/list":
			name := "web_search"
			if request.Params["cursor"] == "fetch" {
				name = "web_fetch"
			}
			result = map[string]interface{}{"tools": []interface{}{map[string]interface{}{"name": name, "inputSchema": map[string]interface{}{"type": "object"}}}}
			if name == "web_search" {
				result.(map[string]interface{})["nextCursor"] = "fetch"
			}
		case "tools/call":
			name := request.Params["name"].(string)
			names = append(names, name)
			args := request.Params["arguments"].(map[string]interface{})
			id, _ := args["session_id"].(string)
			if len(id) != 32 {
				t.Error("expected random session identifier")
			}
			if name == "web_search" {
				sessionID = id
				if args["objective"] != "Go context cancellation" || args["search_queries"].([]interface{})[0] != "Go context cancellation" {
					t.Error("search arguments lost")
				}
			} else {
				if id != sessionID || args["urls"].([]interface{})[0] != "https://go.dev/blog/context" || args["search_queries"].([]interface{})[0] != "Go context cancellation" {
					t.Error("fetch arguments or shared session lost")
				}
			}
			result = map[string]interface{}{"content": []interface{}{map[string]string{"type": "text", "text": "https://go.dev/blog/context: cancellation propagates to child contexts"}}}
		default:
			t.Errorf("unexpected method %s", request.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": request.ID, "result": result}); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	config := parallelConfig()
	config.URL = server.URL
	var out bytes.Buffer
	if err := run(context.Background(), &out, config, "Go context cancellation", "https://go.dev/blog/context"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(names, ",") != "web_search,web_fetch" || strings.Join(methods, ",") != "initialize,notifications/initialized,tools/list,tools/list,tools/call,tools/call" {
		t.Fatalf("unexpected calls: %v / %v", methods, names)
	}
	if !strings.Contains(out.String(), "web_search:") || !strings.Contains(out.String(), "web_fetch:") || strings.Count(out.String(), "cancellation propagates") != 2 {
		t.Fatalf("missing useful output: %s", out.String())
	}
}

func TestDiscoveryFailureAndCancellation(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		t.Run(fmt.Sprintf("cancel=%t", cancel), func(t *testing.T) {
			release := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if cancel {
					<-release
					return
				}
				http.Error(w, "rate limit", http.StatusTooManyRequests)
			}))
			defer server.Close()
			defer close(release)
			config := parallelConfig()
			config.URL = server.URL
			ctx, stop := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer stop()
			err := run(ctx, io.Discard, config, "Go context cancellation", "")
			if err == nil || (!cancel && !strings.Contains(err.Error(), "429")) || (cancel && ctx.Err() == nil) {
				t.Fatalf("expected propagated failure, got %v", err)
			}
		})
	}
}

type observedTransport struct {
	base http.RoundTripper
	t    *testing.T
}

func (o observedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header.Get("User-Agent") != userAgent || r.Header.Get("Authorization") != "" {
		o.t.Error("expected project User-Agent and no Authorization on live request")
	}
	var body []byte
	if r.Body != nil {
		var err error
		body, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
	}
	o.t.Logf("Live request: %s %s User-Agent=%s Authorization absent=%t body=%s", r.Method, r.URL, r.Header.Get("User-Agent"), r.Header.Get("Authorization") == "", body)
	return o.base.RoundTrip(r)
}

func TestParallelLive(t *testing.T) {
	if os.Getenv("PARALLEL_SEARCH_LIVE") != "1" {
		t.Skip("set PARALLEL_SEARCH_LIVE=1 to use the anonymous endpoint")
	}
	config := parallelConfig()
	config.HTTPClient = &http.Client{Transport: observedTransport{base: http.DefaultTransport, t: t}, Timeout: 120 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	var out bytes.Buffer
	if err := run(ctx, &out, config, "Go context cancellation documentation", "https://go.dev/blog/context"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "web_search:") || !strings.Contains(out.String(), "web_fetch:") || !strings.Contains(out.String(), "https://go.dev/blog/context") || !strings.Contains(strings.ToLower(out.String()), "cancel") {
		t.Fatalf("missing useful search/fetch output: %s", out.String())
	}
	t.Log(out.String())
}
