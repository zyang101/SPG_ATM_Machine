"use client"

import { useState, useEffect, useRef } from "react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Thermometer, Power, Wind, Droplets, LogOut, Flame, Snowflake, UserCircle, AlertTriangle, CloudRain } from "lucide-react"
import { ProfileSelector } from "@/components/profile-selector"
import type { ThermostatProfile, HVACMode } from "@/lib/types"
import { getSensorsRecent, getWeatherRecent, getHVACState, setTargetTemperature, clearSession } from "@/lib/api"
import { AIChatbot } from "@/components/ai-chatbot"

export function GuestDashboard() {
  const router = useRouter()
  const userName = typeof window !== "undefined" ? localStorage.getItem("userName") : ""

  const [systemState, setSystemState] = useState({
    currentTemp: 68,
    targetTemp: 72,
    hvacMode: "heating" as HVACMode,
    indoorHumidity: 45,
    carbonMonoxide: 0,
    outdoorTemp: 55,
    outdoorHumidity: 50,
    precipitation: 0,
    currentProfileId: undefined as string | undefined,
    lastUpdated: new Date().toISOString(),
  })
  const [currentTime, setCurrentTime] = useState(() =>
    new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" })
  )
  const [pendingTarget, setPendingTarget] = useState<number | null>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Fetch data from API on mount
  useEffect(() => {
    const token = typeof window !== "undefined" ? localStorage.getItem("apiToken") : null
    if (!token) return

    const loadData = async () => {
      try {
        const [hvac, sensors, weather] = await Promise.all([
          getHVACState(),
          getSensorsRecent(1),
          getWeatherRecent(),
        ])
        const latestIndoor = sensors[0]
        const latestWeather = weather[0]
        const currentTemp =
          typeof hvac.current_temp === "number"
            ? Math.round(hvac.current_temp)
            : latestIndoor && latestIndoor.indoor_temp != null
            ? Math.round(latestIndoor.indoor_temp)
            : 68

        setSystemState({
          currentTemp,
          targetTemp: typeof hvac.target_temp === "number" ? Math.round(hvac.target_temp) : 72,
          hvacMode: (hvac.mode === "heat" ? "heating" : hvac.mode === "cool" ? "cooling" : hvac.mode === "fan" ? "fan" : "off") as HVACMode,
          indoorHumidity: latestIndoor && latestIndoor.humidity != null ? Math.round(latestIndoor.humidity) : 45,
          carbonMonoxide: latestIndoor && latestIndoor.co_ppm != null ? Math.round(latestIndoor.co_ppm) : 0,
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
  }, [])

  useEffect(() => {
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }
    }
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

  useEffect(() => {
    const interval = setInterval(() => {
      setCurrentTime(new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" }))
    }, 1000)
    return () => clearInterval(interval)
  }, [])

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
    // Note: Guest users typically cannot update HVAC state, this is UI-only
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

  const getModeGradient = () => {
    switch (systemState.hvacMode) {
      case "heating":
        return "from-orange-500/20 to-red-600/20"
      case "cooling":
        return "from-sky-500/20 to-blue-600/20"
      case "fan":
        return "from-primary/20 to-primary/20"
      default:
        return "from-zinc-500/20 to-zinc-600/20"
    }
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <div className="border-b border-border bg-card">
        <div className="max-w-2xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-zinc-500 to-zinc-600 flex items-center justify-center">
              <UserCircle className="w-5 h-5 text-white" />
            </div>
            <div>
              <h1 className="font-semibold">Guest Access</h1>
              <p className="text-xs text-muted-foreground">Limited Controls</p>
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
        {/* Info Banner */}
        <Card className="p-3 bg-muted/50">
          <p className="text-sm text-center text-muted-foreground">You have limited access to temperature controls</p>
        </Card>

        {/* Main Temperature Control */}
        <Card className={`p-8 bg-gradient-to-br ${getModeGradient()}`}>
          <div className="text-center space-y-6">
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground uppercase tracking-wide">Current Temperature</p>
              <div className="flex items-center justify-center gap-2">
                <Thermometer className={`w-8 h-8 ${getModeColor()}`} />
                <span className="text-6xl font-bold">{systemState.currentTemp}°</span>
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
              <span className="capitalize">{systemState.hvacMode === "off" ? "System Off" : `${systemState.hvacMode} Mode`}</span>
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
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
                <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                  systemState.carbonMonoxide > 9 ? 'bg-red-500/20' : systemState.carbonMonoxide > 0 ? 'bg-orange-500/20' : 'bg-green-500/10'
                }`}>
                  <AlertTriangle className={`w-5 h-5 ${
                    systemState.carbonMonoxide > 9 ? 'text-red-500' : systemState.carbonMonoxide > 0 ? 'text-orange-500' : 'text-green-500'
                  }`} />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">CO Level</p>
                  <p className={`text-2xl font-bold ${
                    systemState.carbonMonoxide > 9 ? 'text-red-500' : systemState.carbonMonoxide > 0 ? 'text-orange-500' : ''
                  }`}>
                    {systemState.carbonMonoxide} ppm
                  </p>
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
          <ProfileSelector onSelectProfile={handleSelectProfile} currentProfileId={systemState.currentProfileId} />
        </Card>

      </div>
    </div>
  )
}
