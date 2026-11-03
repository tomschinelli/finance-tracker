package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"finance-backend/internal/api"
	"finance-backend/internal/store"
)

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}

// newTestServer truncates every table (so each test starts from a
// completely empty database, ids restarting at 1) and returns a Server
// backed by the shared test container's pool.
func newTestServer(t *testing.T) *api.Server {
	t.Helper()
	const truncate = `truncate table
		payments, balance_adjustments, recurring_definitions, accounts, categories
		restart identity cascade`
	if _, err := testPool.Exec(context.Background(), truncate); err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
	return api.NewServer(testPool, store.New(testPool))
}

// do issues a single request against the handler and returns the raw
// response; callers decode the body with decodeBody.
func do(t *testing.T, h http.Handler, method, path string, body any) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Result()
}

func decodeBody(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}

func requireStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want %d; body: %s", resp.StatusCode, want, body)
	}
}

// --- shared response DTOs (black-box mirrors of the unexported api types) ---

type accountDTO struct {
	ID                   int64   `json:"id"`
	Name                 string  `json:"name"`
	Currency             string  `json:"currency"`
	StartingBalanceCents int64   `json:"starting_balance_cents"`
	CurrentBalanceCents  int64   `json:"current_balance_cents"`
	IsDefault            bool    `json:"is_default"`
	AccountNumber        *string `json:"account_number"`
}

type categoryDTO struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	IsDefault bool   `json:"is_default"`
}

type recurringDTO struct {
	ID             int64    `json:"id"`
	Label          string   `json:"label"`
	CategoryID     int64    `json:"category_id"`
	AccountID      int64    `json:"account_id"`
	AmountCents    int64    `json:"amount_cents"`
	IntervalMonths int32    `json:"interval_months"`
	AnchorDate     string   `json:"anchor_date"`
	Active         bool     `json:"active"`
	Tags           []string `json:"tags"`
}

type paymentDTO struct {
	ID                    int64   `json:"id"`
	RecurringDefinitionID *int64  `json:"recurring_definition_id"`
	Label                 string  `json:"label"`
	CategoryID            int64   `json:"category_id"`
	AccountID             int64   `json:"account_id"`
	AmountCents           int64   `json:"amount_cents"`
	DueDate               string  `json:"due_date"`
	PaidDate              *string `json:"paid_date"`
	ExcludedFromBudget    bool    `json:"excluded_from_budget"`
}

type adjustmentDTO struct {
	ID          int64   `json:"id"`
	AmountCents int64   `json:"amount_cents"`
	Note        *string `json:"note"`
	OccurredOn  string  `json:"occurred_on"`
	CreatedAt   string  `json:"created_at"`
}

// --- shared fixture helpers ---

func createCategory(t *testing.T, h http.Handler, name, kind string) categoryDTO {
	t.Helper()
	resp := do(t, h, "POST", "/categories", map[string]any{"name": name, "kind": kind})
	requireStatus(t, resp, http.StatusCreated)
	var c categoryDTO
	decodeBody(t, resp, &c)
	return c
}

func createAccount(t *testing.T, h http.Handler, name string, startingBalanceCents int64) accountDTO {
	t.Helper()
	resp := do(t, h, "POST", "/accounts", map[string]any{
		"name":                   name,
		"starting_balance_cents": startingBalanceCents,
		"currency":               "EUR",
	})
	requireStatus(t, resp, http.StatusCreated)
	var a accountDTO
	decodeBody(t, resp, &a)
	return a
}
