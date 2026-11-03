package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
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
	IsDefault            bool   `json:"is_default"`
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
			IsDefault:            a.IsDefault,
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

	ctx := r.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		handleDBError(w, err)
		return
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)

	acct, err := q.CreateAccount(ctx, store.CreateAccountParams{
		Name:                 req.Name,
		StartingBalanceCents: req.StartingBalanceCents,
		Currency:             req.Currency,
	})
	if err != nil {
		handleDBError(w, err)
		return
	}

	// The first account ever created becomes the default automatically,
	// since the app needs exactly one default account to fund transfers
	// from. Later accounts stay non-default until explicitly set.
	if _, err := q.GetDefaultAccount(ctx); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			handleDBError(w, err)
			return
		}
		acct, err = q.MarkAccountDefault(ctx, acct.ID)
		if err != nil {
			handleDBError(w, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, accountResponse{
		ID:                   acct.ID,
		Name:                 acct.Name,
		Currency:             acct.Currency,
		StartingBalanceCents: acct.StartingBalanceCents,
		CurrentBalanceCents:  acct.StartingBalanceCents,
		IsDefault:            acct.IsDefault,
	})
}

// setDefaultAccount makes the given account the sole default account,
// unsetting any previous one. The default account is the only one that can
// be funded from outside the ledger directly; every other account is
// funded via a transfer out of it (see POST /transfers).
func (s *Server) setDefaultAccount(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	ctx := r.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		handleDBError(w, err)
		return
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)

	if err := q.UnsetDefaultAccounts(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	acct, err := q.MarkAccountDefault(ctx, id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		handleDBError(w, err)
		return
	}

	bal, err := s.accountBalance(ctx, acct)
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, accountResponse{
		ID:                   acct.ID,
		Name:                 acct.Name,
		Currency:             acct.Currency,
		StartingBalanceCents: acct.StartingBalanceCents,
		CurrentBalanceCents:  bal,
		IsDefault:            acct.IsDefault,
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

	if _, err := s.q.GetAccount(r.Context(), id); err != nil {
		handleDBError(w, err)
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

func (s *Server) listAccountPayments(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	payments, err := s.q.ListPaymentsByAccount(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}

	out := make([]paymentResponse, 0, len(payments))
	for _, p := range payments {
		out = append(out, toPaymentResponse(p))
	}
	writeJSON(w, http.StatusOK, out)
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
