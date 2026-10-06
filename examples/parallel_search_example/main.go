// Example: anonymous web search and page extraction with Parallel Search MCP.
package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/mcp"
)

const (
	parallelURL = "https://search.parallel.ai/mcp"
	userAgent   = "agent-sdk-go/parallel-search-example"
)

func main() {
	query := flag.String("query", "Go context cancellation documentation", "Search objective and keyword query")
	fetchURL := flag.String("fetch", "", "Optional page URL to extract after searching")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	if err := run(ctx, os.Stdout, parallelConfig(), *query, *fetchURL); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func parallelConfig() mcp.HTTPConfig {
	return mcp.HTTPConfig{
		URL:     parallelURL,
		Headers: map[string]string{"User-Agent": userAgent},
	}
}

// run uses discovered tool.Tool values, the same tools an agent can attach with
// WithTools. Direct execution needs neither a Parallel key nor model inference.
func run(ctx context.Context, out io.Writer, config mcp.HTTPConfig, query, fetchURL string) error {
	client, err := mcp.NewHTTPClient(config)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	tools, err := client.ListTools(ctx)
	if err != nil {
		return err
	}

	// One identifier per run, reused by related search and fetch calls.
	var session [16]byte
	if _, err := rand.Read(session[:]); err != nil {
		return err
	}
	sessionID := fmt.Sprintf("%x", session)
	call := func(name string, params map[string]interface{}) error {
		params["session_id"] = sessionID
		for _, t := range tools {
			if t.GetName() == name {
				result, err := t.Execute(ctx, params)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(out, "%s:\n%v\n", name, result)
				return err
			}
		}
		return fmt.Errorf("MCP server did not expose %s", name)
	}

	if err := call("web_search", map[string]interface{}{
		"objective": query, "search_queries": []string{query},
	}); err != nil {
		return err
	}
	if fetchURL != "" {
		return call("web_fetch", map[string]interface{}{
			"urls": []string{fetchURL}, "search_queries": []string{query},
		})
	}
	return nil
}
