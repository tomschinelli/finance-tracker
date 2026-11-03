package api

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"finance-backend/internal/store"
)

type recurringResponse struct {
	ID             int64    `json:"id"`
	Label          string   `json:"label"`
	CategoryID     int64    `json:"category_id"`
	AccountID      int64    `json:"account_id"`
	AmountCents    int64    `json:"amount_cents"`
	IntervalMonths int32    `json:"interval_months"`
	AnchorDate     string   `json:"anchor_date"`
	Active         bool     `json:"active"`
	Tags           []string `json:"tags"`
}

func toRecurringResponse(d store.RecurringDefinition) recurringResponse {
	return recurringResponse{
		ID:             d.ID,
		Label:          d.Label,
		CategoryID:     d.CategoryID,
		AccountID:      d.AccountID,
		AmountCents:    d.AmountCents,
		IntervalMonths: d.IntervalMonths,
		AnchorDate:     formatDate(d.AnchorDate),
		Active:         d.Active,
		Tags:           d.Tags,
	}
}

func (s *Server) listRecurring(w http.ResponseWriter, r *http.Request) {
	var (
		defs []store.RecurringDefinition
		err  error
	)
	if tag := r.URL.Query().Get("tag"); tag != "" {
		defs, err = s.q.ListRecurringDefinitionsByTag(r.Context(), tag)
	} else {
		defs, err = s.q.ListRecurringDefinitions(r.Context())
	}
	if err != nil {
		handleDBError(w, err)
		return
	}
	out := make([]recurringResponse, 0, len(defs))
	for _, d := range defs {
		out = append(out, toRecurringResponse(d))
	}
	writeJSON(w, http.StatusOK, out)
}

type createRecurringRequest struct {
	Label          string   `json:"label"`
	CategoryID     int64    `json:"category_id"`
	AccountID      int64    `json:"account_id"`
	AmountCents    int64    `json:"amount_cents"`
	IntervalMonths int32    `json:"interval_months"`
	AnchorDate     string   `json:"anchor_date"`
	Active         *bool    `json:"active"`
	Tags           []string `json:"tags"`
}

func (s *Server) createRecurring(w http.ResponseWriter, r *http.Request) {
	var req createRecurringRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Label == "" {
		writeError(w, http.StatusBadRequest, "label is required")
		return
	}
	if req.IntervalMonths < 1 {
		writeError(w, http.StatusBadRequest, "interval_months must be at least 1 (1 = monthly, 12 = yearly, or any custom interval)")
		return
	}

	anchor, err := parseDate(req.AnchorDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid anchor_date")
		return
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	def, err := s.q.CreateRecurringDefinition(r.Context(), store.CreateRecurringDefinitionParams{
		Label:          req.Label,
		CategoryID:     req.CategoryID,
		AccountID:      req.AccountID,
		AmountCents:    req.AmountCents,
		IntervalMonths: req.IntervalMonths,
		AnchorDate:     anchor,
		Active:         active,
		Tags:           tags,
	})
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toRecurringResponse(def))
}

type patchRecurringRequest struct {
	AmountCents *int64    `json:"amount_cents"`
	AccountID   *int64    `json:"account_id"`
	Active      *bool     `json:"active"`
	Tags        *[]string `json:"tags"`
}

func (s *Server) updateRecurring(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req patchRecurringRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	params := store.UpdateRecurringDefinitionParams{ID: id}
	if req.AmountCents != nil {
		params.AmountCents = pgtype.Int8{Int64: *req.AmountCents, Valid: true}
	}
	if req.AccountID != nil {
		params.AccountID = pgtype.Int8{Int64: *req.AccountID, Valid: true}
	}
	if req.Active != nil {
		params.Active = pgtype.Bool{Bool: *req.Active, Valid: true}
	}
	if req.Tags != nil {
		params.Tags = *req.Tags
	}

	def, err := s.q.UpdateRecurringDefinition(r.Context(), params)
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toRecurringResponse(def))
}
