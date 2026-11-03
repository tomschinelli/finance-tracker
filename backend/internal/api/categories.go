package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"finance-backend/internal/store"
)

type categoryResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	IsDefault bool   `json:"is_default"`
}

func toCategoryResponse(c store.Category) categoryResponse {
	return categoryResponse{ID: c.ID, Name: c.Name, Kind: c.Kind, IsDefault: c.IsDefault}
}

// ensureDefaultCategory guarantees a catch-all "Others" category always
// exists, creating it on demand if it's missing (e.g. after /data/clear, or
// an import of a snapshot taken before this category existed).
func ensureDefaultCategory(ctx context.Context, q *store.Queries) (store.Category, error) {
	cat, err := q.GetDefaultCategory(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return q.CreateDefaultCategory(ctx)
	}
	return cat, err
}

func (s *Server) listCategories(w http.ResponseWriter, r *http.Request) {
	if _, err := ensureDefaultCategory(r.Context(), s.q); err != nil {
		handleDBError(w, err)
		return
	}
	cats, err := s.q.ListCategories(r.Context())
	if err != nil {
		handleDBError(w, err)
		return
	}
	out := make([]categoryResponse, 0, len(cats))
	for _, c := range cats {
		out = append(out, toCategoryResponse(c))
	}
	writeJSON(w, http.StatusOK, out)
}

type createCategoryRequest struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

func (s *Server) createCategory(w http.ResponseWriter, r *http.Request) {
	var req createCategoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Kind != "income" && req.Kind != "expense" {
		writeError(w, http.StatusBadRequest, "kind must be income or expense")
		return
	}

	cat, err := s.q.CreateCategory(r.Context(), store.CreateCategoryParams{Name: req.Name, Kind: req.Kind})
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toCategoryResponse(cat))
}

type renameCategoryRequest struct {
	Name string `json:"name"`
}

func (s *Server) renameCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req renameCategoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	cat, err := s.q.RenameCategory(r.Context(), store.RenameCategoryParams{ID: id, Name: req.Name})
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toCategoryResponse(cat))
}

func (s *Server) deleteCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	cat, err := s.q.GetCategory(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	if cat.IsDefault {
		writeError(w, http.StatusBadRequest, "the default category cannot be deleted")
		return
	}

	if err := s.q.DeleteCategory(r.Context(), id); err != nil {
		handleDBError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
