package api_test

import (
	"net/http"
	"testing"
)

func TestOverviewMonthly_GeneratesOccurrenceAndComputesBudget(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Rent", "expense")
	acct := createAccount(t, h, "Fix Costs", 0)

	resp := do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Rent",
		"category_id":     cat.ID,
		"account_id":      acct.ID,
		"amount_cents":    80000,
		"interval_months": 1,
		"anchor_date":     "2026-01-01",
	})
	requireStatus(t, resp, http.StatusCreated)

	resp = do(t, h, "GET", "/overview/monthly?year=2026&month=1", nil)
	requireStatus(t, resp, http.StatusOK)
	var overview struct {
		Payments []paymentDTO `json:"payments"`
		Budget   struct {
			PredictedCostsCents int64 `json:"predicted_costs_cents"`
			SetAsideCents       int64 `json:"set_aside_cents"`
			CostsAlreadySet     bool  `json:"costs_already_set"`
			BufferAlreadySet    bool  `json:"buffer_already_set"`
		} `json:"budget"`
	}
	decodeBody(t, resp, &overview)

	if len(overview.Payments) != 1 {
		t.Fatalf("payments = %v, want one generated occurrence for the monthly rent definition", overview.Payments)
	}
	if overview.Budget.PredictedCostsCents != 80000 {
		t.Fatalf("predicted_costs_cents = %d, want 80000", overview.Budget.PredictedCostsCents)
	}
	if overview.Budget.CostsAlreadySet {
		t.Fatalf("costs_already_set should be false before any budget set-aside adjustment exists")
	}
	if overview.Budget.BufferAlreadySet {
		t.Fatalf("buffer_already_set should be false before any budget set-aside adjustment exists")
	}

	// Calling the endpoint again must not duplicate the generated occurrence
	// (idempotent on (recurring_definition_id, due_date)).
	resp = do(t, h, "GET", "/overview/monthly?year=2026&month=1", nil)
	requireStatus(t, resp, http.StatusOK)
	var second struct {
		Payments []paymentDTO `json:"payments"`
	}
	decodeBody(t, resp, &second)
	if len(second.Payments) != 1 {
		t.Fatalf("payments on second call = %v, want still one (no duplicate occurrence)", second.Payments)
	}
}

func TestOverviewMonthly_OverdueNonMonthlyOccurrenceCarriesForwardUntilPaid(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Bills", "expense")
	acct := createAccount(t, h, "Fix Costs", 0)

	resp := do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Quarterly waste fee",
		"category_id":     cat.ID,
		"account_id":      acct.ID,
		"amount_cents":    30000,
		"interval_months": 3,
		"anchor_date":     "2026-01-01",
	})
	requireStatus(t, resp, http.StatusCreated)

	// Generate January's occurrence, leave it unpaid.
	resp = do(t, h, "GET", "/overview/monthly?year=2026&month=1", nil)
	requireStatus(t, resp, http.StatusOK)
	var jan struct {
		Payments []paymentDTO `json:"payments"`
	}
	decodeBody(t, resp, &jan)
	if len(jan.Payments) != 1 || jan.Payments[0].DueDate != "2026-01-01" {
		t.Fatalf("january payments = %v, want exactly the 2026-01-01 occurrence", jan.Payments)
	}
	janOccurrenceID := jan.Payments[0].ID

	// February: nothing is due this specific month (next isn't until April),
	// but the unpaid January occurrence must still surface as overdue.
	resp = do(t, h, "GET", "/overview/monthly?year=2026&month=2", nil)
	requireStatus(t, resp, http.StatusOK)
	var feb struct {
		Payments        []paymentDTO `json:"payments"`
		OverduePayments []paymentDTO `json:"overdue_payments"`
	}
	decodeBody(t, resp, &feb)
	if len(feb.Payments) != 0 {
		t.Fatalf("february payments = %v, want none due in February itself", feb.Payments)
	}
	if len(feb.OverduePayments) != 1 || feb.OverduePayments[0].ID != janOccurrenceID {
		t.Fatalf("february overdue_payments = %v, want the unpaid January occurrence", feb.OverduePayments)
	}

	// Pay it, then re-check February: it must drop out of overdue_payments.
	resp = do(t, h, "PATCH", "/payments/"+itoa(janOccurrenceID), map[string]any{"paid_date": "2026-02-10"})
	requireStatus(t, resp, http.StatusOK)

	resp = do(t, h, "GET", "/overview/monthly?year=2026&month=2", nil)
	requireStatus(t, resp, http.StatusOK)
	decodeBody(t, resp, &feb)
	if len(feb.OverduePayments) != 0 {
		t.Fatalf("overdue_payments after paying = %v, want none", feb.OverduePayments)
	}
}

func TestOverviewMonthly_UpcomingNonMonthlyOccurrencesHiddenFromCurrentMonth(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Bills", "expense")
	acct := createAccount(t, h, "Fix Costs", 0)

	resp := do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Rent",
		"category_id":     cat.ID,
		"account_id":      acct.ID,
		"amount_cents":    80000,
		"interval_months": 1,
		"anchor_date":     "2026-01-01",
	})
	requireStatus(t, resp, http.StatusCreated)
	resp = do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Car insurance",
		"category_id":     cat.ID,
		"account_id":      acct.ID,
		"amount_cents":    60000,
		"interval_months": 12,
		"anchor_date":     "2026-06-01",
	})
	requireStatus(t, resp, http.StatusCreated)

	resp = do(t, h, "GET", "/overview/monthly?year=2026&month=1", nil)
	requireStatus(t, resp, http.StatusOK)
	var jan struct {
		Payments         []paymentDTO `json:"payments"`
		UpcomingPayments []paymentDTO `json:"upcoming_payments"`
	}
	decodeBody(t, resp, &jan)

	for _, p := range jan.Payments {
		if p.Label == "Car insurance" {
			t.Fatalf("january payments unexpectedly includes the future Car insurance occurrence: %v", jan.Payments)
		}
	}
	if len(jan.UpcomingPayments) != 1 || jan.UpcomingPayments[0].Label != "Car insurance" || jan.UpcomingPayments[0].DueDate != "2026-06-01" {
		t.Fatalf("january upcoming_payments = %v, want just the 2026-06-01 Car insurance occurrence", jan.UpcomingPayments)
	}
}

func TestOverviewMonthly_ExcludedAdhocCostNotCountedAgainstBudget(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Rent", "expense")
	acct := createAccount(t, h, "Fix Costs", 0)

	resp := do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Rent",
		"category_id":     cat.ID,
		"account_id":      acct.ID,
		"amount_cents":    80000,
		"interval_months": 1,
		"anchor_date":     "2026-01-01",
	})
	requireStatus(t, resp, http.StatusCreated)

	// An ad hoc cost posted to the same account, flagged as excluded from
	// the budget calculation (e.g. a one-off purchase, not a fixed cost).
	resp = do(t, h, "POST", "/payments", map[string]any{
		"label":                "One-off repair",
		"category_id":          cat.ID,
		"account_id":           acct.ID,
		"amount_cents":         50000,
		"due_date":             "2026-01-10",
		"paid_date":            "2026-01-10",
		"excluded_from_budget": true,
	})
	requireStatus(t, resp, http.StatusCreated)
	var created paymentDTO
	decodeBody(t, resp, &created)
	if !created.ExcludedFromBudget {
		t.Fatalf("excluded_from_budget = false on created payment, want true")
	}

	resp = do(t, h, "GET", "/overview/monthly?year=2026&month=1", nil)
	requireStatus(t, resp, http.StatusOK)
	var overview struct {
		Budget struct {
			PaidAgainstFixedCostsCents int64 `json:"paid_against_fixed_costs_cents"`
		} `json:"budget"`
		RemainingExpectedCents int64 `json:"remaining_expected_cents"`
	}
	decodeBody(t, resp, &overview)

	// Only the (unpaid) 80000 rent occurrence should count; the excluded ad
	// hoc cost must not bleed into either figure.
	if overview.Budget.PaidAgainstFixedCostsCents != 0 {
		t.Fatalf("paid_against_fixed_costs_cents = %d, want 0 (excluded cost must not count)", overview.Budget.PaidAgainstFixedCostsCents)
	}
	if overview.RemainingExpectedCents != 80000 {
		t.Fatalf("remaining_expected_cents = %d, want 80000 (excluded cost must not count)", overview.RemainingExpectedCents)
	}
}

func TestOverviewMonthly_BudgetCostsAndBufferTrackedIndependently(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Bills", "expense")
	source := createAccount(t, h, "Savings", 100000) // becomes default
	dest := createAccount(t, h, "Fix Costs", 0)

	resp := do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Rent",
		"category_id":     cat.ID,
		"account_id":      dest.ID,
		"amount_cents":    80000,
		"interval_months": 1,
		"anchor_date":     "2026-01-01",
	})
	requireStatus(t, resp, http.StatusCreated)
	resp = do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Car insurance",
		"category_id":     cat.ID,
		"account_id":      dest.ID,
		"amount_cents":    12000,
		"interval_months": 12,
		"anchor_date":     "2026-06-01",
	})
	requireStatus(t, resp, http.StatusCreated)

	resp = do(t, h, "GET", "/overview/monthly?year=2026&month=1", nil)
	requireStatus(t, resp, http.StatusOK)
	var overview struct {
		Budget struct {
			PredictedCostsCents int64  `json:"predicted_costs_cents"`
			BufferCents         int64  `json:"buffer_cents"`
			CostsAlreadySet     bool   `json:"costs_already_set"`
			BufferAlreadySet    bool   `json:"buffer_already_set"`
			CostsMarkerNote     string `json:"costs_marker_note"`
			BufferMarkerNote    string `json:"buffer_marker_note"`
		} `json:"budget"`
	}
	decodeBody(t, resp, &overview)

	if overview.Budget.CostsMarkerNote == overview.Budget.BufferMarkerNote {
		t.Fatalf("costs and buffer marker notes must differ, both = %q", overview.Budget.CostsMarkerNote)
	}

	// Only set aside the current month's costs, not the buffer.
	resp = do(t, h, "POST", "/transfers", map[string]any{
		"from_account_id": source.ID,
		"to_account_id":   dest.ID,
		"amount_cents":    overview.Budget.PredictedCostsCents,
		"note":            overview.Budget.CostsMarkerNote,
		"occurred_on":     "2026-01-01",
	})
	requireStatus(t, resp, http.StatusCreated)

	resp = do(t, h, "GET", "/overview/monthly?year=2026&month=1", nil)
	requireStatus(t, resp, http.StatusOK)
	decodeBody(t, resp, &overview)

	if !overview.Budget.CostsAlreadySet {
		t.Fatalf("costs_already_set should be true after transferring the costs marker amount")
	}
	if overview.Budget.BufferAlreadySet {
		t.Fatalf("buffer_already_set should still be false: only the costs part was transferred")
	}
}

func TestOverviewYearly_SplitsPaidAndRemainingByAccount(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	// /overview/yearly only reports on interval_months=12 recurring costs
	// (e.g. annual insurance premiums), not ad hoc payments.
	cat := createCategory(t, h, "Insurance", "expense")
	acct := createAccount(t, h, "Fix Costs", 100000)

	resp := do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Car insurance",
		"category_id":     cat.ID,
		"account_id":      acct.ID,
		"amount_cents":    30000,
		"interval_months": 12,
		"anchor_date":     "2026-06-01",
	})
	requireStatus(t, resp, http.StatusCreated)

	resp = do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Home insurance",
		"category_id":     cat.ID,
		"account_id":      acct.ID,
		"amount_cents":    12000,
		"interval_months": 12,
		"anchor_date":     "2026-09-01",
	})
	requireStatus(t, resp, http.StatusCreated)

	// Generate both years' occurrences, then mark the car insurance one paid.
	resp = do(t, h, "GET", "/overview/yearly?year=2026", nil)
	requireStatus(t, resp, http.StatusOK)
	var initial struct {
		Payments []paymentDTO `json:"payments"`
	}
	decodeBody(t, resp, &initial)
	for _, p := range initial.Payments {
		if p.Label == "Car insurance" {
			resp = do(t, h, "PATCH", "/payments/"+itoa(p.ID), map[string]any{"paid_date": "2026-06-01"})
			requireStatus(t, resp, http.StatusOK)
		}
	}

	resp = do(t, h, "GET", "/overview/yearly?year=2026", nil)
	requireStatus(t, resp, http.StatusOK)
	var overview struct {
		Totals struct {
			PaidCents      int64 `json:"paid_cents"`
			RemainingCents int64 `json:"remaining_cents"`
		} `json:"totals"`
		Accounts []struct {
			AccountID                int64 `json:"account_id"`
			RemainingYearlyCostCents int64 `json:"remaining_yearly_cost_cents"`
			CoversRemaining          bool  `json:"covers_remaining"`
		} `json:"accounts"`
	}
	decodeBody(t, resp, &overview)

	if overview.Totals.PaidCents != 30000 {
		t.Fatalf("paid_cents = %d, want 30000", overview.Totals.PaidCents)
	}
	if overview.Totals.RemainingCents != 12000 {
		t.Fatalf("remaining_cents = %d, want 12000", overview.Totals.RemainingCents)
	}
	if len(overview.Accounts) != 1 {
		t.Fatalf("accounts = %v, want exactly one", overview.Accounts)
	}
	if overview.Accounts[0].RemainingYearlyCostCents != 12000 {
		t.Fatalf("remaining_yearly_cost_cents = %d, want 12000", overview.Accounts[0].RemainingYearlyCostCents)
	}
	if !overview.Accounts[0].CoversRemaining {
		t.Fatalf("covers_remaining should be true: balance 100000 >= remaining 12000")
	}
}
