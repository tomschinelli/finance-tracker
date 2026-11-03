package api_test

import (
	"net/http"
	"testing"
)

func TestCreatePayment_AdHocDefaultsUnpaid(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Groceries", "expense")
	acct := createAccount(t, h, "Living", 0)

	resp := do(t, h, "POST", "/payments", map[string]any{
		"label":        "Supermarket",
		"category_id":  cat.ID,
		"account_id":   acct.ID,
		"amount_cents": 4500,
		"due_date":     "2026-03-10",
	})
	requireStatus(t, resp, http.StatusCreated)
	var p paymentDTO
	decodeBody(t, resp, &p)
	if p.RecurringDefinitionID != nil {
		t.Fatalf("recurring_definition_id = %v, want nil for an ad hoc payment", p.RecurringDefinitionID)
	}
	if p.PaidDate != nil {
		t.Fatalf("paid_date = %v, want nil until explicitly marked paid", p.PaidDate)
	}
}

func TestListPayments_FiltersByYearAndMonth(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Groceries", "expense")
	acct := createAccount(t, h, "Living", 0)

	dates := []string{"2026-01-10", "2026-02-10", "2027-01-10"}
	for _, d := range dates {
		resp := do(t, h, "POST", "/payments", map[string]any{
			"label":        "Supermarket",
			"category_id":  cat.ID,
			"account_id":   acct.ID,
			"amount_cents": 1000,
			"due_date":     d,
		})
		requireStatus(t, resp, http.StatusCreated)
	}

	var year2026 []paymentDTO
	decodeBody(t, do(t, h, "GET", "/payments?year=2026", nil), &year2026)
	if len(year2026) != 2 {
		t.Fatalf("payments in 2026 = %d, want 2", len(year2026))
	}

	var jan2026 []paymentDTO
	decodeBody(t, do(t, h, "GET", "/payments?year=2026&month=1", nil), &jan2026)
	if len(jan2026) != 1 {
		t.Fatalf("payments in 2026-01 = %d, want 1", len(jan2026))
	}
}

func TestUpdatePayment_MarkPaid(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Groceries", "expense")
	acct := createAccount(t, h, "Living", 0)

	resp := do(t, h, "POST", "/payments", map[string]any{
		"label":        "Supermarket",
		"category_id":  cat.ID,
		"account_id":   acct.ID,
		"amount_cents": 4500,
		"due_date":     "2026-03-10",
	})
	requireStatus(t, resp, http.StatusCreated)
	var created paymentDTO
	decodeBody(t, resp, &created)

	resp = do(t, h, "PATCH", "/payments/"+itoa(created.ID), map[string]any{"paid_date": "2026-03-11"})
	requireStatus(t, resp, http.StatusOK)
	var updated paymentDTO
	decodeBody(t, resp, &updated)
	if updated.PaidDate == nil || *updated.PaidDate != "2026-03-11" {
		t.Fatalf("paid_date = %v, want 2026-03-11", updated.PaidDate)
	}
}
