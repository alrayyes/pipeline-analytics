package ingestion_test

import (
	"context"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

// A webhook body is attacker-controlled until its signature checks out, and
// even a valid signature only proves who sent it. Neither function below may
// panic on any input (rules/go-test.md, native fuzzing).

func FuzzVerifySignature(f *testing.F) {
	body := []byte(`{"hello":"world"}`)

	f.Add("shh", body, sign("shh", body))
	f.Add("shh", body, "sha256=")
	f.Add("shh", body, "not hex")
	f.Add("", []byte(nil), "")

	f.Fuzz(func(_ *testing.T, secret string, body []byte, signature string) {
		ingestion.VerifySignature(secret, body, signature)
	})
}

func FuzzProcessGitHubEvent(f *testing.F) {
	f.Add("workflow_run", []byte(githubWorkflowRunPayload))
	f.Add("workflow_job", []byte(githubWorkflowJobPayload))
	f.Add("workflow_run", []byte(`{"workflow_run":{"id":"not a number"}}`))
	f.Add("workflow_job", []byte(`{"workflow_job":{"steps":[null]}}`))
	f.Add("ping", []byte(`not json`))

	f.Fuzz(func(_ *testing.T, eventType string, payload []byte) {
		_ = ingestion.ProcessGitHubEvent(context.Background(), newFakeRunStore(), ingestion.Repo{ID: "repo-1"}, eventType, payload)
	})
}
