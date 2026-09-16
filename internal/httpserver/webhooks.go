package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

type webhooksHandler struct {
	store ingestion.RunStore
}

func (h *webhooksHandler) github(w http.ResponseWriter, r *http.Request) {
	h.receive(w, r, ingestion.ForgeGitHub, "X-Hub-Signature-256", "X-GitHub-Event", ingestion.ProcessGitHubEvent)
}

func (h *webhooksHandler) forgejo(w http.ResponseWriter, r *http.Request) {
	h.receive(w, r, ingestion.ForgeForgejo, "X-Forgejo-Signature", "X-Forgejo-Event", ingestion.ProcessForgejoEvent)
}

type eventProcessor func(ctx context.Context, store ingestion.RunStore, repo ingestion.Repo, eventType string, payload []byte) error

// receive handles a webhook delivery: read the body, resolve which tracked
// repo it's for, verify the signature against that repo's secret, then
// process the event. The 401 response is identical whether the repo isn't
// tracked or the signature is wrong, so a delivery can't be used to probe
// which repos exist.
func (h *webhooksHandler) receive(w http.ResponseWriter, r *http.Request, forge ingestion.Forge, signatureHeader, eventHeader string, process eventProcessor) {
	ctx := r.Context()
	eventType := r.Header.Get(eventHeader)

	slog.DebugContext(ctx, "webhook received", "forge", forge, "event", eventType)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.DebugContext(ctx, "webhook rejected", "forge", forge, "reason", "invalid_body")
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read request body")

		return
	}

	var payload struct {
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		slog.DebugContext(ctx, "webhook rejected", "forge", forge, "reason", "unparseable_payload")
		writeError(w, http.StatusUnauthorized, "invalid_signature", "signature verification failed")

		return
	}

	repo, err := h.store.RepoByIdentifier(ctx, forge, payload.Repository.FullName)
	if err != nil {
		slog.DebugContext(ctx, "webhook rejected", "forge", forge, "reason", "unknown_repo")
		writeError(w, http.StatusUnauthorized, "invalid_signature", "signature verification failed")

		return
	}

	if !ingestion.VerifySignature(repo.WebhookSecret, body, r.Header.Get(signatureHeader)) {
		slog.DebugContext(ctx, "webhook rejected", "forge", forge, "repo", repo.Identifier, "reason", "bad_signature")
		writeError(w, http.StatusUnauthorized, "invalid_signature", "signature verification failed")

		return
	}

	if err := process(ctx, h.store, repo, eventType, body); err != nil {
		slog.ErrorContext(ctx, "process webhook event", "forge", forge, "repo", repo.Identifier, "event", eventType, "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "process webhook event")

		return
	}

	slog.DebugContext(ctx, "webhook processed", "forge", forge, "repo", repo.Identifier, "event", eventType)

	w.WriteHeader(http.StatusAccepted)
}
