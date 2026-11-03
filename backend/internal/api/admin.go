package api

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"finance-backend/internal/store"
)

const exportVersion = 1

type exportPayload struct {
	Version              int                         `json:"version"`
	ExportedAt           string                      `json:"exported_at"`
	Accounts             []exportAccount             `json:"accounts"`
	Categories           []exportCategory            `json:"categories"`
	RecurringDefinitions []exportRecurringDefinition `json:"recurring_definitions"`
	Payments             []exportPayment             `json:"payments"`
	BalanceAdjustments   []exportBalanceAdjustment   `json:"balance_adjustments"`
}

type exportAccount struct {
	ID                   int64  `json:"id"`
	Name                 string `json:"name"`
	StartingBalanceCents int64  `json:"starting_balance_cents"`
	Currency             string `json:"currency"`
	IsDefault            bool   `json:"is_default"`
	CreatedAt            string `json:"created_at"`
}

type exportCategory struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	IsDefault bool   `json:"is_default"`
}

type exportRecurringDefinition struct {
	ID             int64    `json:"id"`
	Label          string   `json:"label"`
	CategoryID     int64    `json:"category_id"`
	AccountID      int64    `json:"account_id"`
	AmountCents    int64    `json:"amount_cents"`
	IntervalMonths int32    `json:"interval_months"`
	AnchorDate     string   `json:"anchor_date"`
	Active         bool     `json:"active"`
	Tags           []string `json:"tags"`
	CreatedAt      string   `json:"created_at"`
}

type exportPayment struct {
	ID                    int64   `json:"id"`
	RecurringDefinitionID *int64  `json:"recurring_definition_id"`
	Label                 string  `json:"label"`
	CategoryID            int64   `json:"category_id"`
	AccountID             int64   `json:"account_id"`
	AmountCents           int64   `json:"amount_cents"`
	DueDate               string  `json:"due_date"`
	PaidDate              *string `json:"paid_date"`
	CreatedAt             string  `json:"created_at"`
}

type exportBalanceAdjustment struct {
	ID          int64   `json:"id"`
	AccountID   int64   `json:"account_id"`
	AmountCents int64   `json:"amount_cents"`
	Note        *string `json:"note"`
	OccurredOn  string  `json:"occurred_on"`
	CreatedAt   string  `json:"created_at"`
}

// exportData returns every row of every table (including ids and
// timestamps) as a full snapshot, so it can be re-imported later via
// importData to restore the exact same state.
func (s *Server) exportData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accounts, err := s.q.ListAccounts(ctx)
	if err != nil {
		handleDBError(w, err)
		return
	}
	categories, err := s.q.ListCategories(ctx)
	if err != nil {
		handleDBError(w, err)
		return
	}
	recs, err := s.q.ListRecurringDefinitions(ctx)
	if err != nil {
		handleDBError(w, err)
		return
	}
	payments, err := s.q.ListAllPayments(ctx)
	if err != nil {
		handleDBError(w, err)
		return
	}
	adjustments, err := s.q.ListAllBalanceAdjustments(ctx)
	if err != nil {
		handleDBError(w, err)
		return
	}

	out := exportPayload{
		Version:    exportVersion,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
	}
	for _, a := range accounts {
		out.Accounts = append(out.Accounts, exportAccount{
			ID:                   a.ID,
			Name:                 a.Name,
			StartingBalanceCents: a.StartingBalanceCents,
			Currency:             a.Currency,
			IsDefault:            a.IsDefault,
			CreatedAt:            a.CreatedAt.Time.Format(time.RFC3339),
		})
	}
	for _, c := range categories {
		out.Categories = append(out.Categories, exportCategory{ID: c.ID, Name: c.Name, Kind: c.Kind, IsDefault: c.IsDefault})
	}
	for _, rd := range recs {
		out.RecurringDefinitions = append(out.RecurringDefinitions, exportRecurringDefinition{
			ID:             rd.ID,
			Label:          rd.Label,
			CategoryID:     rd.CategoryID,
			AccountID:      rd.AccountID,
			AmountCents:    rd.AmountCents,
			IntervalMonths: rd.IntervalMonths,
			AnchorDate:     formatDate(rd.AnchorDate),
			Active:         rd.Active,
			Tags:           rd.Tags,
			CreatedAt:      rd.CreatedAt.Time.Format(time.RFC3339),
		})
	}
	for _, p := range payments {
		var rdID *int64
		if p.RecurringDefinitionID.Valid {
			v := p.RecurringDefinitionID.Int64
			rdID = &v
		}
		out.Payments = append(out.Payments, exportPayment{
			ID:                    p.ID,
			RecurringDefinitionID: rdID,
			Label:                 p.Label,
			CategoryID:            p.CategoryID,
			AccountID:             p.AccountID,
			AmountCents:           p.AmountCents,
			DueDate:               formatDate(p.DueDate),
			PaidDate:              formatDatePtr(p.PaidDate),
			CreatedAt:             p.CreatedAt.Time.Format(time.RFC3339),
		})
	}
	for _, adj := range adjustments {
		var note *string
		if adj.Note.Valid {
			note = &adj.Note.String
		}
		out.BalanceAdjustments = append(out.BalanceAdjustments, exportBalanceAdjustment{
			ID:          adj.ID,
			AccountID:   adj.AccountID,
			AmountCents: adj.AmountCents,
			Note:        note,
			OccurredOn:  formatDate(adj.OccurredOn),
			CreatedAt:   adj.CreatedAt.Time.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, out)
}

// importData unconditionally replaces ALL data (accounts and categories
// included) with the given snapshot, preserving ids so cross-table
// references stay valid, then resets each table's sequence past the
// highest imported id.
func (s *Server) importData(w http.ResponseWriter, r *http.Request) {
	var req exportPayload
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Version != exportVersion {
		writeError(w, http.StatusBadRequest, "unsupported export version")
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

	// Child -> parent order, mirroring the FK dependency graph.
	if err := q.ClearPayments(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ClearBalanceAdjustments(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ClearRecurringDefinitions(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ClearAccounts(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ClearCategories(ctx); err != nil {
		handleDBError(w, err)
		return
	}

	// Parent -> child order for inserts, preserving the original ids.
	for _, c := range req.Categories {
		if err := q.InsertCategoryWithID(ctx, store.InsertCategoryWithIDParams{ID: c.ID, Name: c.Name, Kind: c.Kind, IsDefault: c.IsDefault}); err != nil {
			handleDBError(w, err)
			return
		}
	}
	for _, a := range req.Accounts {
		createdAt, err := time.Parse(time.RFC3339, a.CreatedAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid account created_at")
			return
		}
		if err := q.InsertAccountWithID(ctx, store.InsertAccountWithIDParams{
			ID:                   a.ID,
			Name:                 a.Name,
			StartingBalanceCents: a.StartingBalanceCents,
			Currency:             a.Currency,
			IsDefault:            a.IsDefault,
			CreatedAt:            pgtype.Timestamptz{Time: createdAt, Valid: true},
		}); err != nil {
			handleDBError(w, err)
			return
		}
	}
	for _, rd := range req.RecurringDefinitions {
		anchorDate, err := parseDate(rd.AnchorDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid recurring definition anchor_date")
			return
		}
		createdAt, err := time.Parse(time.RFC3339, rd.CreatedAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid recurring definition created_at")
			return
		}
		tags := rd.Tags
		if tags == nil {
			tags = []string{}
		}
		if err := q.InsertRecurringDefinitionWithID(ctx, store.InsertRecurringDefinitionWithIDParams{
			ID:             rd.ID,
			Label:          rd.Label,
			CategoryID:     rd.CategoryID,
			AccountID:      rd.AccountID,
			AmountCents:    rd.AmountCents,
			IntervalMonths: rd.IntervalMonths,
			AnchorDate:     anchorDate,
			Active:         rd.Active,
			CreatedAt:      pgtype.Timestamptz{Time: createdAt, Valid: true},
			Tags:           tags,
		}); err != nil {
			handleDBError(w, err)
			return
		}
	}
	for _, p := range req.Payments {
		dueDate, err := parseDate(p.DueDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid payment due_date")
			return
		}
		var paidDate pgtype.Date
		if p.PaidDate != nil {
			paidDate, err = parseDate(*p.PaidDate)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid payment paid_date")
				return
			}
		}
		createdAt, err := time.Parse(time.RFC3339, p.CreatedAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid payment created_at")
			return
		}
		var rdID pgtype.Int8
		if p.RecurringDefinitionID != nil {
			rdID = pgtype.Int8{Int64: *p.RecurringDefinitionID, Valid: true}
		}
		if err := q.InsertPaymentWithID(ctx, store.InsertPaymentWithIDParams{
			ID:                    p.ID,
			RecurringDefinitionID: rdID,
			Label:                 p.Label,
			CategoryID:            p.CategoryID,
			AccountID:             p.AccountID,
			AmountCents:           p.AmountCents,
			DueDate:               dueDate,
			PaidDate:              paidDate,
			CreatedAt:             pgtype.Timestamptz{Time: createdAt, Valid: true},
		}); err != nil {
			handleDBError(w, err)
			return
		}
	}
	for _, adj := range req.BalanceAdjustments {
		occurredOn, err := parseDate(adj.OccurredOn)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid adjustment occurred_on")
			return
		}
		createdAt, err := time.Parse(time.RFC3339, adj.CreatedAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid adjustment created_at")
			return
		}
		var note pgtype.Text
		if adj.Note != nil {
			note = pgtype.Text{String: *adj.Note, Valid: true}
		}
		if err := q.InsertBalanceAdjustmentWithID(ctx, store.InsertBalanceAdjustmentWithIDParams{
			ID:          adj.ID,
			AccountID:   adj.AccountID,
			AmountCents: adj.AmountCents,
			Note:        note,
			OccurredOn:  occurredOn,
			CreatedAt:   pgtype.Timestamptz{Time: createdAt, Valid: true},
		}); err != nil {
			handleDBError(w, err)
			return
		}
	}

	if err := q.ResetCategoriesSeq(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ResetAccountsSeq(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ResetRecurringDefinitionsSeq(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ResetPaymentsSeq(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ResetBalanceAdjustmentsSeq(ctx); err != nil {
		handleDBError(w, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// clearData wipes every table (payments, recurring definitions, balance
// adjustments, accounts, categories), leaving the database completely
// empty, then resets each sequence so ids start from 1 again.
func (s *Server) clearData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		handleDBError(w, err)
		return
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)

	// Child -> parent order, mirroring the FK dependency graph.
	if err := q.ClearPayments(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ClearRecurringDefinitions(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ClearBalanceAdjustments(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ClearAccounts(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ClearCategories(ctx); err != nil {
		handleDBError(w, err)
		return
	}

	if err := q.ResetPaymentsSeq(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ResetRecurringDefinitionsSeq(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ResetBalanceAdjustmentsSeq(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ResetAccountsSeq(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	if err := q.ResetCategoriesSeq(ctx); err != nil {
		handleDBError(w, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		handleDBError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
