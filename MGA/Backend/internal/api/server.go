package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"mga_smart_thermostat/internal/auth"
	"mga_smart_thermostat/internal/database"

	"github.com/gin-gonic/gin"
)

type Server struct {
	db            *sql.DB
	sessions      *auth.SessionStore
	users         *database.UsersRepository
	profiles      *database.ProfilesRepository
	schedules     *database.SchedulesRepository
	energy        *database.EnergyRepository
	diagnostics   *database.DiagnosticsRepository
	sensors       *database.SensorsRepository
	weather       *database.WeatherRepository
	hvac          *database.HVACRepository
	techAccess    *database.TechnicianAccessRepository
	loginAttempts *database.LoginAttemptsRepository
}

func NewServer(db *sql.DB) *Server {
	return &Server{
		db:            db,
		sessions:      auth.NewSessionStore(12 * time.Hour),
		users:         database.NewUsersRepository(db),
		profiles:      database.NewProfilesRepository(db),
		schedules:     database.NewSchedulesRepository(db),
		energy:        database.NewEnergyRepository(db),
		diagnostics:   database.NewDiagnosticsRepository(db),
		sensors:       database.NewSensorsRepository(db),
		weather:       database.NewWeatherRepository(db),
		hvac:          database.NewHVACRepository(db),
		techAccess:    database.NewTechnicianAccessRepository(db),
		loginAttempts: database.NewLoginAttemptsRepository(db),
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)

	mux.HandleFunc("/auth/signup_homeowner", s.handleSignupHomeowner)
	mux.HandleFunc("/auth/login_homeowner", s.handleLoginHomeowner)
	mux.HandleFunc("/auth/login_guest", s.handleLoginGuest)
	mux.HandleFunc("/auth/login_technician", s.handleLoginTechnician)

	// Gin router for schedules endpoints
	gin.SetMode(gin.ReleaseMode)
	ginRouter := gin.New()
	ginRouter.Use(s.ginCORS())
	schedulesGroup := ginRouter.Group("/schedules")
	schedulesGroup.Use(s.requireAuthGin())
	{
		schedulesGroup.GET("", s.handleGetSchedules)
		schedulesGroup.POST("", s.handleCreateSchedule)
		schedulesGroup.DELETE("/:id", s.handleDeleteSchedule)
	}
	// Mount gin router as http.Handler
	mux.Handle("/schedules", ginRouter)
	mux.Handle("/schedules/", ginRouter)

	profilesGroup := ginRouter.Group("/profiles")
	profilesGroup.Use(s.requireAuthGin())
	{
		profilesGroup.GET("", s.handleListProfiles)
		profilesGroup.POST("", s.handleCreateProfile)
		profilesGroup.DELETE("/:id", s.handleDeleteProfile)
	}
	mux.Handle("/profiles", ginRouter)
	mux.Handle("/profiles/", ginRouter)

	sensorsGroup := ginRouter.Group("/sensors")
	sensorsGroup.Use(s.requireAuthGin())
	{
		sensorsGroup.GET("", s.handleGetSensorsRecent)
		sensorsGroup.GET("/recent", s.handleGetSensorsRecent)
		sensorsGroup.GET("/energy-consumption", s.handleGetEnergyConsumption)
	}
	mux.Handle("/sensors", ginRouter)
	mux.Handle("/sensors/", ginRouter)

	weatherGroup := ginRouter.Group("/weather")
	weatherGroup.Use(s.requireAuthGin())
	{
		weatherGroup.GET("/recent", s.handleWeatherRecent)
	}
	mux.Handle("/weather", ginRouter)
	mux.Handle("/weather/", ginRouter)

	mux.HandleFunc("/energy/range", s.requireAuth(s.handleEnergyRange))

	// HVAC state
	mux.HandleFunc("/hvac/state", s.requireAuth(s.handleHVACState))

	// Guest management endpoints
	guestsGroup := ginRouter.Group("/guests")
	guestsGroup.Use(s.requireAuthGin())
	{
		guestsGroup.GET("", s.handleGetGuests)
		guestsGroup.POST("", s.handleCreateGuest)
		guestsGroup.DELETE("/:id", s.handleDeleteGuest)
	}
	mux.Handle("/guests", ginRouter)
	mux.Handle("/guests/", ginRouter)

	// Technician access management endpoints
	techAccessGroup := ginRouter.Group("/technicians/access")
	techAccessGroup.Use(s.requireAuthGin())
	{
		techAccessGroup.GET("", s.handleGetTechnicianAccess)
		techAccessGroup.POST("", s.handleGrantTechnicianAccess)
		techAccessGroup.DELETE("/:id", s.handleRevokeTechnicianAccess)
	}
	mux.Handle("/technicians/access", ginRouter)
	mux.Handle("/technicians/access/", ginRouter)

	// Diagnostics endpoints
	diagnosticsGroup := ginRouter.Group("/diagnostics")
	diagnosticsGroup.Use(s.requireAuthGin())
	{
		diagnosticsGroup.GET("", s.handleGetDiagnostics)
		diagnosticsGroup.POST("", s.handleCreateDiagnostic)
	}
	mux.Handle("/diagnostics", ginRouter)
	mux.Handle("/diagnostics/", ginRouter)

	// List technicians endpoint (for homeowner to select when granting access)
	techniciansGroup := ginRouter.Group("/technicians")
	techniciansGroup.Use(s.requireAuthGin())
	{
		techniciansGroup.GET("", s.handleListTechnicians)
	}
	mux.Handle("/technicians", ginRouter)

	// register guest verification endpoints
	mux.HandleFunc("/api/guest/request-verification", RequestGuestVerificationHandler(s.db))
	mux.HandleFunc("/api/guest/verify", VerifyGuestByHomeownerHandler(s.db))
	mux.HandleFunc("/api/guest/status", GuestVerificationStatusHandler(s.db))

	return s.withCORS(s.withJSON(mux))
}

// CORS middleware for gin
func (s *Server) ginCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "http://localhost:3000" || origin == "https://localhost:3000" || origin == "" {
			c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func (s *Server) withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Allow CORS from frontend (defaults to localhost:3000). You can widen if needed via env.
func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// In dev, allow all localhost origins; adjust for production as needed
		if origin == "http://localhost:3000" || origin == "https://localhost:3000" || origin == "" {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) writeError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Middleware
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header.Get("Authorization")
		if !strings.HasPrefix(hdr, "Bearer ") {
			s.writeError(w, 401, "missing token")
			return
		}
		token := strings.TrimPrefix(hdr, "Bearer ")
		sess, err := s.sessions.Get(token)
		if err != nil {
			s.writeError(w, 401, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), ctxSessionKey{}, sess)
		next(w, r.WithContext(ctx))
	}
}

// Gin middleware for authentication
func (s *Server) requireAuthGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		hdr := c.GetHeader("Authorization")
		if hdr == "" || len(hdr) < 7 || hdr[:7] != "Bearer " {
			c.JSON(401, gin.H{"error": "missing token"})
			c.Abort()
			return
		}
		token := hdr[7:]
		sess, err := s.sessions.Get(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		c.Set("session", sess)
		c.Next()
	}
}

type ctxSessionKey struct{}

func sessionFrom(ctx context.Context) (auth.Session, error) {
	v := ctx.Value(ctxSessionKey{})
	if v == nil {
		return auth.Session{}, errors.New("no session")
	}
	sess, ok := v.(auth.Session)
	if !ok {
		return auth.Session{}, errors.New("bad session")
	}
	return sess, nil
}

// Admin helpers (homeowner creates guest or grants technician)
func CreateGuest(ctx context.Context, users *database.UsersRepository, homeownerID int64, username, pin string) (int64, error) {
	hash, err := auth.HashPassword(pin)
	if err != nil {
		return 0, err
	}
	return users.Create(ctx, &database.User{Username: username, Role: RoleGuest, PIN: sql.NullString{String: hash, Valid: true}, HomeownerID: sql.NullInt64{Int64: homeownerID, Valid: true}})
}

func CreateTechnician(ctx context.Context, users *database.UsersRepository, username, password string) (int64, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return 0, err
	}
	return users.Create(ctx, &database.User{Username: username, Role: RoleTechnician, PasswordHash: sql.NullString{String: hash, Valid: true}})
}

func GrantTechnician(ctx context.Context, repo *database.TechnicianAccessRepository, homeownerID, technicianID int64, start, end time.Time) (int64, error) {
	if !end.After(start) {
		return 0, fmt.Errorf("end must be after start")
	}
	return repo.Grant(ctx, &database.TechnicianAccess{HomeownerID: homeownerID, TechnicianID: technicianID, StartTime: start, EndTime: end})
}

// Simple seed endpoint (optional) for quick start
func (s *Server) SeedQuickStart(ctx context.Context) error {
	// If a homeowner exists, skip
	_, err := s.users.GetByUsername(ctx, "Logan")
	if err == nil {
		return nil
	}
	// Create homeowner, technician, guest and grant access
	hID, err := s.users.Create(ctx, &database.User{Username: "Logan", Role: RoleHomeowner, PasswordHash: sql.NullString{String: mustHash("Kostick2025!"), Valid: true}})
	if err != nil {
		return err
	}
	tID, err := s.users.Create(ctx, &database.User{Username: "Matt", Role: RoleTechnician, PasswordHash: sql.NullString{String: mustHash("Green2025!"), Valid: true}})
	if err != nil {
		return err
	}
	_, err = s.users.Create(ctx, &database.User{Username: "Mike", Role: RoleGuest, PIN: sql.NullString{String: mustHash("Rushanan2025!"), Valid: true}, HomeownerID: sql.NullInt64{Int64: hID, Valid: true}})
	if err != nil {
		return err
	}
	_, err = s.techAccess.Grant(ctx, &database.TechnicianAccess{HomeownerID: hID, TechnicianID: tID, StartTime: time.Now().Add(-time.Hour), EndTime: time.Now().Add(24 * time.Hour)})
	return err
}

func mustHash(s string) string {
	h, err := auth.HashPassword(s)
	if err != nil {
		log.Printf("hash error: %v", err)
		return s
	}
	return h
}
