package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"finance-backend/internal/finance"
	"finance-backend/internal/store"
)

// generateOccurrences inserts any missing payments rows for active recurring
// definitions whose due date falls within [from, to]. Idempotent: relies on
// the (recurring_definition_id, due_date) unique constraint + ON CONFLICT.
func (s *Server) generateOccurrences(ctx context.Context, from, to time.Time) error {
	defs, err := s.q.ListActiveRecurringDefinitions(ctx)
	if err != nil {
		return err
	}
	for _, d := range defs {
		for _, dt := range finance.ComputeOccurrenceDates(d.AnchorDate.Time, int(d.IntervalMonths), from, to) {
			_, err := s.q.InsertOccurrence(ctx, store.InsertOccurrenceParams{
				RecurringDefinitionID: pgtype.Int8{Int64: d.ID, Valid: true},
				Label:                 d.Label,
				CategoryID:            d.CategoryID,
				AccountID:             d.AccountID,
				AmountCents:           d.AmountCents,
				DueDate:               pgtype.Date{Time: dt, Valid: true},
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type yearlyAccountOverview struct {
	AccountID                int64  `json:"account_id"`
	Name                     string `json:"name"`
	CurrentBalanceCents      int64  `json:"current_balance_cents"`
	RemainingYearlyCostCents int64  `json:"remaining_yearly_cost_cents"`
	CoversRemaining          bool   `json:"covers_remaining"`
}

func (s *Server) overviewYearly(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	year, err := strconv.Atoi(yearStr)
	if yearStr == "" || err != nil {
		writeError(w, http.StatusBadRequest, "year is required")
		return
	}

	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
	if err := s.generateOccurrences(r.Context(), from, to); err != nil {
		handleDBError(w, err)
		return
	}

	payments, err := s.q.ListYearlyFrequencyPayments(r.Context(), int32(year))
	if err != nil {
		handleDBError(w, err)
		return
	}

	var paidCents, remainingCents int64
	remainingByAccount := map[int64]int64{}
	paymentResponses := make([]paymentResponse, 0, len(payments))
	for _, p := range payments {
		paymentResponses = append(paymentResponses, toPaymentResponse(p))
		if p.PaidDate.Valid {
			paidCents += p.AmountCents
		} else {
			remainingCents += p.AmountCents
			remainingByAccount[p.AccountID] += p.AmountCents
		}
	}

	accounts, err := s.q.ListAccounts(r.Context())
	if err != nil {
		handleDBError(w, err)
		return
	}
	accountsOut := make([]yearlyAccountOverview, 0, len(accounts))
	for _, a := range accounts {
		bal, err := s.accountBalance(r.Context(), a)
		if err != nil {
			handleDBError(w, err)
			return
		}
		remaining := remainingByAccount[a.ID]
		accountsOut = append(accountsOut, yearlyAccountOverview{
			AccountID:                a.ID,
			Name:                     a.Name,
			CurrentBalanceCents:      bal,
			RemainingYearlyCostCents: remaining,
			CoversRemaining:          bal >= remaining,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"year":     year,
		"payments": paymentResponses,
		"totals": map[string]int64{
			"paid_cents":           paidCents,
			"remaining_cents":      remainingCents,
			"monthly_buffer_cents": (paidCents + remainingCents) / 12,
		},
		"accounts": accountsOut,
	})
}

type monthlyAccountOverview struct {
	AccountID           int64  `json:"account_id"`
	Name                string `json:"name"`
	CurrentBalanceCents int64  `json:"current_balance_cents"`
}

type monthlyBudgetOverview struct {
	PredictedCostsCents        int64  `json:"predicted_costs_cents"`
	BufferCents                int64  `json:"buffer_cents"`
	SetAsideCents              int64  `json:"set_aside_cents"`
	PaidAgainstFixedCostsCents int64  `json:"paid_against_fixed_costs_cents"`
	RemainingCents             int64  `json:"remaining_cents"`
	AlreadySet                 bool   `json:"already_set"`
	MarkerNote                 string `json:"marker_note"`
}

// budgetMarkerNote is the balance_adjustments.note convention used to mark
// that this month's set-aside budget has already been deposited into an
// account, so the same month isn't double-counted. No separate table is
// needed: the note text itself is the idempotency marker.
func budgetMarkerNote(year, month int) string {
	return fmt.Sprintf("Monthly budget set-aside %04d-%02d", year, month)
}

func (s *Server) overviewMonthly(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	yearStr, monthStr := q.Get("year"), q.Get("month")
	year, errYear := strconv.Atoi(yearStr)
	month, errMonth := strconv.Atoi(monthStr)
	if yearStr == "" || monthStr == "" || errYear != nil || errMonth != nil || month < 1 || month > 12 {
		writeError(w, http.StatusBadRequest, "year and month are required")
		return
	}

	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, -1)
	to := monthEnd
	if today := time.Now().UTC(); today.After(to) {
		to = today
	}
	// from is intentionally far in the past: generation naturally starts at
	// each definition's own anchor date, so this just avoids excluding
	// overdue occurrences from earlier months.
	from := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := s.generateOccurrences(r.Context(), from, to); err != nil {
		handleDBError(w, err)
		return
	}

	payments, err := s.q.ListPaymentsByYearMonth(r.Context(), store.ListPaymentsByYearMonthParams{
		Column1: int32(year),
		Column2: int32(month),
	})
	if err != nil {
		handleDBError(w, err)
		return
	}

	categories, err := s.q.ListCategories(r.Context())
	if err != nil {
		handleDBError(w, err)
		return
	}
	kindByCategory := make(map[int64]string, len(categories))
	for _, c := range categories {
		kindByCategory[c.ID] = c.Kind
	}

	var paidAgainstFixed, remainingExpected int64
	paymentResponses := make([]paymentResponse, 0, len(payments))
	for _, p := range payments {
		paymentResponses = append(paymentResponses, toPaymentResponse(p))
		if kindByCategory[p.CategoryID] != "expense" {
			continue
		}
		if p.PaidDate.Valid {
			paidAgainstFixed += p.AmountCents
		} else {
			remainingExpected += p.AmountCents
		}
	}

	sums, err := s.q.SumActiveDefinitionsByIntervalKind(r.Context())
	if err != nil {
		handleDBError(w, err)
		return
	}
	var monthlyExpense int64
	bufferEntries := map[int]int64{}
	for _, row := range sums {
		if row.Kind != "expense" {
			continue
		}
		if row.IntervalMonths == 1 {
			monthlyExpense += row.Total
			continue
		}
		bufferEntries[int(row.IntervalMonths)] += row.Total
	}
	// Predicted costs is what's actually due this month (fixed monthly
	// expenses only); non-monthly costs are smoothed into the buffer
	// instead, since the account is topped up for them every month
	// regardless of which month they actually fall due in.
	predictedCosts := monthlyExpense
	buffer := finance.MonthlyBufferCents(bufferEntries)
	setAside := predictedCosts + buffer

	budgetNote := budgetMarkerNote(year, month)
	budgetAdjustments, err := s.q.FindAdjustmentsByNote(r.Context(), pgtype.Text{String: budgetNote, Valid: true})
	if err != nil {
		handleDBError(w, err)
		return
	}

	accounts, err := s.q.ListAccounts(r.Context())
	if err != nil {
		handleDBError(w, err)
		return
	}
	accountsOut := make([]monthlyAccountOverview, 0, len(accounts))
	for _, a := range accounts {
		bal, err := s.accountBalance(r.Context(), a)
		if err != nil {
			handleDBError(w, err)
			return
		}
		accountsOut = append(accountsOut, monthlyAccountOverview{AccountID: a.ID, Name: a.Name, CurrentBalanceCents: bal})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"year":     year,
		"month":    month,
		"payments": paymentResponses,
		"budget": monthlyBudgetOverview{
			PredictedCostsCents:        predictedCosts,
			BufferCents:                buffer,
			SetAsideCents:              setAside,
			PaidAgainstFixedCostsCents: paidAgainstFixed,
			RemainingCents:             setAside - paidAgainstFixed,
			AlreadySet:                 len(budgetAdjustments) > 0,
			MarkerNote:                 budgetNote,
		},
		"accounts":                 accountsOut,
		"remaining_expected_cents": remainingExpected,
	})
}
