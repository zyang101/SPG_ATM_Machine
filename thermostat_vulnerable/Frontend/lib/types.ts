export type UserRole = "homeowner" | "technician" | "guest"

export type HVACMode = "off" | "heating" | "cooling" | "fan"

export interface ThermostatProfile {
  id: string
  name: string
  targetTemp: number
  createdAt: string
}

export interface GuestUser {
  id: string
  name: string
  pin: string
  createdBy: string
  createdAt: string
}

export interface TechnicianAccess {
  id: string
  technicianName: string
  accessGrantedBy: string
  grantedAt: string
  expiresAt: string
  isActive: boolean
}
