package api

import (
	"database/sql"

	"strconv"
	"time"

	"mga_smart_thermostat/internal/auth"
	"mga_smart_thermostat/internal/database"

	"github.com/gin-gonic/gin"
)


// Guest Management Handlers
func (s *Server) handleGetGuests(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Only homeowners and technicians can list guests
	if sess.Role == RoleGuest {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	homeownerID := sess.UserID
	if sess.Role != RoleHomeowner {
		homeownerID = sess.HomeownerID
	}

	guests, err := s.users.ListGuestsForHomeowner(c.Request.Context(), homeownerID)
	if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	// Return guests without sensitive data
	var result []map[string]any
	for _, g := range guests {
		result = append(result, map[string]any{
			"id":         g.ID,
			"username":   g.Username,
			"role":       g.Role,
			"created_at": g.CreatedAt.Format(time.RFC3339),
		})
	}
	c.JSON(200, result)
}

func (s *Server) handleCreateGuest(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Only homeowners and technicians can create guests
	if sess.Role == RoleGuest {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	homeownerID := sess.UserID
	if sess.Role != RoleHomeowner {
		homeownerID = sess.HomeownerID
	}

	var body struct {
		Username string `json:"username"`
		PIN      string `json:"pin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}

	if body.Username == "" || body.PIN == "" {
		c.JSON(400, gin.H{"error": "missing fields"})
		return
	}

	pinHash, err := auth.HashPassword(body.PIN)
	if err != nil {
		c.JSON(500, gin.H{"error": "hash error"})
		return
	}

	id, err := s.users.Create(c.Request.Context(), &database.User{
		Username:    body.Username,
		Role:        RoleGuest,
		PIN:         sql.NullString{String: pinHash, Valid: true},
		HomeownerID: sql.NullInt64{Int64: homeownerID, Valid: true},
	})
	if err != nil {
		c.JSON(409, gin.H{"error": "user exists or create failed"})
		return
	}

	c.JSON(201, gin.H{"id": id})
}

func (s *Server) handleDeleteGuest(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Only homeowners and technicians can delete guests
	if sess.Role == RoleGuest {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	homeownerID := sess.UserID
	if sess.Role != RoleHomeowner {
		homeownerID = sess.HomeownerID
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	// Verify guest belongs to homeowner
	guest, err := s.users.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "guest not found"})
		return
	}

	if guest.Role != RoleGuest || !guest.HomeownerID.Valid || guest.HomeownerID.Int64 != homeownerID {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	if err := s.users.Delete(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": "delete failed"})
		return
	}

	c.JSON(200, gin.H{"status": "deleted"})
}