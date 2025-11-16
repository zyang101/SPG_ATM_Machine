package api

import (
    "context"
    "time"

    "mga_smart_thermostat/internal/database"
)

const (
    RoleHomeowner  = "homeowner"
    RoleGuest      = "guest"
    RoleTechnician = "technician"
)

type TechnicianWindowChecker interface {
    IsAllowedNow(ctx context.Context, homeownerID, technicianID int64, now time.Time) (bool, error)
}

func TechnicianAllowed(ctx context.Context, repo TechnicianWindowChecker, homeownerID, technicianID int64) (bool, error) {
    return repo.IsAllowedNow(ctx, homeownerID, technicianID, time.Now().UTC())
}

// OwnershipForGuest returns the associated homeowner id for a guest.
func OwnershipForGuest(u *database.User) (int64, bool) {
    if u.Role != RoleGuest || !u.HomeownerID.Valid { return 0, false }
    return u.HomeownerID.Int64, true
}



