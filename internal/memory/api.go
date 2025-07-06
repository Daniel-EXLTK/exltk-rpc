package memory

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type MemoryAPI struct {
	firestore *FirestoreMemory
}

func NewMemoryAPI(firestore *FirestoreMemory) *MemoryAPI {
	return &MemoryAPI{firestore: firestore}
}

func (api *MemoryAPI) SaveHandler(w http.ResponseWriter, r *http.Request) {
	var session SessionContext
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	session.UpdatedAt = NowUTC()
	if err := api.firestore.SaveSessionContext(r.Context(), session); err != nil {
		log.Printf("[ERROR] SaveSessionContext: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (api *MemoryAPI) GetHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session_id", http.StatusBadRequest)
		return
	}
	session, err := api.firestore.GetSessionContext(r.Context(), sessionID)
	if err != nil {
		log.Printf("[ERROR] GetSessionContext: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(session)
}

func (api *MemoryAPI) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.SessionID == "" {
		http.Error(w, "Missing session_id", http.StatusBadRequest)
		return
	}
	if err := api.firestore.DeleteSessionContext(r.Context(), req.SessionID); err != nil {
		log.Printf("[ERROR] DeleteSessionContext: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func NowUTC() (t time.Time) {
	return context.Background().Value("now").(time.Time)
}
