"use client"

import { useState, useEffect, useRef, useCallback } from "react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import {
  Thermometer,
  Power,
  Wind,
  Droplets,
  LogOut,
  Flame,
  Snowflake,
  Wrench,
  Activity,
  AlertCircle,
  CheckCircle,
  Users,
  X,
  Clock,
  AlertTriangle,
  Zap,
  CloudRain,
} from "lucide-react"
import { ProfileSelector } from "@/components/profile-selector"
import { ProfileManager } from "@/components/profile-manager"
import { UserManager } from "@/components/user-manager"
import { AIChatbot } from "@/components/ai-chatbot"
import type { ThermostatProfile, HVACMode } from "@/lib/types"
import {
  getSensorsRecent,
  getWeatherRecent,
  getHVACState,
  setTargetTemperature,
  getDiagnostics,
  createDiagnosticLog,
  clearSession,
  getEnergyConsumption,
  listSchedules,
  type DiagnosticLogDTO,
  type ScheduleDTO,
  type HVACStateDTO,
} from "@/lib/api"

export function TechnicianDashboard() {
  const router = useRouter()
  const userName = typeof window !== "undefined" ? localStorage.getItem("userName") : ""

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
  })
  const [showProfileManager, setShowProfileManager] = useState(false)
  const [showUserManager, setShowUserManager] = useState(false)
  const [showDiagnostics, setShowDiagnostics] = useState(false)
  const [diagnostics, setDiagnostics] = useState<DiagnosticLogDTO[]>([])
  const [isLoadingDiagnostics, setIsLoadingDiagnostics] = useState(false)
  const [isCreatingDiagnostic, setIsCreatingDiagnostic] = useState(false)
  const [newDiagnostic, setNewDiagnostic] = useState({ level: "INFO" as "INFO" | "WARN" | "ERROR", message: "" })
  const [pendingTarget, setPendingTarget] = useState<number | null>(null)
  const [profileRefreshTrigger, setProfileRefreshTrigger] = useState(0)
  const [currentTime, setCurrentTime] = useState(() =>
    new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" }),
  )
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

const SCHEDULE_CHECK_STORAGE_KEY = "lastScheduleCheckMs"
const initialScheduleCheck =
  typeof window !== "undefined"
    ? Number(localStorage.getItem(SCHEDULE_CHECK_STORAGE_KEY) ?? Date.now())
    : Date.now()
const lastScheduleCheckRef = useRef<number>(initialScheduleCheck)

const buildScheduleCandidates = (startTime: string, reference: Date): Date[] => {
  if (!startTime) return []
  const trimmed = startTime.trim()
  const simpleTime = /^(\d{1,2}):(\d{2})(?::(\d{2}))?$/.exec(trimmed)
  if (simpleTime) {
    const hours = Number(simpleTime[1])
    const minutes = Number(simpleTime[2])
    const seconds = simpleTime[3] ? Number(simpleTime[3]) : 0
    if (Number.isNaN(hours) || Number.isNaN(minutes) || Number.isNaN(seconds)) {
      return []
    }
    const today = new Date(reference)
    today.setHours(hours, minutes, seconds, 0)
    const yesterday = new Date(today)
    yesterday.setDate(today.getDate() - 1)
    return [today, yesterday]
  }
  const parsed = new Date(trimmed)
  if (Number.isNaN(parsed.getTime())) {
    return []
  }
  return [parsed]
}

const applySchedulesIfDue = useCallback(
  async (rows: ScheduleDTO[]) => {
    if (!Array.isArray(rows) || rows.length === 0) {
      const nowPlaceholder = Date.now()
      lastScheduleCheckRef.current = nowPlaceholder
      if (typeof window !== "undefined") {
        localStorage.setItem(SCHEDULE_CHECK_STORAGE_KEY, String(nowPlaceholder))
      }
      return { applied: false, latestState: null as HVACStateDTO | null }
    }

    const nowMs = Date.now()
    const nowDate = new Date(nowMs)
    const lastCheck = lastScheduleCheckRef.current ?? nowMs
    let latestState: HVACStateDTO | null = null

    const dueSchedules = rows
      .flatMap((row) =>
        buildScheduleCandidates(row.start_time, nowDate).map((candidate) => ({
          timestamp: candidate.getTime(),
          row,
        })),
      )
      .filter(
        ({ timestamp }) =>
          timestamp > lastCheck && timestamp <= nowMs && !Number.isNaN(timestamp),
      )
      .sort((a, b) => a.timestamp - b.timestamp)

    for (const item of dueSchedules) {
      try {
        latestState = await setTargetTemperature(item.row.target_temp)
      } catch (error) {
        console.error("Failed to apply scheduled temperature:", error)
      }
    }

    lastScheduleCheckRef.current = nowMs
    if (typeof window !== "undefined") {
      localStorage.setItem(SCHEDULE_CHECK_STORAGE_KEY, String(nowMs))
    }

    return { applied: dueSchedules.length > 0, latestState }
  },
  [],
)

  // Fetch data from API on mount
  useEffect(() => {
    const token = typeof window !== "undefined" ? localStorage.getItem("apiToken") : null
    if (!token) return

    const loadData = async () => {
      try {
        const [hvac, sensors, weather, energy, scheduleRowsRaw] = await Promise.all([
          getHVACState(),
          getSensorsRecent(1),
          getWeatherRecent(),
          getEnergyConsumption(),
          listSchedules(),
        ])
        const scheduleRows = Array.isArray(scheduleRowsRaw) ? scheduleRowsRaw : []
        const scheduleResult = await applySchedulesIfDue(scheduleRows)
        const hvacState = scheduleResult.latestState ?? hvac

        const latestIndoor = sensors[0]
        const latestWeather = weather[0]
        
        const currentTemp =
          typeof hvacState.current_temp === "number"
            ? Math.round(hvacState.current_temp)
            : latestIndoor && latestIndoor.indoor_temp != null
            ? Math.round(latestIndoor.indoor_temp)
            : 68

        setSystemState({
          currentTemp,
          targetTemp: typeof hvacState.target_temp === "number" ? Math.round(hvacState.target_temp) : 72,
          hvacMode: (hvacState.mode === "heat" ? "heating" : hvacState.mode === "cool" ? "cooling" : hvacState.mode === "fan" ? "fan" : "off") as HVACMode,
          indoorHumidity: latestIndoor && latestIndoor.humidity != null ? Math.round(latestIndoor.humidity) : 45,
          carbonMonoxide: latestIndoor && latestIndoor.co_ppm != null ? Math.round(latestIndoor.co_ppm) : 0,
          energyConsumption: typeof energy?.kilowatts_used === "number" ? energy.kilowatts_used : 0,
          outdoorTemp: latestWeather && typeof latestWeather.temp === "number" ? Math.round(latestWeather.temp) : 55,
          outdoorHumidity: latestWeather && typeof latestWeather.humidity === "number" ? Math.round(latestWeather.humidity) : 50,
          precipitation: latestWeather && typeof latestWeather.precipitation_mm === "number" ? Math.round(latestWeather.precipitation_mm * 10) / 10 : 0,
          currentProfileId: undefined,
          lastUpdated: new Date().toISOString(),
        })
        setPendingTarget(null)
      } catch (error) {
        console.error("Failed to load system data:", error)
      }
    }

    loadData()
    const interval = setInterval(loadData, 30000)
    return () => clearInterval(interval)
  }, [applySchedulesIfDue])

  useEffect(() => {
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }
    }
  }, [])

  useEffect(() => {
    const interval = setInterval(() => {
      setCurrentTime(new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" }))
    }, 1000)
    return () => clearInterval(interval)
  }, [])

  useEffect(() => {
    if (pendingTarget === null) return
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }
    debounceRef.current = setTimeout(async () => {
      try {
        const hvac = await setTargetTemperature(pendingTarget)
        setSystemState((prev) => ({
          ...prev,
          targetTemp: typeof hvac.target_temp === "number" ? Math.round(hvac.target_temp) : prev.targetTemp,
          currentTemp: typeof hvac.current_temp === "number" ? Math.round(hvac.current_temp) : prev.currentTemp,
          hvacMode: (hvac.mode === "heat" ? "heating" : hvac.mode === "cool" ? "cooling" : hvac.mode === "fan" ? "fan" : "off") as HVACMode,
          lastUpdated: new Date().toISOString(),
        }))
      } catch (error) {
        console.error("Failed to update target temperature:", error)
      } finally {
        setPendingTarget(null)
      }
    }, 3000)
  }, [pendingTarget])

  const handleSignOut = () => {
    clearSession()
    router.push("/")
  }

  const adjustTemp = (delta: number) => {
    const newTargetTemp = Math.max(60, Math.min(85, systemState.targetTemp + delta))
    setSystemState({ ...systemState, targetTemp: newTargetTemp })
    setPendingTarget(newTargetTemp)
  }

  const handleSelectProfile = (profile: ThermostatProfile) => {
    setSystemState({ 
      ...systemState, 
      targetTemp: profile.targetTemp, 
      currentProfileId: profile.id 
    })
    setPendingTarget(profile.targetTemp)
  }

  const handleHvacModeChange = (mode: HVACMode) => {
    setSystemState({ ...systemState, hvacMode: mode })
  }

  const loadDiagnostics = async () => {
    setIsLoadingDiagnostics(true)
    try {
      const logs = await getDiagnostics()
      setDiagnostics(logs)
      setShowDiagnostics(true)
    } catch (error) {
      console.error("Failed to load diagnostics:", error)
      alert("Failed to load diagnostics. Please try again.")
    } finally {
      setIsLoadingDiagnostics(false)
    }
  }

  const handleCreateDiagnostic = async () => {
    if (!newDiagnostic.message.trim()) {
      alert("Please enter a diagnostic message")
      return
    }

    setIsCreatingDiagnostic(true)
    try {
      await createDiagnosticLog(newDiagnostic.level, newDiagnostic.message)
      setNewDiagnostic({ level: "INFO", message: "" })
      await loadDiagnostics()
      alert("Diagnostic log created successfully")
    } catch (error) {
      console.error("Failed to create diagnostic:", error)
      alert("Failed to create diagnostic log. Please try again.")
    } finally {
      setIsCreatingDiagnostic(false)
    }
  }


  const getModeColor = () => {
    switch (systemState.hvacMode) {
      case "heating":
        return "text-orange-500"
      case "cooling":
        return "text-sky-500"
      case "fan":
        return "text-primary"
      default:
        return "text-muted-foreground"
    }
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <div className="border-b border-border bg-card">
        <div className="max-w-2xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-orange-500 to-red-600 flex items-center justify-center">
              <Wrench className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="font-semibold">Welcome, {userName}</h1>
              <p className="text-xs text-muted-foreground">Technician</p>
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
        {/* System Status */}
        <Card className="p-4 bg-gradient-to-br from-green-500/10 to-emerald-600/10 border-green-500/20">
          <div className="flex items-center gap-3">
            <CheckCircle className="w-6 h-6 text-green-500" />
            <div>
              <p className="font-semibold">System Status: Operational</p>
              <p className="text-sm text-muted-foreground">All systems functioning normally</p>
            </div>
          </div>
        </Card>

        {/* Main Temperature Control */}
        <Card className="p-6">
          <div className="text-center space-y-4">
            <div className="flex items-center justify-center gap-2">
              <Thermometer className={`w-6 h-6 ${getModeColor()}`} />
              <span className="text-5xl font-bold">{systemState.currentTemp}°</span>
            </div>

            <div className="flex items-center justify-center gap-4">
              <Button size="lg" variant="secondary" className="w-14 h-14 rounded-full" onClick={() => adjustTemp(-1)}>
                −
              </Button>
              <div className="text-center min-w-[100px]">
                <p className="text-xs text-muted-foreground">Target</p>
                <p className="text-3xl font-bold">{systemState.targetTemp}°</p>
              </div>
              <Button size="lg" variant="secondary" className="w-14 h-14 rounded-full" onClick={() => adjustTemp(1)}>
                +
              </Button>
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

        <Card className="p-4">
          <ProfileSelector 
            onSelectProfile={handleSelectProfile} 
            currentProfileId={systemState.currentProfileId} 
            refreshTrigger={profileRefreshTrigger}
            onManage={() => setShowProfileManager(true)}
          />
        </Card>

        {/* Diagnostics Section */}
        <Card className="p-4">
          <h3 className="font-semibold mb-4 flex items-center gap-2">
            <Activity className="w-5 h-5" />
            System Diagnostics
          </h3>
          <div className="space-y-3">
            <Button
              onClick={loadDiagnostics}
              disabled={isLoadingDiagnostics}
              className="w-full bg-orange-600 hover:bg-orange-700"
            >
              {isLoadingDiagnostics ? "Loading..." : "View Diagnostic Logs"}
            </Button>
            <Button
              onClick={() => setShowDiagnostics(true)}
              variant="outline"
              className="w-full"
            >
              Create Diagnostic Entry
            </Button>
          </div>
        </Card>

        {/* User Management */}
        <Button
          variant="outline"
          className="w-full h-16 flex items-center justify-center gap-2 bg-transparent"
          onClick={() => setShowUserManager(true)}
        >
          <Users className="w-5 h-5" />
          <span className="text-sm">Guest Accounts</span>
        </Button>
      </div>

      {showProfileManager && <ProfileManager onClose={() => setShowProfileManager(false)} onProfileChange={() => setProfileRefreshTrigger(prev => prev + 1)} />}
      {showUserManager && <UserManager onClose={() => setShowUserManager(false)} userRole="technician" />}
      
      {/* Test Results Modal */}
      {showDiagnostics && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden border">
            {/* Header */}
            <div className="bg-gradient-to-r from-orange-500 to-red-600 px-8 py-6 text-white">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center">
                    <Activity className="w-6 h-6" />
                  </div>
                  <div>
                    <h2 className="text-2xl font-bold">Diagnostic Logs</h2>
                    <p className="text-orange-100">System diagnostic entries and logs</p>
                  </div>
                </div>
                <Button 
                  variant="ghost" 
                  size="icon" 
                  onClick={() => setShowDiagnostics(false)}
                  className="text-white hover:bg-white/20"
                >
                  <X className="w-6 h-6" />
                </Button>
              </div>
            </div>

            <div className="p-8 overflow-y-auto max-h-[calc(90vh-120px)]">
              {/* Create New Diagnostic */}
              <Card className="p-6 mb-6 border-2 border-orange-200 bg-orange-50/50 dark:bg-orange-950/20 dark:border-orange-800">
                <h3 className="text-lg font-semibold mb-4">Create New Diagnostic Entry</h3>
                <div className="space-y-4">
                  <div>
                    <label className="text-sm font-medium mb-2 block">Level</label>
                    <select
                      value={newDiagnostic.level}
                      onChange={(e) => setNewDiagnostic({ ...newDiagnostic, level: e.target.value as "INFO" | "WARN" | "ERROR" })}
                      className="w-full p-2 border rounded"
                    >
                      <option value="INFO">INFO</option>
                      <option value="WARN">WARN</option>
                      <option value="ERROR">ERROR</option>
                    </select>
                  </div>
                  <div>
                    <label className="text-sm font-medium mb-2 block">Message</label>
                    <textarea
                      value={newDiagnostic.message}
                      onChange={(e) => setNewDiagnostic({ ...newDiagnostic, message: e.target.value })}
                      placeholder="Enter diagnostic message..."
                      className="w-full p-2 border rounded min-h-[100px]"
                    />
                  </div>
                  <Button
                    onClick={handleCreateDiagnostic}
                    disabled={isCreatingDiagnostic || !newDiagnostic.message.trim()}
                    className="bg-orange-600 hover:bg-orange-700"
                  >
                    {isCreatingDiagnostic ? "Creating..." : "Create Diagnostic Entry"}
                  </Button>
                </div>
              </Card>

              {/* Diagnostic Logs Summary */}
              {diagnostics.length > 0 && (
                <div className="mb-6">
                  <div className="grid grid-cols-3 gap-4">
                    <Card className="p-4 text-center">
                      <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-2">
                        <Activity className="w-6 h-6 text-blue-600" />
                      </div>
                      <p className="text-2xl font-bold text-blue-600">
                        {diagnostics.filter(d => d.Level === 'INFO').length}
                      </p>
                      <p className="text-sm text-muted-foreground">Info</p>
                    </Card>
                    <Card className="p-4 text-center">
                      <div className="w-12 h-12 bg-yellow-100 rounded-full flex items-center justify-center mx-auto mb-2">
                        <AlertCircle className="w-6 h-6 text-yellow-600" />
                      </div>
                      <p className="text-2xl font-bold text-yellow-600">
                        {diagnostics.filter(d => d.Level === 'WARN').length}
                      </p>
                      <p className="text-sm text-muted-foreground">Warnings</p>
                    </Card>
                    <Card className="p-4 text-center">
                      <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-2">
                        <AlertTriangle className="w-6 h-6 text-red-600" />
                      </div>
                      <p className="text-2xl font-bold text-red-600">
                        {diagnostics.filter(d => d.Level === 'ERROR').length}
                      </p>
                      <p className="text-sm text-muted-foreground">Errors</p>
                    </Card>
                  </div>
                </div>
              )}

              {/* Diagnostic Logs List */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-foreground">Recent Diagnostic Logs</h3>
                {diagnostics.length === 0 ? (
                  <Card className="p-8 text-center border-dashed">
                    <p className="text-muted-foreground">No diagnostic logs found</p>
                  </Card>
                ) : (
                  diagnostics.map((log) => (
                    <Card key={log.ID} className="p-4 border">
                      <div className="flex items-start justify-between">
                        <div className="flex items-center gap-3">
                          <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                            log.Level === 'INFO' ? 'bg-blue-100' :
                            log.Level === 'WARN' ? 'bg-yellow-100' : 'bg-red-100'
                          }`}>
                            {log.Level === 'INFO' ? (
                              <Activity className="w-5 h-5 text-blue-600" />
                            ) : log.Level === 'WARN' ? (
                              <AlertCircle className="w-5 h-5 text-yellow-600" />
                            ) : (
                              <AlertTriangle className="w-5 h-5 text-red-600" />
                            )}
                          </div>
                          <div>
                            <h4 className="font-semibold text-foreground">{log.Message}</h4>
                            <p className="text-sm text-muted-foreground">
                              {new Date(log.LoggedAt).toLocaleString()}
                            </p>
                          </div>
                        </div>
                        <Badge className={`${
                          log.Level === 'INFO' ? 'bg-blue-100 text-blue-800' :
                          log.Level === 'WARN' ? 'bg-yellow-100 text-yellow-800' : 'bg-red-100 text-red-800'
                        }`}>
                          {log.Level}
                        </Badge>
                      </div>
                    </Card>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
