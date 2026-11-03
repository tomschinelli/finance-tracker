package api_test

import (
	"net/http"
	"testing"
)

// seedFullDataset creates one row in every table that /data/clear and
// /data/export touch, so tests can assert the whole graph is affected.
func seedFullDataset(t *testing.T, h http.Handler) {
	t.Helper()
	cat := createCategory(t, h, "Rent", "expense")
	acct := createAccount(t, h, "Main", 10000)

	resp := do(t, h, "POST", "/recurring", map[string]any{
		"label":           "Rent",
		"category_id":     cat.ID,
		"account_id":      acct.ID,
		"amount_cents":    50000,
		"interval_months": 1,
		"anchor_date":     "2026-01-01",
	})
	requireStatus(t, resp, http.StatusCreated)

	resp = do(t, h, "POST", "/payments", map[string]any{
		"label":        "Ad hoc expense",
		"category_id":  cat.ID,
		"account_id":   acct.ID,
		"amount_cents": 1234,
		"due_date":     "2026-01-15",
	})
	requireStatus(t, resp, http.StatusCreated)

	resp = do(t, h, "POST", "/accounts/"+itoa(acct.ID)+"/adjustments", map[string]any{
		"amount_cents": 500,
		"occurred_on":  "2026-01-10",
	})
	requireStatus(t, resp, http.StatusCreated)
}

func TestClearData_WipesEverythingIncludingAccounts(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	seedFullDataset(t, h)

	resp := do(t, h, "POST", "/data/clear", nil)
	requireStatus(t, resp, http.StatusNoContent)

	var accounts []accountDTO
	decodeBody(t, do(t, h, "GET", "/accounts", nil), &accounts)
	if len(accounts) != 0 {
		t.Fatalf("accounts after clear = %v, want empty", accounts)
	}

	var categories []categoryDTO
	decodeBody(t, do(t, h, "GET", "/categories", nil), &categories)
	if len(categories) != 1 || categories[0].Name != "Others" {
		t.Fatalf("categories after clear = %v, want only the default Others category", categories)
	}

	var recurring []recurringDTO
	decodeBody(t, do(t, h, "GET", "/recurring", nil), &recurring)
	if len(recurring) != 0 {
		t.Fatalf("recurring definitions after clear = %v, want empty", recurring)
	}

	var payments []paymentDTO
	decodeBody(t, do(t, h, "GET", "/payments?year=2026", nil), &payments)
	if len(payments) != 0 {
		t.Fatalf("payments after clear = %v, want empty", payments)
	}
}

func TestClearData_ResetsSequencesToOne(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	seedFullDataset(t, h)

	resp := do(t, h, "POST", "/data/clear", nil)
	requireStatus(t, resp, http.StatusNoContent)

	acct := createAccount(t, h, "Fresh start", 0)
	if acct.ID != 1 {
		t.Fatalf("account id after clear = %d, want 1 (sequence not reset)", acct.ID)
	}
	if !acct.IsDefault {
		t.Fatalf("first account created after a full clear should become the default again")
	}
}

func TestExportImport_RoundTrip(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	seedFullDataset(t, h)

	exportResp := do(t, h, "GET", "/data/export", nil)
	requireStatus(t, exportResp, http.StatusOK)
	var snapshot map[string]any
	decodeBody(t, exportResp, &snapshot)

	clearResp := do(t, h, "POST", "/data/clear", nil)
	requireStatus(t, clearResp, http.StatusNoContent)

	var accountsAfterClear []accountDTO
	decodeBody(t, do(t, h, "GET", "/accounts", nil), &accountsAfterClear)
	if len(accountsAfterClear) != 0 {
		t.Fatalf("accounts after clear = %v, want empty", accountsAfterClear)
	}

	importResp := do(t, h, "POST", "/data/import", snapshot)
	requireStatus(t, importResp, http.StatusNoContent)

	var accounts []accountDTO
	decodeBody(t, do(t, h, "GET", "/accounts", nil), &accounts)
	if len(accounts) != 1 || accounts[0].Name != "Main" {
		t.Fatalf("accounts after import = %v, want one account named Main", accounts)
	}

	var categories []categoryDTO
	decodeBody(t, do(t, h, "GET", "/categories", nil), &categories)
	if len(categories) != 2 || categories[0].Name != "Rent" || categories[1].Name != "Others" {
		t.Fatalf("categories after import = %v, want Rent plus a freshly re-created default Others category", categories)
	}

	var payments []paymentDTO
	decodeBody(t, do(t, h, "GET", "/payments?year=2026", nil), &payments)
	if len(payments) != 1 || payments[0].Label != "Ad hoc expense" {
		t.Fatalf("payments after import = %v, want one payment labeled 'Ad hoc expense'", payments)
	}

	var recurring []recurringDTO
	decodeBody(t, do(t, h, "GET", "/recurring", nil), &recurring)
	if len(recurring) != 1 || recurring[0].Label != "Rent" {
		t.Fatalf("recurring definitions after import = %v, want one definition labeled Rent", recurring)
	}
}
