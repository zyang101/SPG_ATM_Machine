package logic

import (
	"database/sql"
	"math/rand"
	// "mga_smart_thermostat/internal/database"
)

type SensorSuite struct {
	Db *sql.DB
}

func (s *SensorSuite) GetHumiditySensorReading() int {
	randHumidity := rand.Intn(21) + 30
	return randHumidity
}

func (s *SensorSuite) GetTemperatureSensorReading() float64 {
	//db loookup for the current temperature - not random
	return float64(rand.Intn(21) + 30)
}

func (s *SensorSuite) GetCOSensorReading() int {
	randCO2 := rand.Intn(10)
	return randCO2
}
