package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// Handler serves minimal admin HTTP endpoints for draft moderation.
type Handler struct {
	drafts domain.DraftRepository
}

// NewHandler creates admin HTTP handler.
func NewHandler(drafts domain.DraftRepository) (*Handler, error) {
	if drafts == nil {
		return nil, fmt.Errorf("draft repository is nil")
	}
	return &Handler{drafts: drafts}, nil
}

// Register wires admin routes into provided mux.
func (h *Handler) Register(mux *http.ServeMux) error {
	if h == nil {
		return fmt.Errorf("admin handler is nil")
	}
	if mux == nil {
		return fmt.Errorf("mux is nil")
	}

	mux.HandleFunc("GET /admin/drafts", h.handleListDrafts)
	mux.HandleFunc("POST /admin/drafts/", h.handleDraftAction)
	return nil
}

func (h *Handler) handleListDrafts(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.drafts == nil {
		http.Error(w, "admin handler is unavailable", http.StatusInternalServerError)
		return
	}

	status, err := parseStatus(r.URL.Query().Get("status"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	drafts, err := h.drafts.ListByStatus(r.Context(), status, limit)
	if err != nil {
		http.Error(w, "failed to list drafts", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, drafts)
}

func (h *Handler) handleDraftAction(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.drafts == nil {
		http.Error(w, "admin handler is unavailable", http.StatusInternalServerError)
		return
	}

	draftID, action, err := parseDraftActionPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	status, ok := actionToStatus(action)
	if !ok {
		http.Error(w, "unsupported draft action", http.StatusNotFound)
		return
	}

	if err := h.drafts.UpdateStatus(r.Context(), draftID, status); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "draft not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update draft status", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"id": draftID, "status": status})
}

func parseStatus(raw string) (domain.DraftStatus, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return domain.DraftStatusPending, nil
	}

	status := domain.DraftStatus(raw)
	switch status {
	case domain.DraftStatusPending, domain.DraftStatusApproved, domain.DraftStatusRejected, domain.DraftStatusPosted:
		return status, nil
	default:
		return "", fmt.Errorf("invalid draft status")
	}
}

func parseLimit(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultListLimit, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid limit")
	}
	if limit <= 0 {
		return 0, fmt.Errorf("invalid limit")
	}
	if limit > maxListLimit {
		return maxListLimit, nil
	}
	return limit, nil
}

func parseDraftActionPath(path string) (int64, string, error) {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 4 || parts[0] != "admin" || parts[1] != "drafts" {
		return 0, "", fmt.Errorf("invalid draft action path")
	}
	id, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || id <= 0 {
		return 0, "", fmt.Errorf("invalid draft id")
	}
	return id, parts[3], nil
}

func actionToStatus(action string) (domain.DraftStatus, bool) {
	switch strings.TrimSpace(action) {
	case "approve":
		return domain.DraftStatusApproved, true
	case "reject":
		return domain.DraftStatusRejected, true
	default:
		return "", false
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
