const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

type Session = {
  Token: string;
  UserID: number;
  Username: string;
  Role: "homeowner" | "technician" | "guest";
  HomeownerID: number;
  ExpiresAt: string;
};

function getToken(): string | null {
  return typeof window !== "undefined" ? localStorage.getItem("apiToken") : null;
}

function setSession(sess: Session) {
  if (typeof window === "undefined") return;
  localStorage.setItem("apiToken", sess.Token);
  localStorage.setItem("userRole", sess.Role);
  localStorage.setItem("userName", sess.Username);
  if (typeof document !== "undefined") {
    const expiresAt = new Date(sess.ExpiresAt);
    const fallback = new Date(Date.now() + 12 * 60 * 60 * 1000);
    const expiryDate = Number.isNaN(expiresAt.getTime()) ? fallback : expiresAt;
    const cookieExpiry = expiryDate.toUTCString();
    document.cookie = `apiToken=${sess.Token}; path=/; expires=${cookieExpiry}; SameSite=Lax`;
    document.cookie = `userRole=${sess.Role}; path=/; expires=${cookieExpiry}; SameSite=Lax`;
  }
}

export function clearSession() {
  if (typeof window !== "undefined") {
    localStorage.removeItem("apiToken");
    localStorage.removeItem("userRole");
    localStorage.removeItem("userName");
  }
  if (typeof document !== "undefined") {
    document.cookie = "apiToken=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; SameSite=Lax";
    document.cookie = "userRole=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; SameSite=Lax";
  }
}

async function api(path: string, init: RequestInit = {}) {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init.headers as Record<string, string> | undefined),
  };
  if (token) headers["Authorization"] = `Bearer ${token}`;
  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });
  if (!res.ok) {
    let errText = "";
    try {
      const errorData = await res.json();
      errText = errorData.error || errorData.message || res.statusText;
    } catch {
      errText = await res.text().catch(() => res.statusText);
    }
    throw new Error(`API ${res.status}: ${errText}`);
  }
  return res;
}

export async function loginHomeowner(username: string, password: string) {
  const res = await api("/auth/login_homeowner", {
    method: "POST",
    body: JSON.stringify({ Username: username, Password: password }),
  });
  const sess: Session = await res.json();
  setSession(sess);
  return sess;
}

export async function loginGuest(username: string, pin: string, homeowner: string) {
  const res = await api("/auth/login_guest", {
    method: "POST",
    body: JSON.stringify({ Username: username, PIN: pin, Homeowner: homeowner }),
  });
  const sess: Session = await res.json();
  setSession(sess);
  return sess;
}

export async function loginTechnician(username: string, password: string, homeowner: string) {
  const res = await api("/auth/login_technician", {
    method: "POST",
    body: JSON.stringify({ Username: username, Password: password, Homeowner: homeowner }),
  });
  const sess: Session = await res.json();
  setSession(sess);
  return sess;
}

// Profiles
export type ProfileDTO = {
  id: number;
  homeowner_id: number;
  name: string;
  target_temp: number;
  created_at: string;
};

export async function listProfiles(): Promise<ProfileDTO[]> {
  const res = await api("/profiles");
  return res.json();
}

export async function createProfile(input: { name: string; target_temp: number }): Promise<number> {
  const res = await api("/profiles", {
    method: "POST",
    body: JSON.stringify({ name: input.name, target_temp: input.target_temp }),
  });
  const data = await res.json();
  return data.id as number;
}

export async function deleteProfile(id: number) {
  await api(`/profiles/${id}`, { method: "DELETE" });
}

// Schedules
export type ScheduleDTO = {
  id: number;
  name: string;
  start_time: string;
  target_temp: number;
};

export async function listSchedules(): Promise<ScheduleDTO[]> {
  const res = await api("/schedules");
  return res.json();
}

export async function createSchedule(input: {
  name: string;
  start_time: string;
  target_temp: number;
}): Promise<number> {
  const res = await api("/schedules", {
    method: "POST",
    body: JSON.stringify({
      name: input.name,
      start_time: input.start_time,
      target_temp: input.target_temp,
    }),
  });
  const data = await res.json();
  return data.id as number;
}

export async function deleteSchedule(id: number) {
  await api(`/schedules/${id}`, { method: "DELETE" });
}

// Sensors & Weather
export type SensorReadingDTO = {
  id: number;
  homeowner_id: number;
  recorded_at: string;
  indoor_temp: number | null;
  humidity: number | null;
  co_ppm: number | null;
};

export async function getSensorsRecent(limit?: number): Promise<SensorReadingDTO[]> {
  const path = typeof limit === "number" ? `/sensors/recent?limit=${limit}` : "/sensors/recent";
  const res = await api(path);
  return res.json();
}

export type EnergyConsumptionDTO = {
  kilowatts_used: number;
};

export async function getEnergyConsumption(): Promise<EnergyConsumptionDTO> {
  const res = await api("/sensors/energy-consumption");
  return res.json();
}

export type WeatherDTO = {
  id: number;
  homeowner_id: number;
  recorded_at: string;
  temp: number | null;
  humidity: number | null;
  precipitation_mm: number | null;
};

export async function getWeatherRecent(): Promise<WeatherDTO[]> {
  const res = await api("/weather/recent");
  return res.json();
}

// HVAC State
export async function getHVACState(): Promise<{ mode: string; target_temp: number; current_temp: number }> {
  const res = await api("/hvac/state");
  return res.json();
}

export type HVACStateDTO = { mode: string; target_temp: number; current_temp: number };

export async function setTargetTemperature(target_temp: number): Promise<HVACStateDTO> {
  const res = await api("/hvac/state", { method: "POST", body: JSON.stringify({ target_temp }) });
  return res.json();
}

// Guest Management
export type GuestDTO = {
  id: number;
  username: string;
  role: string;
  created_at: string;
};

export async function listGuests(): Promise<GuestDTO[]> {
  const res = await api("/guests");
  return res.json();
}

export async function createGuest(username: string, pin: string): Promise<number> {
  const res = await api("/guests", {
    method: "POST",
    body: JSON.stringify({ username, pin }),
  });
  const data = await res.json();
  return data.id as number;
}

export async function deleteGuest(id: number): Promise<void> {
  await api(`/guests/${id}`, { method: "DELETE" });
}

// Technician Access Management
export type TechnicianDTO = {
  id: number;
  username: string;
  role: string;
};

export type TechnicianAccessDTO = {
  id: number;
  technician_id: number;
  technician_name: string;
  start_time: string;
  end_time: string;
  is_active: boolean;
};

export async function listTechnicians(): Promise<TechnicianDTO[]> {
  const res = await api("/technicians");
  return res.json();
}

export async function listTechnicianAccess(): Promise<TechnicianAccessDTO[]> {
  const res = await api("/technicians/access");
  return res.json();
}

export async function grantTechnicianAccess(
  technicianUsername: string,
  startTime: string,
  endTime: string
): Promise<number> {
  const res = await api("/technicians/access", {
    method: "POST",
    body: JSON.stringify({
      technician_username: technicianUsername,
      start_time: startTime,
      end_time: endTime,
    }),
  });
  const data = await res.json();
  return data.id as number;
}

export async function revokeTechnicianAccess(id: number): Promise<void> {
  await api(`/technicians/access/${id}`, { method: "DELETE" });
}

// Diagnostics
export type DiagnosticLogDTO = {
  ID: number;
  HomeownerID: number;
  LoggedAt: string;
  Level: "INFO" | "WARN" | "ERROR";
  Message: string;
};

export async function getDiagnostics(): Promise<DiagnosticLogDTO[]> {
  const url = "/diagnostics";
  const res = await api(url);
  return res.json();
}

export async function createDiagnosticLog(
  level: "INFO" | "WARN" | "ERROR",
  message: string,
): Promise<number> {
  const res = await api("/diagnostics", {
    method: "POST",
    body: JSON.stringify({
      level,
      message,
    }),
  });
  const data = await res.json();
  return data.id as number;
}

