package api

import (
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"

	"finance-backend/internal/store"
)

type paymentResponse struct {
	ID                    int64   `json:"id"`
	RecurringDefinitionID *int64  `json:"recurring_definition_id"`
	Label                 string  `json:"label"`
	CategoryID            int64   `json:"category_id"`
	AccountID             int64   `json:"account_id"`
	AmountCents           int64   `json:"amount_cents"`
	DueDate               string  `json:"due_date"`
	PaidDate              *string `json:"paid_date"`
	ExcludedFromBudget    bool    `json:"excluded_from_budget"`
}

func toPaymentResponse(p store.Payment) paymentResponse {
	var rdID *int64
	if p.RecurringDefinitionID.Valid {
		v := p.RecurringDefinitionID.Int64
		rdID = &v
	}
	return paymentResponse{
		ID:                    p.ID,
		RecurringDefinitionID: rdID,
		Label:                 p.Label,
		CategoryID:            p.CategoryID,
		AccountID:             p.AccountID,
		AmountCents:           p.AmountCents,
		DueDate:               formatDate(p.DueDate),
		PaidDate:              formatDatePtr(p.PaidDate),
		ExcludedFromBudget:    p.ExcludedFromBudget,
	}
}

func (s *Server) listPayments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	yearStr := q.Get("year")
	if yearStr == "" {
		writeError(w, http.StatusBadRequest, "year is required")
		return
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid year")
		return
	}

	var payments []store.Payment
	if monthStr := q.Get("month"); monthStr != "" {
		month, err := strconv.Atoi(monthStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid month")
			return
		}
		payments, err = s.q.ListPaymentsByYearMonth(r.Context(), store.ListPaymentsByYearMonthParams{
			Column1: int32(year),
			Column2: int32(month),
		})
		if err != nil {
			handleDBError(w, err)
			return
		}
	} else {
		payments, err = s.q.ListPaymentsByYear(r.Context(), int32(year))
		if err != nil {
			handleDBError(w, err)
			return
		}
	}

	out := make([]paymentResponse, 0, len(payments))
	for _, p := range payments {
		out = append(out, toPaymentResponse(p))
	}
	writeJSON(w, http.StatusOK, out)
}

type createPaymentRequest struct {
	Label              string  `json:"label"`
	CategoryID         int64   `json:"category_id"`
	AccountID          int64   `json:"account_id"`
	AmountCents        int64   `json:"amount_cents"`
	DueDate            string  `json:"due_date"`
	PaidDate           *string `json:"paid_date"`
	ExcludedFromBudget bool    `json:"excluded_from_budget"`
}

func (s *Server) createPayment(w http.ResponseWriter, r *http.Request) {
	var req createPaymentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Label == "" {
		writeError(w, http.StatusBadRequest, "label is required")
		return
	}

	dueDate, err := parseDate(req.DueDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid due_date")
		return
	}

	var paidDate pgtype.Date
	if req.PaidDate != nil {
		paidDate, err = parseDate(*req.PaidDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid paid_date")
			return
		}
	}

	p, err := s.q.CreatePayment(r.Context(), store.CreatePaymentParams{
		Label:              req.Label,
		CategoryID:         req.CategoryID,
		AccountID:          req.AccountID,
		AmountCents:        req.AmountCents,
		DueDate:            dueDate,
		PaidDate:           paidDate,
		ExcludedFromBudget: req.ExcludedFromBudget,
	})
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toPaymentResponse(p))
}

type patchPaymentRequest struct {
	PaidDate           *string `json:"paid_date"`
	AmountCents        *int64  `json:"amount_cents"`
	AccountID          *int64  `json:"account_id"`
	Label              *string `json:"label"`
	CategoryID         *int64  `json:"category_id"`
	ExcludedFromBudget *bool   `json:"excluded_from_budget"`
}

func (s *Server) updatePayment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req patchPaymentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	params := store.UpdatePaymentParams{ID: id}
	if req.PaidDate != nil {
		d, err := parseDate(*req.PaidDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid paid_date")
			return
		}
		params.PaidDate = d
	}
	if req.AmountCents != nil {
		params.AmountCents = pgtype.Int8{Int64: *req.AmountCents, Valid: true}
	}
	if req.AccountID != nil {
		params.AccountID = pgtype.Int8{Int64: *req.AccountID, Valid: true}
	}
	if req.Label != nil {
		params.Label = pgtype.Text{String: *req.Label, Valid: true}
	}
	if req.CategoryID != nil {
		params.CategoryID = pgtype.Int8{Int64: *req.CategoryID, Valid: true}
	}
	if req.ExcludedFromBudget != nil {
		params.ExcludedFromBudget = pgtype.Bool{Bool: *req.ExcludedFromBudget, Valid: true}
	}

	p, err := s.q.UpdatePayment(r.Context(), params)
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPaymentResponse(p))
}
