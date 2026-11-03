package api

import (
	"net/http"

	"finance-backend/internal/store"
)

type categoryResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

func toCategoryResponse(c store.Category) categoryResponse {
	return categoryResponse{ID: c.ID, Name: c.Name, Kind: c.Kind}
}

func (s *Server) listCategories(w http.ResponseWriter, r *http.Request) {
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
