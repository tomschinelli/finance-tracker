package api_test

import (
	"net/http"
	"testing"
)

func TestCreateRecurring_Success(t *testing.T) {
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
	var def recurringDTO
	decodeBody(t, resp, &def)
	if !def.Active {
		t.Fatalf("recurring definition should default to active")
	}
}

func TestCreateRecurring_InvalidIntervalRejected(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Rent", "expense")
	acct := createAccount(t, h, "Fix Costs", 0)

	resp := do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Rent",
		"category_id":     cat.ID,
		"account_id":      acct.ID,
		"amount_cents":    80000,
		"interval_months": 0,
		"anchor_date":     "2026-01-01",
	})
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestCreateRecurring_CustomInterval(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Rent", "expense")
	acct := createAccount(t, h, "Fix Costs", 0)

	resp := do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Pest control",
		"category_id":     cat.ID,
		"account_id":      acct.ID,
		"amount_cents":    6000,
		"interval_months": 2,
		"anchor_date":     "2026-01-01",
	})
	requireStatus(t, resp, http.StatusCreated)
	var def recurringDTO
	decodeBody(t, resp, &def)
	if def.IntervalMonths != 2 {
		t.Fatalf("interval_months = %d, want 2", def.IntervalMonths)
	}
}

func TestCreateRecurring_WithTags(t *testing.T) {
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
		"tags":            []string{"housing", "essential"},
	})
	requireStatus(t, resp, http.StatusCreated)
	var def recurringDTO
	decodeBody(t, resp, &def)
	if len(def.Tags) != 2 || def.Tags[0] != "housing" || def.Tags[1] != "essential" {
		t.Fatalf("tags = %v, want [housing essential]", def.Tags)
	}
}

func TestListRecurring_FilterByTag(t *testing.T) {
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
		"tags":            []string{"housing"},
	})
	requireStatus(t, resp, http.StatusCreated)

	resp = do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Netflix",
		"category_id":     cat.ID,
		"account_id":      acct.ID,
		"amount_cents":    1500,
		"interval_months": 1,
		"anchor_date":     "2026-01-01",
		"tags":            []string{"subscriptions"},
	})
	requireStatus(t, resp, http.StatusCreated)

	var filtered []recurringDTO
	decodeBody(t, do(t, h, "GET", "/recurring?tag=housing", nil), &filtered)
	if len(filtered) != 1 || filtered[0].Label != "Rent" {
		t.Fatalf("recurring filtered by tag=housing = %v, want only Rent", filtered)
	}

	var all []recurringDTO
	decodeBody(t, do(t, h, "GET", "/recurring", nil), &all)
	if len(all) != 2 {
		t.Fatalf("recurring with no filter = %v, want both definitions", all)
	}
}

func TestUpdateRecurring_PartialUpdateAppliesOnlyGivenFields(t *testing.T) {
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
	var created recurringDTO
	decodeBody(t, resp, &created)

	resp = do(t, h, "PATCH", "/recurring/"+itoa(created.ID), map[string]any{"active": false})
	requireStatus(t, resp, http.StatusOK)
	var updated recurringDTO
	decodeBody(t, resp, &updated)
	if updated.Active {
		t.Fatalf("active should be false after patch")
	}
	if updated.AmountCents != 80000 {
		t.Fatalf("amount_cents = %d, want unchanged 80000", updated.AmountCents)
	}
}
