"use client"

import { useEffect, useState } from "react"
import { useRouter, useParams } from "next/navigation"
import { HomeownerDashboard } from "@/components/dashboards/homeowner-dashboard"
import { TechnicianDashboard } from "@/components/dashboards/technician-dashboard"
import { GuestDashboard } from "@/components/dashboards/guest-dashboard"

export default function DashboardPage() {
  const router = useRouter()
  const params = useParams()
  const [isLoading, setIsLoading] = useState(true)
  const [userRole, setUserRole] = useState<string | null>(null)

  useEffect(() => {
    // Check authentication
    const storedRole = localStorage.getItem("userRole")
    const token = localStorage.getItem("apiToken")

    if (!storedRole || storedRole !== params.role || !token) {
      router.push("/")
      return
    }

    setUserRole(storedRole)
    setIsLoading(false)
  }, [params.role, router])

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto" />
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    )
  }

  if (userRole === "homeowner") {
    return <HomeownerDashboard />
  }

  if (userRole === "technician") {
    return <TechnicianDashboard />
  }

  if (userRole === "guest") {
    return <GuestDashboard />
  }

  return null
}
