# Parallel Search MCP example

Search the web and extract page excerpts with the SDK's HTTP MCP client. This
example discovers `web_search` and `web_fetch` through `ListTools` and executes
the returned `tool.Tool` values directly, so no LLM or API keys are needed.

From the repository root, with Go installed:

```bash
go run ./examples/parallel_search_example

go run ./examples/parallel_search_example \
  -query 'Go context cancellation documentation' \
  -fetch https://go.dev/blog/context
```

`-query` sets the search objective and keyword query. `-fetch` optionally extracts
excerpts from a specific URL after the search. Output includes the server's
results, source URLs, and any warnings or extraction errors. The example has a
120-second overall deadline and Ctrl+C cancels in-flight requests.

The endpoint is `https://search.parallel.ai/mcp`. Requests identify the example
with `User-Agent: agent-sdk-go/parallel-search-example` and send no Authorization
header. The example does not read environment keys or saved credentials. A random
session identifier is shared by search and fetch within each run.

[Parallel's Search MCP documentation](https://docs.parallel.ai/integrations/mcp/search-mcp)
describes the anonymous free tier for exploration and light use. It has rate
limits, uses `fast` search mode, and caps excerpts to roughly 25,000 characters
per call. Free access does not imply unlimited usage.

## Attach the tools to an agent

The same connection can supply an agent's tools using the existing SDK API:

```go
client, err := mcp.NewHTTPClient(mcp.HTTPConfig{
    URL: "https://search.parallel.ai/mcp",
    Headers: map[string]string{
        "User-Agent": "agent-sdk-go/parallel-search-example",
    },
})
if err != nil {
    return err
}
defer func() { _ = client.Close() }()

tools, err := client.ListTools(ctx)
if err != nil {
    return err
}
assistant.WithTools(tools...)
```

Configure your model provider and runner as in
[the MCP agent example](../mcp_example). Model inference is separate and may
require a provider key or a local model. Existing SDK and example defaults are
unchanged.

## Tests

```bash
go test ./examples/parallel_search_example
```

Tests use a local MCP fixture by default. To also run the opt-in anonymous live
search and fetch check (requires network access and uses the free tier):

```bash
PARALLEL_SEARCH_LIVE=1 go test -v ./examples/parallel_search_example
```
