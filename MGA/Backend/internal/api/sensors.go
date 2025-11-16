package api

import (
	"github.com/gin-gonic/gin"
	// "database/sql"
	"mga_smart_thermostat/internal/auth"
	"mga_smart_thermostat/internal/database"
	"mga_smart_thermostat/internal/logic"
	"time"
)

//api to get target temperature

// GET /sensors - Get all sensors state (legacy handler)
func (s *Server) handleGetSensorsRecent(c *gin.Context) {
	session, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	reading, err := s.sensors.ListRecent(c.Request.Context(), session.HomeownerID, 50)

	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	if len(reading) == 0 {
		updateSensorReadings(c, s)
		reading, err = s.sensors.ListRecent(c.Request.Context(), session.HomeownerID, 50)
		if err != nil {
			c.JSON(500, gin.H{"error": "db error"})
			return
		}
	}

	c.JSON(200, reading)

}

func updateSensorReadings(c *gin.Context, s *Server) {

	debug := false

	session, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	// Set busy timeout for SQLite
	_, err = s.db.Exec("PRAGMA busy_timeout = 5000") // 5 second timeout
	if err != nil {
		if debug {
			print("Failed to set busy timeout: %v", err)
		}
		c.JSON(500, gin.H{"error": "database configuration error"})
		return
	}

	sensors := &logic.SensorSuite{Db: s.db}
	currentReading := &database.SensorReading{
		HomeownerID: session.HomeownerID,
		RecordedAt:  time.Now(),
		IndoorTemp:  float64(sensors.GetTemperatureSensorReading()),
		Humidity:    float64(sensors.GetHumiditySensorReading()),
		COPPM:       float64(sensors.GetCOSensorReading()),
	}

	//retry 3 times incase something else has locked the database
	maxRetries := 3
	for i := range maxRetries {
		id, err := s.sensors.Insert(c.Request.Context(), currentReading)
		if err == nil {
			if debug {
				print("Successfully inserted sensor reading with ID: %d", id)
			}
			return
		}

		if i < maxRetries-1 {
			if debug {
				print("Insert attempt %d failed: %v - retrying...", i+1, err)
			}
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
			continue
		}
		if debug {
			print("Final insert attempt failed: %v", err)
		}
		c.JSON(500, gin.H{"error": "database error", "detail": err.Error()})
		return
	}

}

func homeownerIDForSession(sess auth.Session) (int64, bool) {
	switch sess.Role {
	case RoleHomeowner:
		return sess.UserID, true
	case RoleTechnician:
		if sess.HomeownerID == 0 {
			return 0, false
		}
		return sess.HomeownerID, true
	default:
		if sess.HomeownerID == 0 {
			return 0, false
		}
		return sess.HomeownerID, true
	}
}

func (s *Server) handleGetEnergyConsumption(c *gin.Context) {
	session, err := getSessionFromGin(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	if session.Role != RoleHomeowner && session.Role != RoleTechnician {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	homeownerID, ok := homeownerIDForSession(session)
	if !ok {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}

	count, err := s.hvac.CountByHomeowner(c.Request.Context(), homeownerID)
	if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	c.JSON(200, gin.H{
		"kilowatts_used": count * 20,
	})
}
