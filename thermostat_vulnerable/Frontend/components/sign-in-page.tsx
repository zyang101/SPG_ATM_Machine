"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card } from "@/components/ui/card"
import { Thermometer, User, Wrench, UserCircle } from "lucide-react"
import { loginHomeowner, loginGuest, loginTechnician } from "@/lib/api"

type UserRole = "homeowner" | "technician" | "guest"

export function SignInPage() {
  const router = useRouter()
  const [selectedRole, setSelectedRole] = useState<UserRole | null>(null)
  const [username, setUsername] = useState("")
  const [secret, setSecret] = useState("")
  const [homeowner, setHomeowner] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [failureCount, setFailureCount] = useState(0)
  const [showLockoutVideo, setShowLockoutVideo] = useState(false)

  useEffect(() => {
    setUsername("")
    setSecret("")
    setHomeowner("")
  }, [selectedRole])

  const handleSignIn = async () => {
    if (!selectedRole) return
    setIsLoading(true)
    try {
      if (selectedRole === "homeowner") {
        if (!username || !secret) return
        await loginHomeowner(username, secret)
      } else if (selectedRole === "technician") {
        if (!username || !secret || !homeowner) return
        await loginTechnician(username, secret, homeowner)
      } else if (selectedRole === "guest") {
        if (!username || !secret || !homeowner) return
        await loginGuest(username, secret, homeowner)
      }
      setFailureCount(0)
      router.push(`/dashboard/${selectedRole}`)
    } catch (e) {
      // Extract error message from API response
      let errorMessage = "Sign in failed. Check credentials and try again."
      if (e instanceof Error) {
        const message = e.message
        // Parse API error response
        if (message.includes("401")) {
          errorMessage = "❌ Invalid credentials: Username or password is incorrect."
        } else if (message.includes("403")) {
          if (message.includes("homeowner not found")) {
            errorMessage = "❌ Homeowner not found: Please check the homeowner username."
          } else if (message.includes("access window")) {
            errorMessage = "❌ Access window issue: The access period has expired or hasn't started yet. Please contact the homeowner."
          } else {
            errorMessage = "❌ Access denied: " + message
          }
        } else if (message.includes("400")) {
          errorMessage = "❌ Invalid request: Please check all fields are filled correctly."
        } else {
          errorMessage = "❌ Error: " + message
        }
      }
      const nextFailureCount = failureCount + 1
      setFailureCount(nextFailureCount)
      if (nextFailureCount >= 6) {
        setShowLockoutVideo(true)
      } else {
        alert(errorMessage)
      }
    } finally {
      setIsLoading(false)
    }
  }

  const roles = [
    {
      id: "homeowner" as UserRole,
      title: "Homeowner",
      description: "Full system control and management",
      icon: User,
      color: "from-sky-500 to-blue-600",
    },
    {
      id: "technician" as UserRole,
      title: "Technician",
      description: "System diagnostics and maintenance",
      icon: Wrench,
      color: "from-orange-500 to-red-600",
    },
    {
      id: "guest" as UserRole,
      title: "Guest",
      description: "Limited temperature control",
      icon: UserCircle,
      color: "from-zinc-500 to-zinc-600",
    },
  ]

  return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center p-4 relative">
      {showLockoutVideo && (
        <div className="fixed inset-0 z-50 bg-black">
          <iframe
            className="w-full h-full"
            src="https://www.youtube.com/embed/f4fOc0ouBIA?autoplay=1&controls=0&modestbranding=1&rel=0"
            title="Security Warning"
            allow="autoplay; encrypted-media"
            allowFullScreen
          />
        </div>
      )}
      <div className={`w-full max-w-md space-y-8 ${showLockoutVideo ? "pointer-events-none select-none opacity-0" : ""}`}>
        {/* Header */}
        <div className="text-center space-y-2">
          <div className="flex justify-center mb-4">
            <div className="w-16 h-16 rounded-full bg-gradient-to-br from-sky-500 to-blue-600 flex items-center justify-center">
              <Thermometer className="w-8 h-8 text-white" />
            </div>
          </div>
          <h1 className="text-3xl font-bold text-balance">Smart Thermostat</h1>
          <p className="text-muted-foreground text-pretty">Select your role to access the system</p>
        </div>

        {/* Role Selection */}
        <div className="space-y-3">
          {roles.map((role) => {
            const Icon = role.icon
            const isSelected = selectedRole === role.id

            return (
              <Card
                key={role.id}
                className={`p-4 cursor-pointer transition-all border-2 ${
                  isSelected 
                    ? "border-primary bg-primary/10 shadow-md scale-[1.02]" 
                    : "border-border hover:border-primary/50 hover:bg-muted/50"
                }`}
                onClick={() => setSelectedRole(role.id)}
              >
                <div className="flex items-start gap-4">
                  <div
                    className={`w-12 h-12 rounded-lg bg-gradient-to-br ${role.color} flex items-center justify-center flex-shrink-0`}
                  >
                    <Icon className="w-6 h-6 text-white" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <h3 className="font-semibold text-lg">{role.title}</h3>
                    <p className="text-sm text-muted-foreground text-pretty">{role.description}</p>
                  </div>
                </div>
              </Card>
            )
          })}
        </div>

        {/* Credential Inputs */}
        {selectedRole && (
          <div className="space-y-4 animate-in fade-in slide-in-from-bottom-4 duration-300">
            {/* Username */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Username</label>
              <Input
                type="text"
                placeholder={selectedRole === "homeowner" ? "" : selectedRole === "technician" ? "" : ""}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleSignIn()}
                className="h-12"
              />
            </div>

            {/* Password or PIN */}
            <div className="space-y-2">
              <label className="text-sm font-medium">{selectedRole === "guest" ? "PIN" : "Password"}</label>
              <Input
                type="password"
                placeholder={selectedRole === "guest" ? "" : ""}
                value={secret}
                onChange={(e) => setSecret(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleSignIn()}
                className="h-12"
              />
            </div>

            {/* Homeowner target for guest/technician */}
            {(selectedRole === "guest" || selectedRole === "technician") && (
              <div className="space-y-2">
                <label className="text-sm font-medium">Homeowner Username</label>
                <Input
                  type="text"
                  placeholder=""
                  value={homeowner}
                  onChange={(e) => setHomeowner(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleSignIn()}
                  className="h-12"
                />
              </div>
            )}

            <Button
              onClick={handleSignIn}
              disabled={isLoading}
              className="w-full h-12 text-base font-semibold"
            >
              {isLoading ? "Signing in..." : "Sign In"}
            </Button>
          </div>
        )}

      </div>
    </div>
  )
}
