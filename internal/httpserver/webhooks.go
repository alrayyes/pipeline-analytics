package httpserver

import (
	"context"
	"encoding/json"
	"io"
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "could not read request body")

		return
	}

	var payload struct {
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_signature", "signature verification failed")

		return
	}

	repo, err := h.store.RepoByIdentifier(r.Context(), forge, payload.Repository.FullName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_signature", "signature verification failed")

		return
	}

	if !ingestion.VerifySignature(repo.WebhookSecret, body, r.Header.Get(signatureHeader)) {
		writeError(w, http.StatusUnauthorized, "invalid_signature", "signature verification failed")

		return
	}

	if err := process(r.Context(), h.store, repo, r.Header.Get(eventHeader), body); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "process webhook event")

		return
	}

	w.WriteHeader(http.StatusAccepted)
}
