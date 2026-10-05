package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The webhook endpoint is public and the body is the sender's. A valid
// signature still only proves who sent it, so the body is signed here and
// reaches the decoders. Nothing may panic; the status code is free to be
// anything the handler chooses.
func FuzzWebhookReceiver(f *testing.F) {
	f.Add("github", "workflow_run", webhookTestPayload)
	f.Add("github", "workflow_job", webhookTestPayload)
	f.Add("github", "workflow_run", `{"repository":{"full_name":"alrayyes/pipeline-analytics"},"workflow_run":{"id":"x"}}`)
	f.Add("forgejo", "push", `{"repository":{"full_name":"alrayyes/pipeline-analytics"}}`)
	f.Add("github", "", `not json`)

	f.Fuzz(func(t *testing.T, forge, eventType, body string) {
		if forge != "github" && forge != "forgejo" {
			t.Skip()
		}

		srv := newTestServerWithReconciler(t, &fakeReconciler{})
		secret := registerTestRepo(t, srv, forge)

		req := httptest.NewRequest(http.MethodPost, "/webhooks/"+forge, strings.NewReader(body))
		eventHeader, signatureHeader := "X-GitHub-Event", "X-Hub-Signature-256"
		if forge == "forgejo" {
			eventHeader, signatureHeader = "X-Forgejo-Event", "X-Forgejo-Signature"
		}

		req.Header.Set(eventHeader, eventType)
		req.Header.Set(signatureHeader, signBody(t, secret, body))
		srv.ServeHTTP(httptest.NewRecorder(), req)
	})
}

// The MCP endpoint parses JSON-RPC from the request body. It sits behind a
// session, so the request carries one: the fuzzer's reach is the decoder and
// the tool handlers, not the login check.
func FuzzMCPRequest(f *testing.F) {
	f.Add(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"fuzz","version":"0"}}}`)
	f.Add(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	f.Add(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_pipelines","arguments":{"window":"7d"}}}`)
	f.Add(`[{"jsonrpc":"2.0","id":4,"method":"tools/list"}]`)
	f.Add(`{"jsonrpc":"2.0","method":"tools/call","params":null}`)
	f.Add(`not json`)

	f.Fuzz(func(t *testing.T, body string) {
		srv := newTestServerWithAuth(t)

		req := httptest.NewRequest(http.MethodPost, "/api/mcp", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.AddCookie(srv.sessionCookie)
		srv.ServeHTTP(httptest.NewRecorder(), req)
	})
}
