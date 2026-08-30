package api

import (
	"errors"
	"net/http"
	"net/url"
)

func parseAdminAPIKeyQuery(v url.Values) (AdminAPIKeyQuery, error) {
	page, size, err := parsePage(v, map[string]bool{"page": true, "page_size": true, "q": true, "status": true})
	if err != nil {
		return AdminAPIKeyQuery{}, err
	}
	if len(v.Get("q")) > 100 {
		return AdminAPIKeyQuery{}, errors.New("q too long")
	}
	if x := v.Get("status"); x != "" && !map[string]bool{"active": true, "revoked": true, "suspended": true}[x] {
		return AdminAPIKeyQuery{}, errors.New("invalid status")
	}
	return AdminAPIKeyQuery{Page: page, PageSize: size, Status: v.Get("status"), Q: v.Get("q")}, nil
}

func (s *Server) adminAPIKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", "")
		return
	}
	if !s.adminAuthorized(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin authorization required", "")
		return
	}
	q, err := parseAdminAPIKeyQuery(r.URL.Query())
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_QUERY", "invalid query", "")
		return
	}
	store, ok := s.Store.(AdminStore)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "admin service unavailable", "")
		return
	}
	v, e := store.ListAdminAPIKeys(r.Context(), q)
	if e != nil {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "admin service unavailable", "")
		return
	}
	writeJSON(w, http.StatusOK, v)
}