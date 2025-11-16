package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"mga_smart_thermostat/internal/database"
)

func (s *Server) handleListProfiles(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	homeownerID := sess.UserID
	if sess.Role != RoleHomeowner {
		homeownerID = sess.HomeownerID
	}
	if homeownerID == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "no homeowner context"})
		return
	}

	profiles, err := s.profiles.ListByHomeowner(c.Request.Context(), homeownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	c.JSON(http.StatusOK, profiles)
}

func (s *Server) handleCreateProfile(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if sess.Role == RoleGuest {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	homeownerID := sess.UserID
	if sess.Role != RoleHomeowner {
		homeownerID = sess.HomeownerID
	}
	if homeownerID == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "no homeowner context"})
		return
	}

	var body struct {
		Name       string  `json:"name"`
		TargetTemp float64 `json:"target_temp"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing name"})
		return
	}

	if body.TargetTemp < 40 || body.TargetTemp > 95 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_temp out of range"})
		return
	}

	id, err := s.profiles.Create(c.Request.Context(), &database.Profile{
		HomeownerID: homeownerID,
		Name:        body.Name,
		TargetTemp:  body.TargetTemp,
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "create failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (s *Server) handleDeleteProfile(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if sess.Role == RoleGuest {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := s.profiles.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
