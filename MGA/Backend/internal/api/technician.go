package api

import (
	"log"
	"strconv"
	"time"

	"mga_smart_thermostat/internal/database"

	"github.com/gin-gonic/gin"
)


// Technician Access Management Handlers
func (s *Server) handleGetTechnicianAccess(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Only homeowners can view technician access
	if sess.Role != RoleHomeowner {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	// Clean up expired access records before listing
	now := time.Now().UTC()
	if err := s.techAccess.DeleteExpired(c.Request.Context(), now); err != nil {
		// Log error but continue - don't fail the request
		log.Printf("Error cleaning up expired technician access: %v", err)
	}

	accessList, err := s.techAccess.ListForHomeowner(c.Request.Context(), sess.UserID)
	if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	// Enrich with technician usernames
	var result []map[string]any
	for _, a := range accessList {
		tech, err := s.users.GetByID(c.Request.Context(), a.TechnicianID)
		techUsername := ""
		if err == nil {
			techUsername = tech.Username
		}

		// Compare times in UTC to ensure consistency
		now := time.Now().UTC()
		startUTC := a.StartTime.UTC()
		endUTC := a.EndTime.UTC()
		isActive := now.After(startUTC) && now.Before(endUTC)

		result = append(result, map[string]any{
			"id":              a.ID,
			"technician_id":   a.TechnicianID,
			"technician_name": techUsername,
			"start_time":      a.StartTime.Format(time.RFC3339),
			"end_time":        a.EndTime.Format(time.RFC3339),
			"is_active":       isActive,
		})
	}
	c.JSON(200, result)
}

func (s *Server) handleGrantTechnicianAccess(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Only homeowners can grant technician access
	if sess.Role != RoleHomeowner {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	var body struct {
		TechnicianUsername string `json:"technician_username"`
		StartTime          string `json:"start_time"`
		EndTime            string `json:"end_time"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}

	// Get technician by username
	tech, err := s.users.GetByUsername(c.Request.Context(), body.TechnicianUsername)
	if err != nil || tech.Role != RoleTechnician {
		c.JSON(404, gin.H{"error": "technician not found"})
		return
	}

	// Parse times
	startTime, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid start_time"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, body.EndTime)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid end_time"})
		return
	}

	if !endTime.After(startTime) {
		c.JSON(400, gin.H{"error": "end_time must be after start_time"})
		return
	}

	id, err := s.techAccess.Grant(c.Request.Context(), &database.TechnicianAccess{
		HomeownerID:  sess.UserID,
		TechnicianID: tech.ID,
		StartTime:    startTime,
		EndTime:      endTime,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "grant failed"})
		return
	}

	c.JSON(201, gin.H{"id": id})
}

func (s *Server) handleRevokeTechnicianAccess(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Only homeowners can revoke technician access
	if sess.Role != RoleHomeowner {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	// Verify access belongs to homeowner
	access, err := s.techAccess.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "access not found"})
		return
	}

	if access.HomeownerID != sess.UserID {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	if err := s.techAccess.Revoke(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": "revoke failed"})
		return
	}

	c.JSON(200, gin.H{"status": "revoked"})
}





// List Technicians Handler
func (s *Server) handleListTechnicians(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Only homeowners can list technicians (for granting access)
	if sess.Role != RoleHomeowner {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	technicians, err := s.users.ListTechnicians(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	var result []map[string]any
	for _, t := range technicians {
		result = append(result, map[string]any{
			"id":       t.ID,
			"username": t.Username,
			"role":     t.Role,
		})
	}
	c.JSON(200, result)
}