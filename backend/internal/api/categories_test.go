package api_test

import (
	"net/http"
	"testing"
)

func TestCreateCategory_Success(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Groceries", "expense")
	if cat.Name != "Groceries" || cat.Kind != "expense" {
		t.Fatalf("category = %+v, want Groceries/expense", cat)
	}

	var categories []categoryDTO
	decodeBody(t, do(t, h, "GET", "/categories", nil), &categories)
	if len(categories) != 2 {
		t.Fatalf("categories = %v, want Groceries plus the default Others category", categories)
	}
}

func TestCreateCategory_InvalidKindRejected(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	resp := do(t, h, "POST", "/categories", map[string]any{"name": "Groceries", "kind": "savings"})
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestCreateCategory_NameRequired(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	resp := do(t, h, "POST", "/categories", map[string]any{"name": "", "kind": "expense"})
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestListCategories_AlwaysIncludesDefaultOthersCategory(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	var categories []categoryDTO
	decodeBody(t, do(t, h, "GET", "/categories", nil), &categories)
	if len(categories) != 1 || categories[0].Name != "Others" || !categories[0].IsDefault {
		t.Fatalf("categories on a fresh database = %v, want exactly the default Others category", categories)
	}
}

func TestRenameCategory_Success(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Groceries", "expense")

	resp := do(t, h, "PATCH", "/categories/"+itoa(cat.ID), map[string]any{"name": "Food"})
	requireStatus(t, resp, http.StatusOK)
	var renamed categoryDTO
	decodeBody(t, resp, &renamed)
	if renamed.Name != "Food" {
		t.Fatalf("name = %q, want Food", renamed.Name)
	}
}

func TestRenameCategory_NameRequired(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Groceries", "expense")

	resp := do(t, h, "PATCH", "/categories/"+itoa(cat.ID), map[string]any{"name": ""})
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestDeleteCategory_Success(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	cat := createCategory(t, h, "Groceries", "expense")

	resp := do(t, h, "DELETE", "/categories/"+itoa(cat.ID), nil)
	requireStatus(t, resp, http.StatusNoContent)

	var categories []categoryDTO
	decodeBody(t, do(t, h, "GET", "/categories", nil), &categories)
	if len(categories) != 1 || categories[0].Name != "Others" {
		t.Fatalf("categories after delete = %v, want only the default Others category", categories)
	}
}

func TestDeleteCategory_DefaultCategoryRejected(t *testing.T) {
	srv := newTestServer(t)
	h := srv.Routes()

	var categories []categoryDTO
	decodeBody(t, do(t, h, "GET", "/categories", nil), &categories)

	resp := do(t, h, "DELETE", "/categories/"+itoa(categories[0].ID), nil)
	requireStatus(t, resp, http.StatusBadRequest)
}

func TestDeleteCategory_InUseCategoryRejected(t *testing.T) {
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

	resp = do(t, h, "DELETE", "/categories/"+itoa(cat.ID), nil)
	requireStatus(t, resp, http.StatusBadRequest)
}
