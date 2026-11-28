package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMCPAcrossReplicas reproduces the Kubernetes deployment: two replicas, each
// with its own handler (no shared state), behind a Service that spreads requests
// round-robin. A session-based handler fails here ("session not found" when a
// request lands on the replica that did not see the initialize), so the
// handler must be stateless.
func TestMCPAcrossReplicas(t *testing.T) {
	withTestCalendar(t)

	replicas := []http.Handler{newMCPHandler(), newMCPHandler()}
	var next atomic.Uint64
	lb := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		replicas[next.Add(1)%uint64(len(replicas))].ServeHTTP(w, r)
	}))
	defer lb.Close()

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: lb.URL}, nil)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer session.Close()

	// Several calls so that both replicas serve some of them.
	for i := 0; i < 6; i++ {
		tools, err := session.ListTools(ctx, nil)
		if err != nil {
			t.Fatalf("list tools (call %d) failed: %v", i, err)
		}
		if len(tools.Tools) != 4 {
			t.Fatalf("call %d: got %d tools, want 4", i, len(tools.Tools))
		}
		res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "get_current_datetime"})
		if err != nil {
			t.Fatalf("call tool (call %d) failed: %v", i, err)
		}
		if res.IsError {
			t.Fatalf("call %d: tool returned an error result: %v", i, res.Content)
		}
	}
}
