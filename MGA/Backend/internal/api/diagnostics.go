package api

import (
	"database/sql"

	// "strconv"
	"time"

	"mga_smart_thermostat/internal/database"

	"github.com/gin-gonic/gin"
)

// Diagnostics Handlers
func (s *Server) handleGetDiagnostics(c *gin.Context) {
	sess, err := getSessionFromGin(c)

	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Guests cannot access diagnostics
	if sess.Role == RoleGuest || sess.Role == RoleHomeowner {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	homeownerID := sess.HomeownerID

	logs, err := s.diagnostics.ListRecent(c.Request.Context(), sql.NullInt64{Int64: homeownerID, Valid: true}, 50)
	if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	if len(logs) == 0 {
		s.diagnostics.Insert(c.Request.Context(), &database.DiagnosticLog{
			HomeownerID: sql.NullInt64{Int64: homeownerID, Valid: true},
			LoggedAt:    time.Now(),
			Level:       "INFO",
			Message:     "System created successfully.",
		})
	}

	logs, err = s.diagnostics.ListRecent(c.Request.Context(), sql.NullInt64{Int64: homeownerID, Valid: true}, 50)
	if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	print(logs)

	c.JSON(200, logs)
}

func (s *Server) handleCreateDiagnostic(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Only technicians can create diagnostics
	if sess.Role != RoleTechnician {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	homeownerID := sess.HomeownerID

	if homeownerID == 0 {
		c.JSON(403, gin.H{"error": "no homeowner access"})
		return
	}

	var body struct {
		Level   string `json:"level"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}

	if body.Level == "" || body.Message == "" {
		c.JSON(400, gin.H{"error": "missing fields"})
		return
	}

	if body.Level != "INFO" && body.Level != "WARN" && body.Level != "ERROR" {
		c.JSON(400, gin.H{"error": "invalid level"})
		return
	}

	id, err := s.diagnostics.Insert(c.Request.Context(), &database.DiagnosticLog{
		HomeownerID: sql.NullInt64{Int64: homeownerID, Valid: true},
		LoggedAt:    time.Now(),
		Level:       body.Level,
		Message:     body.Message,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "insert failed"})
		return
	}

	c.JSON(201, gin.H{"id": id})
}
