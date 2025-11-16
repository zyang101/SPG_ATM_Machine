"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  Thermometer,
  Power,
  Wind,
  Droplets,
  Users,
  LogOut,
  Flame,
  Snowflake,
  Home,
  X,
  AlertTriangle,
  Zap,
  CloudRain,
} from "lucide-react";
import { ProfileManager } from "@/components/profile-manager";
import { UserManager } from "@/components/user-manager";
import { ScheduleManager } from "@/components/schedule-manager";
import { AIChatbot } from "@/components/ai-chatbot";
import type { ThermostatProfile, HVACMode } from "@/lib/types";
import {
  getSensorsRecent,
  getEnergyConsumption,
  getWeatherRecent,
  getHVACState,
  setTargetTemperature,
  listSchedules,
  listProfiles,
  clearSession,
} from "@/lib/api";
import type { ScheduleDTO, ProfileDTO, HVACStateDTO } from "@/lib/api";

export function HomeownerDashboard() {
  const router = useRouter();
  const userName =
    typeof window !== "undefined" ? localStorage.getItem("userName") : "";

  // Initialize state with defaults, will be updated from API
  const [systemState, setSystemState] = useState({
    currentTemp: 68,
    targetTemp: 72,
    hvacMode: "heating" as HVACMode,
    indoorHumidity: 45,
    carbonMonoxide: 0,
    energyConsumption: 0,
    outdoorTemp: 55,
    outdoorHumidity: 50,
    precipitation: 0,
    currentProfileId: undefined as string | undefined,
    lastUpdated: new Date().toISOString(),
  });
  const [schedules, setSchedules] = useState<ScheduleDTO[]>([]);
  const [profiles, setProfiles] = useState<ProfileDTO[]>([]);
  const [showProfileManager, setShowProfileManager] = useState(false);
  const [showUserManager, setShowUserManager] = useState(false);
  const [showScheduleManager, setShowScheduleManager] = useState(false);
  const [pendingTarget, setPendingTarget] = useState<number | null>(null);
  const [currentTime, setCurrentTime] = useState(() =>
    new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" }),
  );
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

const SCHEDULE_CHECK_STORAGE_KEY = "lastScheduleCheckMs";
const initialScheduleCheck =
  typeof window !== "undefined"
    ? Number(localStorage.getItem(SCHEDULE_CHECK_STORAGE_KEY) ?? Date.now())
    : Date.now();
const lastScheduleCheckRef = useRef<number>(initialScheduleCheck);

const buildScheduleCandidates = (startTime: string, reference: Date): Date[] => {
  if (!startTime) return [];
  const trimmed = startTime.trim();
  const simpleTime = /^(\d{1,2}):(\d{2})(?::(\d{2}))?$/.exec(trimmed);
  if (simpleTime) {
    const hours = Number(simpleTime[1]);
    const minutes = Number(simpleTime[2]);
    const seconds = simpleTime[3] ? Number(simpleTime[3]) : 0;
    if (
      Number.isNaN(hours) ||
      Number.isNaN(minutes) ||
      Number.isNaN(seconds)
    ) {
      return [];
    }
    const today = new Date(reference);
    today.setHours(hours, minutes, seconds, 0);
    const yesterday = new Date(today);
    yesterday.setDate(today.getDate() - 1);
    return [today, yesterday];
  }
  const parsed = new Date(trimmed);
  if (Number.isNaN(parsed.getTime())) {
    return [];
  }
  return [parsed];
};

const applySchedulesIfDue = useCallback(
  async (rows: ScheduleDTO[]) => {
    if (!Array.isArray(rows) || rows.length === 0) {
      const nowPlaceholder = Date.now();
      lastScheduleCheckRef.current = nowPlaceholder;
      if (typeof window !== "undefined") {
        localStorage.setItem(
          SCHEDULE_CHECK_STORAGE_KEY,
          String(nowPlaceholder),
        );
      }
      return { applied: false, latestState: null as HVACStateDTO | null };
    }

    const nowMs = Date.now();
    const nowDate = new Date(nowMs);
    const lastCheck = lastScheduleCheckRef.current ?? nowMs;
    let latestState: HVACStateDTO | null = null;

    const dueSchedules = rows
      .flatMap((row) =>
        buildScheduleCandidates(row.start_time, nowDate).map((candidate) => ({
          timestamp: candidate.getTime(),
          row,
        })),
      )
      .filter(
        ({ timestamp }) =>
          timestamp > lastCheck &&
          timestamp <= nowMs &&
          !Number.isNaN(timestamp),
      )
      .sort((a, b) => a.timestamp - b.timestamp);

    for (const item of dueSchedules) {
      try {
        latestState = await setTargetTemperature(item.row.target_temp);
      } catch (error) {
        console.error("Failed to apply scheduled temperature:", error);
      }
    }

    lastScheduleCheckRef.current = nowMs;
    if (typeof window !== "undefined") {
      localStorage.setItem(SCHEDULE_CHECK_STORAGE_KEY, String(nowMs));
    }

    return { applied: dueSchedules.length > 0, latestState };
  },
  [],
);
const fetchSchedules = useCallback(async (): Promise<ScheduleDTO[]> => {
    const token =
      typeof window !== "undefined" ? localStorage.getItem("apiToken") : null;
    if (!token) {
      setSchedules([]);
    return [];
    }
    try {
      const data = await listSchedules();
      const normalized = Array.isArray(data) ? data : [];
      setSchedules(normalized);
    return normalized;
    } catch (error) {
      console.error("Failed to load schedules:", error);
      setSchedules([]);
    return [];
    }
  }, []);

  const fetchProfiles = useCallback(async () => {
    const token =
      typeof window !== "undefined" ? localStorage.getItem("apiToken") : null;
    if (!token) {
      setProfiles([]);
      return;
    }
    try {
      const rows = await listProfiles();
      setProfiles(Array.isArray(rows) ? rows : []);
    } catch (error) {
      console.error("Failed to load profiles:", error);
      setProfiles([]);
    }
  }, []);

  // Fetch all data from API on mount
  useEffect(() => {
    const token =
      typeof window !== "undefined" ? localStorage.getItem("apiToken") : null;
    if (!token) {
      setSchedules([]);
      setProfiles([]);
      return;
    }

    const loadData = async () => {
      try {
        const [hvac, sensors, weather, energy] = await Promise.all([
          getHVACState(),
          getSensorsRecent(1),
          getWeatherRecent(),
          getEnergyConsumption(),
        ]);
        const scheduleRows = await fetchSchedules();
        const scheduleResult = await applySchedulesIfDue(scheduleRows);

        const hvacState = scheduleResult.latestState ?? hvac;
        const latestIndoor = sensors[0];
        const latestWeather = weather[0];
        const currentTemp =
          typeof hvacState.current_temp === "number"
            ? Math.round(hvacState.current_temp)
            : latestIndoor && latestIndoor.indoor_temp != null
              ? Math.round(latestIndoor.indoor_temp)
              : 68;

        setSystemState({
          currentTemp,
          targetTemp:
            typeof hvacState.target_temp === "number"
              ? Math.round(hvacState.target_temp)
              : 72,
          hvacMode: (hvacState.mode === "heat"
            ? "heating"
            : hvacState.mode === "cool"
              ? "cooling"
              : hvacState.mode === "fan"
                ? "fan"
                : "off") as HVACMode,
          indoorHumidity:
            latestIndoor && latestIndoor.humidity != null
              ? Math.round(latestIndoor.humidity)
              : 45,
          carbonMonoxide:
            latestIndoor && latestIndoor.co_ppm != null
              ? Math.round(latestIndoor.co_ppm)
              : 0,
          energyConsumption:
            typeof energy?.kilowatts_used === "number" ? energy.kilowatts_used : 0,
          outdoorTemp:
            latestWeather && typeof latestWeather.temp === "number"
              ? Math.round(latestWeather.temp)
              : 55,
          outdoorHumidity:
            latestWeather && typeof latestWeather.humidity === "number"
              ? Math.round(latestWeather.humidity)
              : 50,
          precipitation:
            latestWeather && typeof latestWeather.precipitation_mm === "number"
              ? Math.round(latestWeather.precipitation_mm * 10) / 10
              : 0,
          currentProfileId: undefined,
          lastUpdated: new Date().toISOString(),
        });
        setPendingTarget(null);
      } catch (error) {
        console.error("Failed to load system data:", error);
        setSchedules([]);
        setProfiles([]);
      } finally {
        await fetchProfiles();
      }
    };

    loadData();
    // Refresh data every 30 seconds
    const interval = setInterval(loadData, 30000);
    return () => clearInterval(interval);
  }, [fetchSchedules, fetchProfiles, applySchedulesIfDue]);

  useEffect(() => {
    const interval = setInterval(() => {
      setCurrentTime(new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" }));
    }, 1000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, []);

  useEffect(() => {
    if (pendingTarget === null) return;
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }
    debounceRef.current = setTimeout(async () => {
      try {
        const hvac = await setTargetTemperature(pendingTarget);
        setSystemState((prev) => ({
          ...prev,
          targetTemp:
            typeof hvac.target_temp === "number"
              ? Math.round(hvac.target_temp)
              : prev.targetTemp,
          currentTemp:
            typeof hvac.current_temp === "number"
              ? Math.round(hvac.current_temp)
              : prev.currentTemp,
          hvacMode: (hvac.mode === "heat"
            ? "heating"
            : hvac.mode === "cool"
              ? "cooling"
              : hvac.mode === "fan"
                ? "fan"
                : "off") as HVACMode,
          lastUpdated: new Date().toISOString(),
        }));
      } catch (error) {
        console.error("Failed to update target temperature:", error);
      } finally {
        setPendingTarget(null);
      }
    }, 3000);
  }, [pendingTarget]);

  const handleOpenScheduleManager = () => {
    setShowScheduleManager(true);
  };

  const handleScheduleManagerClose = () => {
    setShowScheduleManager(false);
    void fetchSchedules();
  };

  const handleOpenProfileManager = () => {
    setShowProfileManager(true);
  };

  const handleProfileManagerClose = () => {
    setShowProfileManager(false);
    void fetchProfiles();
  };

  const handleSignOut = () => {
    clearSession();
    router.push("/");
  };

  const adjustTemp = (delta: number) => {
    const newTargetTemp = Math.max(
      60,
      Math.min(85, systemState.targetTemp + delta),
    );
    setSystemState({ ...systemState, targetTemp: newTargetTemp });
    setPendingTarget(newTargetTemp);
  };

  const handleSelectProfile = (profile: ThermostatProfile) => {
    const updatedState = {
      ...systemState,
      targetTemp: profile.targetTemp,
      currentProfileId: profile.id,
    };
    setSystemState(updatedState);
    setPendingTarget(profile.targetTemp);
  };

  const handleApplyProfile = (profile: ProfileDTO) => {
    handleSelectProfile({
      id: String(profile.id),
      name: profile.name,
      targetTemp: Math.round(profile.target_temp),
      createdAt: profile.created_at,
    });
  };

  const handleHvacModeChange = (mode: HVACMode) => {
    setSystemState({ ...systemState, hvacMode: mode });
  };

  const getModeColor = () => {
    switch (systemState.hvacMode) {
      case "heating":
        return "text-orange-500";
      case "cooling":
        return "text-sky-500";
      case "fan":
        return "text-primary";
      default:
        return "text-muted-foreground";
    }
  };

  const getModeGradient = () => {
    switch (systemState.hvacMode) {
      case "heating":
        return "from-orange-500/20 to-red-600/20";
      case "cooling":
        return "from-sky-500/20 to-blue-600/20";
      case "fan":
        return "from-primary/20 to-primary/20";
      default:
        return "from-zinc-500/20 to-zinc-600/20";
    }
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <div className="border-b border-border bg-card">
        <div className="max-w-2xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-sky-500 to-blue-600 flex items-center justify-center">
              <Home className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="font-semibold">Welcome, {userName}</h1>
              <p className="text-xs text-muted-foreground">Homeowner</p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <div className="text-right">
              <p className="text-xs text-muted-foreground uppercase tracking-wide">Current Time</p>
              <p className="text-sm font-semibold text-foreground">{currentTime}</p>
            </div>
            <Button variant="ghost" size="icon" onClick={handleSignOut}>
              <LogOut className="w-5 h-5" />
            </Button>
          </div>
        </div>
      </div>

      <div className="max-w-2xl mx-auto p-4 space-y-4">
        {/* Main Temperature Control */}
        <Card className={`p-8 bg-gradient-to-br ${getModeGradient()}`}>
          <div className="text-center space-y-6">
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground uppercase tracking-wide">
                Current Temperature
              </p>
              <div className="flex items-center justify-center gap-2">
                <Thermometer className={`w-8 h-8 ${getModeColor()}`} />
                <span className="text-6xl font-bold">
                  {systemState.currentTemp}°
                </span>
              </div>
            </div>

            <div className="flex items-center justify-center gap-4">
              <Button
                size="lg"
                variant="secondary"
                className="w-16 h-16 rounded-full text-2xl"
                onClick={() => adjustTemp(-1)}
              >
                −
              </Button>
              <div className="text-center min-w-[120px]">
                <p className="text-sm text-muted-foreground">Target</p>
                <p className="text-4xl font-bold">{systemState.targetTemp}°</p>
              </div>
              <Button
                size="lg"
                variant="secondary"
                className="w-16 h-16 rounded-full text-2xl"
                onClick={() => adjustTemp(1)}
              >
                +
              </Button>
            </div>

            <div className="flex items-center justify-center gap-2 text-sm text-muted-foreground">
              <span className={getModeColor()}>●</span>
              <span className="capitalize">
                {systemState.hvacMode === "off"
                  ? "System Off"
                  : `${systemState.hvacMode} Mode`}
              </span>
            </div>
          </div>
        </Card>

        {/* HVAC Mode Controls */}
        <div className="grid grid-cols-4 gap-2">
          <Button
            variant={systemState.hvacMode === "heating" ? "default" : "outline"}
            className="h-20 flex-col gap-2"
            onClick={() => handleHvacModeChange("heating")}
          >
            <Flame className="w-6 h-6" />
            <span className="text-xs">Heat</span>
          </Button>
          <Button
            variant={systemState.hvacMode === "cooling" ? "default" : "outline"}
            className="h-20 flex-col gap-2"
            onClick={() => handleHvacModeChange("cooling")}
          >
            <Snowflake className="w-6 h-6" />
            <span className="text-xs">Cool</span>
          </Button>
          <Button
            variant={systemState.hvacMode === "fan" ? "default" : "outline"}
            className="h-20 flex-col gap-2"
            onClick={() => handleHvacModeChange("fan")}
          >
            <Wind className="w-6 h-6" />
            <span className="text-xs">Fan</span>
          </Button>
          <Button
            variant={systemState.hvacMode === "off" ? "default" : "outline"}
            className="h-20 flex-col gap-2"
            onClick={() => handleHvacModeChange("off")}
          >
            <Power className="w-6 h-6" />
            <span className="text-xs">Off</span>
          </Button>
        </div>

        {/* Indoor Stats */}
        <div className="space-y-2">
          <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide px-1">
            Indoor Stats
          </h3>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <Card className="p-4">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center">
                  <Droplets className="w-5 h-5 text-blue-500" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Humidity</p>
                  <p className="text-2xl font-bold">{systemState.indoorHumidity}%</p>
                </div>
              </div>
            </Card>
            <Card className="p-4">
              <div className="flex items-center gap-3">
                <div
                  className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                    systemState.carbonMonoxide > 9
                      ? "bg-red-500/20"
                      : systemState.carbonMonoxide > 0
                        ? "bg-orange-500/20"
                        : "bg-green-500/10"
                  }`}
                >
                  <AlertTriangle
                    className={`w-5 h-5 ${
                      systemState.carbonMonoxide > 9
                        ? "text-red-500"
                        : systemState.carbonMonoxide > 0
                          ? "text-orange-500"
                          : "text-green-500"
                    }`}
                  />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">CO Level</p>
                  <p
                    className={`text-2xl font-bold ${
                      systemState.carbonMonoxide > 9
                        ? "text-red-500"
                        : systemState.carbonMonoxide > 0
                          ? "text-orange-500"
                          : ""
                    }`}
                  >
                    {systemState.carbonMonoxide} ppm
                  </p>
                </div>
              </div>
            </Card>
            <Card className="p-4">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-emerald-500/10 flex items-center justify-center">
                  <Zap className="w-5 h-5 text-emerald-500" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Energy</p>
                  <p className="text-2xl font-bold">{systemState.energyConsumption} kW</p>
                </div>
              </div>
            </Card>
          </div>
        </div>

        {/* Outdoor Stats */}
        <div className="space-y-2">
          <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide px-1">
            Outdoor Stats
          </h3>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <Card className="p-4">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-sky-500/10 flex items-center justify-center">
                  <Thermometer className="w-5 h-5 text-sky-500" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Temperature</p>
                  <p className="text-2xl font-bold">{systemState.outdoorTemp}°</p>
                </div>
              </div>
            </Card>
            <Card className="p-4">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center">
                  <Droplets className="w-5 h-5 text-blue-500" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Humidity</p>
                  <p className="text-2xl font-bold">{systemState.outdoorHumidity}%</p>
                </div>
              </div>
            </Card>
            <Card className="p-4">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-indigo-500/10 flex items-center justify-center">
                  <CloudRain className="w-5 h-5 text-indigo-500" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Precipitation</p>
                  <p className="text-2xl font-bold">{systemState.precipitation} mm</p>
                </div>
              </div>
            </Card>
          </div>
        </div>

        {/* AI Chatbot */}
        <AIChatbot outdoorTemp={systemState.outdoorTemp} />

        <Card className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="font-semibold text-foreground">Saved profiles</h3>
              <p className="text-sm text-muted-foreground">
                Quickly apply common comfort settings
              </p>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={handleOpenProfileManager}
            >
              Manage
            </Button>
          </div>
          {profiles.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              You haven&apos;t created any profiles yet. Use the manage button
              to add your first profile.
            </p>
          ) : (
            <div className="space-y-3">
              {profiles.slice(0, 4).map((profile) => {
                const profileId = String(profile.id);
                const isActive = systemState.currentProfileId === profileId;
                return (
                  <div
                    key={profile.id}
                    className={`flex items-center justify-between rounded-lg border px-3 py-2 ${
                      isActive ? "border-primary bg-primary/5" : "border-border"
                    }`}
                  >
                    <div>
                      <p className="font-medium text-foreground">
                        {profile.name}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        {Math.round(profile.target_temp)}° • Added{" "}
                        {new Date(profile.created_at).toLocaleDateString()}
                      </p>
                    </div>
                    <Button
                      variant={isActive ? "default" : "outline"}
                      size="sm"
                      onClick={() => handleApplyProfile(profile)}
                    >
                      {isActive ? "Active" : "Apply"}
                    </Button>
                  </div>
                );
              })}
              {profiles.length > 4 && (
                <p className="text-xs text-muted-foreground">
                  Showing {Math.min(4, profiles.length)} of {profiles.length}{" "}
                  profiles
                </p>
              )}
            </div>
          )}
        </Card>

        <Card className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="font-semibold text-foreground">
                Upcoming schedules
              </h3>
              <p className="text-sm text-muted-foreground">
                Automated temperature changes for your home
              </p>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={handleOpenScheduleManager}
            >
              Manage
            </Button>
          </div>
          {schedules.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              You haven&apos;t created any schedules yet. Use the manage button
              to add your first schedule.
            </p>
          ) : (
            <div className="space-y-3">
              {schedules.slice(0, 4).map((schedule) => (
                <div
                  key={schedule.id}
                  className="flex items-center justify-between"
                >
                  <div>
                    <p className="font-medium text-foreground">
                      {schedule.name || `Schedule #${schedule.id}`}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {schedule.start_time} • {Math.round(schedule.target_temp)}
                      °
                    </p>
                  </div>
                </div>
              ))}
              {schedules.length > 4 && (
                <p className="text-xs text-muted-foreground">
                  Showing {Math.min(4, schedules.length)} of {schedules.length}{" "}
                  schedules
                </p>
              )}
            </div>
          )}
        </Card>

        <Button
          variant="outline"
          className="w-full h-16 flex items-center justify-center gap-2 bg-transparent"
          onClick={() => setShowUserManager(true)}
        >
          <Users className="w-5 h-5" />
          <span className="text-sm">User Accounts</span>
        </Button>
      </div>

      {showProfileManager && (
        <ProfileManager onClose={handleProfileManagerClose} />
      )}
      {showUserManager && (
        <UserManager
          onClose={() => setShowUserManager(false)}
          userRole="homeowner"
        />
      )}
      {showScheduleManager && (
        <ScheduleManager onClose={handleScheduleManagerClose} />
      )}
    </div>
  );
}
