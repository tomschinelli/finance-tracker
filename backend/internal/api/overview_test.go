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
			AlreadySet          bool  `json:"already_set"`
		} `json:"budget"`
	}
	decodeBody(t, resp, &overview)

	if len(overview.Payments) != 1 {
		t.Fatalf("payments = %v, want one generated occurrence for the monthly rent definition", overview.Payments)
	}
	if overview.Budget.PredictedCostsCents != 80000 {
		t.Fatalf("predicted_costs_cents = %d, want 80000", overview.Budget.PredictedCostsCents)
	}
	if overview.Budget.AlreadySet {
		t.Fatalf("already_set should be false before any budget set-aside adjustment exists")
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
