package api

import (
	"net/http"

	"finance-backend/internal/store"
)

type Server struct {
	q *store.Queries
}

func NewServer(q *store.Queries) *Server {
	return &Server{q: q}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /accounts", s.listAccounts)
	mux.HandleFunc("POST /accounts", s.createAccount)
	mux.HandleFunc("POST /accounts/{id}/adjustments", s.createAdjustment)
	mux.HandleFunc("GET /accounts/{id}/adjustments", s.listAdjustments)

	mux.HandleFunc("GET /categories", s.listCategories)
	mux.HandleFunc("POST /categories", s.createCategory)

	mux.HandleFunc("GET /recurring", s.listRecurring)
	mux.HandleFunc("POST /recurring", s.createRecurring)
	mux.HandleFunc("PATCH /recurring/{id}", s.updateRecurring)

	mux.HandleFunc("GET /payments", s.listPayments)
	mux.HandleFunc("POST /payments", s.createPayment)
	mux.HandleFunc("PATCH /payments/{id}", s.updatePayment)

	mux.HandleFunc("GET /overview/yearly", s.overviewYearly)
	mux.HandleFunc("GET /overview/monthly", s.overviewMonthly)

	return withCORS(mux)
}

// withCORS allows the Nuxt frontend (a different origin/port, both in native
// dev and in Docker Compose) to call this API from the browser. No
// credentials/cookies are used, so a wildcard origin is safe.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
