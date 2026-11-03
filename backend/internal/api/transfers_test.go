package api_test

import (
	"net/http"
	"testing"
)

func TestCreateTransfer_MovesBalanceBetweenAccounts(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	source := createAccount(t, h, "Savings", 1000) // becomes default
	dest := createAccount(t, h, "Fix Costs", 0)

	resp := do(t, h, "POST", "/transfers", map[string]any{
		"from_account_id": source.ID,
		"to_account_id":   dest.ID,
		"amount_cents":    400,
		"occurred_on":     "2026-01-01",
	})
	requireStatus(t, resp, http.StatusCreated)

	var accounts []accountDTO
	decodeBody(t, do(t, h, "GET", "/accounts", nil), &accounts)
	balances := map[int64]int64{}
	for _, a := range accounts {
		balances[a.ID] = a.CurrentBalanceCents
	}
	if balances[source.ID] != 600 {
		t.Fatalf("source balance = %d, want 600 after transferring out 400 from 1000", balances[source.ID])
	}
	if balances[dest.ID] != 400 {
		t.Fatalf("dest balance = %d, want 400 after transferring in 400 from 0", balances[dest.ID])
	}
}

func TestCreateTransfer_SameAccountRejected(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	acct := createAccount(t, h, "Savings", 1000)

	resp := do(t, h, "POST", "/transfers", map[string]any{
		"from_account_id": acct.ID,
		"to_account_id":   acct.ID,
		"amount_cents":    100,
		"occurred_on":     "2026-01-01",
	})
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestCreateTransfer_NonPositiveAmountRejected(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	source := createAccount(t, h, "Savings", 1000)
	dest := createAccount(t, h, "Fix Costs", 0)

	resp := do(t, h, "POST", "/transfers", map[string]any{
		"from_account_id": source.ID,
		"to_account_id":   dest.ID,
		"amount_cents":    0,
		"occurred_on":     "2026-01-01",
	})
	requireStatus(t, resp, http.StatusBadRequest)
}
