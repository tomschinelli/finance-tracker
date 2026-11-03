package api

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"finance-backend/internal/store"
)

type createTransferRequest struct {
	FromAccountID int64   `json:"from_account_id"`
	ToAccountID   int64   `json:"to_account_id"`
	AmountCents   int64   `json:"amount_cents"`
	Note          *string `json:"note"`
	OccurredOn    string  `json:"occurred_on"`
}

type transferResponse struct {
	FromAccountID int64              `json:"from_account_id"`
	ToAccountID   int64              `json:"to_account_id"`
	AmountCents   int64              `json:"amount_cents"`
	Debit         adjustmentResponse `json:"debit"`
	Credit        adjustmentResponse `json:"credit"`
}

// createTransfer moves money from one account to another: a pair of
// balance_adjustments rows (a negative on the source, a positive on the
// destination), inserted together in one DB transaction so they can't
// diverge. This is how any non-default account is funded — the caller
// picks both accounts explicitly; only the default account can be funded
// without a transfer (see createAdjustment).
func (s *Server) createTransfer(w http.ResponseWriter, r *http.Request) {
	var req createTransferRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.FromAccountID == 0 || req.ToAccountID == 0 || req.AmountCents <= 0 {
		writeError(w, http.StatusBadRequest, "from_account_id, to_account_id, and a positive amount_cents are required")
		return
	}
	if req.FromAccountID == req.ToAccountID {
		writeError(w, http.StatusBadRequest, "from and to accounts must differ")
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

	ctx := r.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		handleDBError(w, err)
		return
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)

	debit, err := q.CreateAdjustment(ctx, store.CreateAdjustmentParams{
		AccountID:   req.FromAccountID,
		AmountCents: -req.AmountCents,
		Note:        note,
		OccurredOn:  occurredOn,
	})
	if err != nil {
		handleDBError(w, err)
		return
	}
	credit, err := q.CreateAdjustment(ctx, store.CreateAdjustmentParams{
		AccountID:   req.ToAccountID,
		AmountCents: req.AmountCents,
		Note:        note,
		OccurredOn:  occurredOn,
	})
	if err != nil {
		handleDBError(w, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, transferResponse{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		AmountCents:   req.AmountCents,
		Debit:         toAdjustmentResponse(debit),
		Credit:        toAdjustmentResponse(credit),
	})
}
