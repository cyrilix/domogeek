package main

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPEndToEnd(t *testing.T) {
	withTestCalendar(t)

	srv := httptest.NewServer(newMCPHandler())
	defer srv.Close()

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: srv.URL}, nil)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools failed: %v", err)
	}
	if len(tools.Tools) != 3 {
		t.Fatalf("got %d tools, want 3", len(tools.Tools))
	}
	for _, tool := range tools.Tools {
		t.Logf("tool: %s - %s", tool.Name, tool.Description)
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_calendar_for_date",
		Arguments: map[string]any{"date": "2020-01-01"},
	})
	if err != nil {
		t.Fatalf("call tool failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool call returned an error result: %v", result.Content)
	}
	t.Logf("structured content: %+v", result.StructuredContent)

	holidays, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_holidays",
		Arguments: map[string]any{"year": 2020},
	})
	if err != nil {
		t.Fatalf("call get_holidays failed: %v", err)
	}
	t.Logf("holidays structured content: %+v", holidays.StructuredContent)

	badDate, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get_calendar_for_date",
		Arguments: map[string]any{"date": "nope"},
	})
	if err != nil {
		t.Fatalf("call with bad date should be a tool error, not a protocol error: %v", err)
	}
	if !badDate.IsError {
		t.Fatalf("expected IsError=true for an invalid date")
	}
}
