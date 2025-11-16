package api

import (
	"github.com/gin-gonic/gin"
	"mga_smart_thermostat/internal/logic"
	"time"
)

type weatherResponse struct {
	ID            int64     `json:"id"`
	HomeownerID   int64     `json:"homeowner_id"`
	RecordedAt    time.Time `json:"recorded_at"`
	OutdoorTemp   *float64  `json:"temp"`
	Humidity      *float64  `json:"humidity"`
	Precipitation *float64  `json:"precipitation_mm"`
}

func (s *Server) handleWeatherRecent(c *gin.Context) {
	sess, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Generate outdoor stats using random values
	stats, err := logic.FetchOutdoorStats()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch outdoor stats", "detail": err.Error()})
		return
	}

	temp := stats.TemperatureF
	humidity := stats.Humidity
	precipitation := stats.PrecipitationMM

	tempStats := weatherResponse{
		ID:            sess.UserID,
		HomeownerID:   sess.HomeownerID,
		RecordedAt:    stats.Timestamp,
		OutdoorTemp:   &temp,
		Humidity:      &humidity,
		Precipitation: &precipitation,
	}

	c.JSON(200, []weatherResponse{tempStats})
}
