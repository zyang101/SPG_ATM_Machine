package api

import (

	"encoding/json"

	"net/http"
	"time"

)


func (s *Server) handleEnergyRange(w http.ResponseWriter, r *http.Request) {
	sess, err := sessionFrom(r.Context())
	if err != nil {
		s.writeError(w, 401, "unauthorized")
		return
	}
	homeownerID := sess.UserID
	if sess.Role != RoleHomeowner {
		homeownerID = sess.HomeownerID
	}
	q := r.URL.Query()
	start, err1 := time.Parse(time.RFC3339, q.Get("start"))
	end, err2 := time.Parse(time.RFC3339, q.Get("end"))
	if err1 != nil || err2 != nil {
		s.writeError(w, 400, "invalid time range")
		return
	}
	items, err := s.energy.ListRange(r.Context(), homeownerID, start, end)
	if err != nil {
		s.writeError(w, 500, "db error")
		return
	}
	_ = json.NewEncoder(w).Encode(items)
}