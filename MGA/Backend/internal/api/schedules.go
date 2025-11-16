package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"mga_smart_thermostat/internal/auth"
	"mga_smart_thermostat/internal/database"
)



// Helper to get session from gin context
func getSessionFromGin(c *gin.Context) (auth.Session, error) {
	v, exists := c.Get("session")
	if !exists {
		return auth.Session{}, gin.Error{Err: gin.Error{Err: nil}, Type: gin.ErrorTypePrivate, Meta: "no session"}
	}
	sess, ok := v.(auth.Session)
	if !ok {
		return auth.Session{}, gin.Error{Err: gin.Error{Err: nil}, Type: gin.ErrorTypePrivate, Meta: "bad session"}
	}
	return sess, nil
}

// GET /schedules - Get all schedules
func (s *Server) handleGetSchedules(c *gin.Context) {
	_, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Get all schedules from repository (single homeowner system)
	schedules, err := s.schedules.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	c.JSON(200, schedules)
}

// POST /schedules - Create a new schedule
func (s *Server) handleCreateSchedule(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Guests cannot create schedules
	if sess.Role == RoleGuest {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	// Get homeowner ID
	homeownerID := sess.UserID
	if sess.Role != RoleHomeowner {
		homeownerID = sess.HomeownerID
	}

	// Parse request body
	var body struct {
		Name        string  `json:"name" binding:"required"`
		StartTime   string  `json:"start_time" binding:"required"`
		TargetTemp  float64 `json:"target_temp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}

	// Create schedule
	schedule := &database.Schedule{
		HomeownerID: homeownerID,
		Name:        body.Name,
		StartTime:   body.StartTime,
		TargetTemp:  body.TargetTemp,
	}

	id, err := s.schedules.Create(c.Request.Context(), schedule)
	if err != nil {
		c.JSON(400, gin.H{"error": "create failed", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{"id": id})
}

// DELETE /schedules/:id - Delete a schedule
func (s *Server) handleDeleteSchedule(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Guests cannot delete schedules
	if sess.Role == RoleGuest {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	// Get schedule ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "bad id"})
		return
	}

	// Delete schedule
	if err := s.schedules.Delete(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": "delete failed"})
		return
	}

	c.JSON(200, gin.H{"status": "deleted"})
}
