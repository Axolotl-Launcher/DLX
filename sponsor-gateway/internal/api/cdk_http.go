package api

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/OwO-Network/DLX/sponsor-gateway/internal/entitlement"
)

type adminCDKDTO struct {
	ID         string                `json:"id"`
	BatchID    string                `json:"batch_id"`
	AmountFen  entitlement.Fen       `json:"amount_fen"`
	Status     entitlement.CDKStatus `json:"status"`
	RedeemedAt *time.Time            `json:"redeemed_at,omitempty"`
}

func toAdminCDKDTO(cdk entitlement.CDK) adminCDKDTO {
	return adminCDKDTO{
		ID: cdk.ID, BatchID: cdk.BatchID, AmountFen: cdk.AmountFen,
		Status: cdk.Status, RedeemedAt: cdk.RedeemedAt,
	}
}

func (s *Server) adminAuthorized(r *http.Request) bool {
	supplied := r.Header.Get("X-Admin-Token")
	if supplied == "" {
		supplied = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	}
	if s.AdminToken == "" {
		return false
	}
	expected := sha256.Sum256([]byte(s.AdminToken))
	actual := sha256.Sum256([]byte(supplied))
	return subtle.ConstantTimeCompare(expected[:], actual[:]) == 1
}

func (s *Server) adminCDKs(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuthorized(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin authorization required", "")
		return
	}
	repo, ok := s.Store.(CDKAdminStore)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "CDK service unavailable", "")
		return
	}
	switch r.Method {
	case http.MethodGet:
		cdks, err := repo.ListCDKs(r.Context())
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "CDK service unavailable", "")
			return
		}
		result := make([]adminCDKDTO, 0, len(cdks))
		for _, cdk := range cdks {
			result = append(result, toAdminCDKDTO(cdk))
		}
		writeJSON(w, http.StatusOK, result)
	case http.MethodPost:
		var req struct {
			AmountFen int64  `json:"amount_fen"`
			Quantity  int    `json:"quantity"`
			Note      string `json:"note"`
		}
		if err := decodeJSON(w, r, 16<<10, &req); err != nil || req.AmountFen <= 0 || req.Quantity <= 0 || req.Quantity > 10000 || len(req.Note) > 500 {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid CDK batch", "")
			return
		}
		if s.Hasher == nil {
			writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "CDK service unavailable", "")
			return
		}
		plaintexts := make([]string, req.Quantity)
		digests := make([]string, req.Quantity)
		for i := range plaintexts {
			raw := make([]byte, 24)
			if _, err := rand.Read(raw); err != nil {
				writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "CDK generation unavailable", "")
				return
			}
			plaintexts[i] = "cdk_" + hex.EncodeToString(raw)
			digests[i] = s.Hasher.Hash(plaintexts[i])
		}
		batchID, err := repo.CreateCDKs(r.Context(), req.AmountFen, req.Quantity, req.Note, digests)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "CDK creation failed", "")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"batch_id": batchID, "amount_fen": req.AmountFen, "quantity": req.Quantity, "note": req.Note, "cdks": plaintexts})
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", "")
	}
}

func (s *Server) redeemCDK(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", "")
		return
	}
	if s.Hasher == nil || s.Store == nil || s.Sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "CDK service unavailable", "")
		return
	}
	userID, ok := s.sessionUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "login required", "")
		return
	}
	repo, ok := s.Store.(CDKRedeemer)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "CDK service unavailable", "")
		return
	}
	var req struct {
		CDK string `json:"cdk"`
	}
	if err := decodeJSON(w, r, 16<<10, &req); err != nil || strings.TrimSpace(req.CDK) == "" {
		writeError(w, http.StatusBadRequest, "INVALID_CDK", "invalid CDK", "")
		return
	}
	entry, err := repo.RedeemCDK(r.Context(), s.Hasher.Hash(strings.TrimSpace(req.CDK)), userID, time.Now().UTC(), s.threshold())
	if err != nil {
		if errors.Is(err, entitlement.ErrCDKNotFound) || errors.Is(err, entitlement.ErrCDKUsed) {
			writeError(w, http.StatusConflict, "CDK_UNAVAILABLE", "CDK is unavailable", "")
			return
		}
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "CDK redemption failed", "")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"amount_fen": entry.AmountFen, "status": "redeemed"})
}
