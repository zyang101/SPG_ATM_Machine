PRAGMA foreign_keys = ON;

-- Users: homeowners, guests, technicians
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL,
  role TEXT NOT NULL CHECK(role IN ('homeowner','guest','technician')),
  password_hash TEXT,
  pin TEXT,
  homeowner_id INTEGER,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (homeowner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Technician access windows approved by homeowner
CREATE TABLE IF NOT EXISTS technician_access (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  homeowner_id INTEGER NOT NULL,
  technician_id INTEGER NOT NULL,
  start_time DATETIME NOT NULL,
  end_time DATETIME NOT NULL,
  FOREIGN KEY (homeowner_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (technician_id) REFERENCES users(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS profiles;
CREATE TABLE IF NOT EXISTS profiles (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  homeowner_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  target_temp REAL NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(homeowner_id, name),
  FOREIGN KEY (homeowner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Schedules
DROP TABLE IF EXISTS schedules;
CREATE TABLE IF NOT EXISTS schedules (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  homeowner_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  start_time TEXT NOT NULL, -- HH:MM
  target_temp REAL NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (homeowner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Energy usage samples
CREATE TABLE IF NOT EXISTS energy_usage (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  homeowner_id INTEGER NOT NULL,
  recorded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  kwh REAL NOT NULL,
  cost REAL,
  FOREIGN KEY (homeowner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Diagnostics logs
CREATE TABLE IF NOT EXISTS diagnostics_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  homeowner_id INTEGER,
  logged_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  level TEXT NOT NULL CHECK(level IN ('INFO','WARN','ERROR')),
  message TEXT NOT NULL,
  FOREIGN KEY (homeowner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Sensor readings (indoor)
CREATE TABLE IF NOT EXISTS sensor_readings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  homeowner_id INTEGER NOT NULL,
  recorded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  indoor_temp REAL,
  humidity REAL,
  co_ppm REAL,
  FOREIGN KEY (homeowner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Outdoor weather snapshots
CREATE TABLE IF NOT EXISTS outdoor_weather (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  homeowner_id INTEGER NOT NULL,
  recorded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  temp REAL,
  humidity REAL,
  precipitation_mm REAL,
  FOREIGN KEY (homeowner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- HVAC events
CREATE TABLE IF NOT EXISTS hvac_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  homeowner_id INTEGER NOT NULL,
  occurred_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  mode TEXT NOT NULL CHECK(mode IN ('heat','cool','auto','off','fan')),
  state TEXT NOT NULL CHECK(state IN ('on','off')),
  duration_sec INTEGER,
  FOREIGN KEY (homeowner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Simple key/value state store
CREATE TABLE IF NOT EXISTS system_state (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Guest verification requests (created by homeowner approval flow)
CREATE TABLE IF NOT EXISTS guest_verifications (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  token TEXT UNIQUE NOT NULL,
  guest_user_id INTEGER NOT NULL,
  homeowner_user_id INTEGER NOT NULL,
  target_url TEXT,
  status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','denied','consumed')),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME,
  FOREIGN KEY (guest_user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (homeowner_user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Helper table for guest-specific data (validateGuestPin checks this table)
CREATE TABLE IF NOT EXISTS guests (
  user_id INTEGER PRIMARY KEY,
  pin TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Login Attempts
CREATE TABLE IF NOT EXISTS login_attempts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL,
  password TEXT NOT NULL,
  role TEXT NOT NULL CHECK(role IN ('homeowner','guest','technician')),
  success BOOLEAN NOT NULL,
  attempted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- index to speed frontend polling by status
CREATE INDEX IF NOT EXISTS idx_guest_verifications_status ON guest_verifications(status);




