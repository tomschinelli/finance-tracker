package api

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"finance-backend/internal/finance"
	"finance-backend/internal/store"
)

type accountResponse struct {
	ID                   int64  `json:"id"`
	Name                 string `json:"name"`
	Currency             string `json:"currency"`
	StartingBalanceCents int64  `json:"starting_balance_cents"`
	CurrentBalanceCents  int64  `json:"current_balance_cents"`
}

func (s *Server) accountBalance(ctx context.Context, acct store.Account) (int64, error) {
	paid, err := s.q.SumPaidPaymentsByAccount(ctx, acct.ID)
	if err != nil {
		return 0, err
	}
	adj, err := s.q.SumAdjustmentsByAccount(ctx, acct.ID)
	if err != nil {
		return 0, err
	}
	return finance.AccountBalanceCents(acct.StartingBalanceCents, paid, adj), nil
}

func (s *Server) listAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := s.q.ListAccounts(r.Context())
	if err != nil {
		handleDBError(w, err)
		return
	}

	out := make([]accountResponse, 0, len(accounts))
	for _, a := range accounts {
		bal, err := s.accountBalance(r.Context(), a)
		if err != nil {
			handleDBError(w, err)
			return
		}
		out = append(out, accountResponse{
			ID:                   a.ID,
			Name:                 a.Name,
			Currency:             a.Currency,
			StartingBalanceCents: a.StartingBalanceCents,
			CurrentBalanceCents:  bal,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

type createAccountRequest struct {
	Name                 string `json:"name"`
	StartingBalanceCents int64  `json:"starting_balance_cents"`
	Currency             string `json:"currency"`
}

func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Currency == "" {
		req.Currency = "EUR"
	}

	acct, err := s.q.CreateAccount(r.Context(), store.CreateAccountParams{
		Name:                 req.Name,
		StartingBalanceCents: req.StartingBalanceCents,
		Currency:             req.Currency,
	})
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, accountResponse{
		ID:                   acct.ID,
		Name:                 acct.Name,
		Currency:             acct.Currency,
		StartingBalanceCents: acct.StartingBalanceCents,
		CurrentBalanceCents:  acct.StartingBalanceCents,
	})
}

type adjustmentResponse struct {
	ID          int64   `json:"id"`
	AmountCents int64   `json:"amount_cents"`
	Note        *string `json:"note"`
	OccurredOn  string  `json:"occurred_on"`
	CreatedAt   string  `json:"created_at"`
}

func toAdjustmentResponse(a store.BalanceAdjustment) adjustmentResponse {
	var note *string
	if a.Note.Valid {
		note = &a.Note.String
	}
	return adjustmentResponse{
		ID:          a.ID,
		AmountCents: a.AmountCents,
		Note:        note,
		OccurredOn:  formatDate(a.OccurredOn),
		CreatedAt:   a.CreatedAt.Time.Format(time.RFC3339),
	}
}

type createAdjustmentRequest struct {
	AmountCents int64   `json:"amount_cents"`
	Note        *string `json:"note"`
	OccurredOn  string  `json:"occurred_on"`
}

func (s *Server) createAdjustment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	var req createAdjustmentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	occurredOn, err := parseDate(req.OccurredOn)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid occurred_on")
		return
	}

	var note pgtype.Text
	if req.Note != nil {
		note = pgtype.Text{String: *req.Note, Valid: true}
	}

	adj, err := s.q.CreateAdjustment(r.Context(), store.CreateAdjustmentParams{
		AccountID:   id,
		AmountCents: req.AmountCents,
		Note:        note,
		OccurredOn:  occurredOn,
	})
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toAdjustmentResponse(adj))
}

func (s *Server) listAdjustments(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	adjs, err := s.q.ListAdjustmentsByAccount(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}

	out := make([]adjustmentResponse, 0, len(adjs))
	for _, a := range adjs {
		out = append(out, toAdjustmentResponse(a))
	}
	writeJSON(w, http.StatusOK, out)
}
