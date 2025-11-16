package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"mga_smart_thermostat/internal/database"
)


// HVAC State handlers
// GET returns current hvac mode and target temp from system_state; POST updates them and logs an event
func (s *Server) handleHVACState(w http.ResponseWriter, r *http.Request) {
	sess, err := sessionFrom(r.Context())
	if err != nil {
		s.writeError(w, 401, "unauthorized")
		return
	}
	homeownerID := sess.UserID
	if sess.Role != RoleHomeowner {
		homeownerID = sess.HomeownerID
	}

	db, _ := s.db, s.db
	stateRepo := database.NewSystemStateRepository(db)
	switch r.Method {
	case http.MethodGet:
		mode, _, _ := stateRepo.Get(r.Context(), "hvac_mode")
		if mode == "" {
			mode = "off"
		}
		targetStr, _, _ := stateRepo.Get(r.Context(), "target_temp")
		var target float64 = 72
		if targetStr != "" {
			fmt.Sscanf(targetStr, "%f", &target)
		}

		currentStr, _, _ := stateRepo.Get(r.Context(), "current_temp")
		var current float64 = target
		if currentStr != "" {
			fmt.Sscanf(currentStr, "%f", &current)
		} else {
			if latest, err := s.sensors.GetLatest(r.Context(), homeownerID); err == nil {
				current = latest.IndoorTemp
			}
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"mode":         mode,
			"target_temp":  target,
			"current_temp": current,
		})
	case http.MethodPost:
		var body struct {
			TargetTemp float64 `json:"target_temp"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			s.writeError(w, 400, "invalid json")
			return
		}

		latest, err := s.sensors.GetLatest(r.Context(), homeownerID)
		var currentTemp float64
		var humidity float64
		var coPPM float64
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				s.writeError(w, 500, "db error")
				return
			}
		} else {
			currentTemp = latest.IndoorTemp
			humidity = latest.Humidity
			coPPM = latest.COPPM
			if updateErr := s.sensors.UpdateByID(r.Context(), latest.ID, body.TargetTemp, humidity, coPPM); updateErr != nil {
				s.writeError(w, 500, "db error")
				return
			}
		}

		if errors.Is(err, sql.ErrNoRows) {
			currentTemp = body.TargetTemp
			if _, insertErr := s.sensors.Insert(r.Context(), &database.SensorReading{
				HomeownerID: homeownerID,
				IndoorTemp:  body.TargetTemp,
				Humidity:    humidity,
				COPPM:       coPPM,
			}); insertErr != nil {
				s.writeError(w, 500, "db error")
				return
			}
		}

		mode := ""
		if body.TargetTemp < currentTemp {
			mode = "cool"
		} else if body.TargetTemp > currentTemp {
			mode = "heat"
		} else {
			if existingMode, ok, _ := stateRepo.Get(r.Context(), "hvac_mode"); ok && existingMode != "" {
				mode = existingMode
			} else {
				mode = "off"
			}
		}

		if err := stateRepo.Set(r.Context(), "target_temp", fmt.Sprintf("%0.2f", body.TargetTemp)); err != nil {
			s.writeError(w, 500, "db error")
			return
		}
		if err := stateRepo.Set(r.Context(), "current_temp", fmt.Sprintf("%0.2f", body.TargetTemp)); err != nil {
			s.writeError(w, 500, "db error")
			return
		}
		if err := stateRepo.Set(r.Context(), "hvac_mode", mode); err != nil {
			s.writeError(w, 500, "db error")
			return
		}

		_, _ = s.hvac.Insert(r.Context(), &database.HVACEvent{HomeownerID: homeownerID, Mode: mode, State: "on"})

		_ = json.NewEncoder(w).Encode(map[string]any{
			"mode":         mode,
			"target_temp":  body.TargetTemp,
			"current_temp": body.TargetTemp,
		})
	default:
		s.writeError(w, 405, "method not allowed")
	}
}