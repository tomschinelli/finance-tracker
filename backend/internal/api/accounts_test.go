package api_test

import (
	"net/http"
	"testing"
)

func TestCreateAccount_FirstAccountBecomesDefault(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	first := createAccount(t, h, "Savings", 1000)
	if !first.IsDefault {
		t.Fatalf("first account created should be the default")
	}
	if first.CurrentBalanceCents != 1000 {
		t.Fatalf("current_balance_cents = %d, want 1000 for a brand new account", first.CurrentBalanceCents)
	}

	second := createAccount(t, h, "Fix Costs", 0)
	if second.IsDefault {
		t.Fatalf("second account should not become the default")
	}
}

func TestCreateAccount_NameRequired(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	resp := do(t, h, "POST", "/accounts", map[string]any{"starting_balance_cents": 0})
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestSetDefaultAccount_MovesDefaultFlag(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	first := createAccount(t, h, "Savings", 0)
	second := createAccount(t, h, "Fix Costs", 0)

	resp := do(t, h, "POST", "/accounts/"+itoa(second.ID)+"/default", nil)
	requireStatus(t, resp, http.StatusOK)
	var updated accountDTO
	decodeBody(t, resp, &updated)
	if !updated.IsDefault {
		t.Fatalf("account set as default should report is_default = true")
	}

	var accounts []accountDTO
	decodeBody(t, do(t, h, "GET", "/accounts", nil), &accounts)
	byID := map[int64]bool{}
	for _, a := range accounts {
		byID[a.ID] = a.IsDefault
	}
	if byID[first.ID] {
		t.Fatalf("original default account should no longer be default")
	}
	if !byID[second.ID] {
		t.Fatalf("new default account should be default")
	}
}

func TestCreateAdjustment_AllowedOnNonDefaultAccount(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	createAccount(t, h, "Savings", 0) // becomes default
	other := createAccount(t, h, "Fix Costs", 1000)

	resp := do(t, h, "POST", "/accounts/"+itoa(other.ID)+"/adjustments", map[string]any{
		"amount_cents": 100,
		"occurred_on":  "2026-01-01",
	})
	requireStatus(t, resp, http.StatusCreated)

	var accounts []accountDTO
	decodeBody(t, do(t, h, "GET", "/accounts", nil), &accounts)
	byID := map[int64]int64{}
	for _, a := range accounts {
		byID[a.ID] = a.CurrentBalanceCents
	}
	if byID[other.ID] != 1100 {
		t.Fatalf("accounts = %v, want balance 1100 after +100 adjustment on a 1000 starting balance", accounts)
	}
}

func TestCreateAdjustment_AllowedOnDefaultAccount(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	acct := createAccount(t, h, "Savings", 1000)

	resp := do(t, h, "POST", "/accounts/"+itoa(acct.ID)+"/adjustments", map[string]any{
		"amount_cents": 500,
		"note":         "salary",
		"occurred_on":  "2026-01-01",
	})
	requireStatus(t, resp, http.StatusCreated)

	var accounts []accountDTO
	decodeBody(t, do(t, h, "GET", "/accounts", nil), &accounts)
	if len(accounts) != 1 || accounts[0].CurrentBalanceCents != 1500 {
		t.Fatalf("accounts = %v, want balance 1500 after +500 adjustment on a 1000 starting balance", accounts)
	}
}
