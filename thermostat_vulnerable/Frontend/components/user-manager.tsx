"use client"

import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Alert, AlertDescription } from "@/components/ui/alert"
import {
  listGuests,
  createGuest,
  deleteGuest as deleteGuestAPI,
  listTechnicianAccess,
  grantTechnicianAccess,
  revokeTechnicianAccess as revokeTechnicianAccessAPI,
  listTechnicians,
  type GuestDTO,
  type TechnicianAccessDTO,
  type TechnicianDTO,
} from "@/lib/api"
import { 
  Plus, 
  Trash2, 
  X, 
  UserCircle, 
  Wrench, 
  Clock, 
  CheckCircle, 
  XCircle,
  Save,
  Users,
  Shield,
  Key,
  Calendar
} from "lucide-react"

interface UserManagerProps {
  onClose: () => void
  userRole: "homeowner" | "technician"
}

export function UserManager({ onClose, userRole }: UserManagerProps) {
  const [guests, setGuests] = useState<GuestDTO[]>([])
  const [technicians, setTechnicians] = useState<TechnicianAccessDTO[]>([])
  const [availableTechnicians, setAvailableTechnicians] = useState<TechnicianDTO[]>([])
  const [isCreatingGuest, setIsCreatingGuest] = useState(false)
  const [isGrantingAccess, setIsGrantingAccess] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [newGuest, setNewGuest] = useState({ name: "", pin: "" })
  const [newTechnician, setNewTechnician] = useState({ technicianUsername: "", hours: 24 })

  useEffect(() => {
    loadUsers()
    if (userRole === "homeowner") {
      loadTechnicians()
    }
  }, [userRole])

  const loadUsers = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [guestsData, techAccessData] = await Promise.all([
        listGuests(),
        userRole === "homeowner" ? listTechnicianAccess() : Promise.resolve([]),
      ])
      setGuests(Array.isArray(guestsData) ? guestsData : [])
      if (userRole === "homeowner") {
        setTechnicians(Array.isArray(techAccessData) ? techAccessData : [])
      } else {
        setTechnicians([])
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load users")
      console.error("Failed to load users:", err)
      // Ensure arrays are always set even on error
      setGuests([])
      if (userRole === "homeowner") {
        setTechnicians([])
      }
    } finally {
      setIsLoading(false)
    }
  }

  const loadTechnicians = async () => {
    try {
      const techs = await listTechnicians()
      setAvailableTechnicians(techs)
    } catch (err) {
      console.error("Failed to load technicians:", err)
    }
  }

  const handleCreateGuest = async () => {
    if (!newGuest.name.trim()) return
    if (newGuest.pin.length > 99) {
      setError("PIN too long")
      return
    }

    setIsLoading(true)
    setError(null)
    try {
      await createGuest(newGuest.name, newGuest.pin)
      await loadUsers()
      setIsCreatingGuest(false)
      setNewGuest({ name: "", pin: "" })
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create guest")
      console.error("Failed to create guest:", err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleDeleteGuest = async (id: number) => {
    if (!confirm("Are you sure you want to delete this guest account?")) return

    setIsLoading(true)
    setError(null)
    try {
      await deleteGuestAPI(id)
      await loadUsers()
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete guest")
      console.error("Failed to delete guest:", err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleGrantAccess = async () => {
    if (!newTechnician.technicianUsername.trim()) {
      setError("Please select a technician")
      return
    }

    setIsLoading(true)
    setError(null)
    try {
      // Create dates in local time, then convert to ISO (UTC) for storage
      // The backend will store these as UTC, and we'll convert back to local for display
      const startTime = new Date()
      const endTime = new Date()
      endTime.setHours(endTime.getHours() + newTechnician.hours)

      // Convert to ISO string (UTC) for API
      await grantTechnicianAccess(
        newTechnician.technicianUsername,
        startTime.toISOString(),
        endTime.toISOString()
      )
      await loadUsers()
      setIsGrantingAccess(false)
      setNewTechnician({ technicianUsername: "", hours: 24 })
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to grant access")
      console.error("Failed to grant access:", err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleRevokeAccess = async (id: number) => {
    if (!confirm("Are you sure you want to revoke this technician's access?")) return

    setIsLoading(true)
    setError(null)
    try {
      await revokeTechnicianAccessAPI(id)
      await loadUsers()
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to revoke access")
      console.error("Failed to revoke access:", err)
    } finally {
      setIsLoading(false)
    }
  }

  const formatDate = (dateString: string) => {
    // Parse the date string - RFC3339 format from backend is in UTC
    // new Date() automatically parses UTC and converts to local timezone
    const date = new Date(dateString)
    // Verify the date is valid
    if (isNaN(date.getTime())) {
      return "Invalid date"
    }
    // toLocaleString automatically converts UTC to local timezone
    return date.toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "2-digit",
    })
  }

  const getAccessDuration = (hours: number) => {
    if (hours === 24) return "24 hours"
    if (hours === 48) return "2 days"
    if (hours === 72) return "3 days"
    if (hours === 168) return "1 week"
    return `${hours} hours`
  }

  const resetForms = () => {
    setIsCreatingGuest(false)
    setIsGrantingAccess(false)
    setNewGuest({ name: "", pin: "" })
    setNewTechnician({ name: "", hours: 24 })
  }

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-5xl max-h-[90vh] overflow-hidden border">
        {/* Header */}
        <div className="bg-gradient-to-r from-sky-500 to-blue-600 px-8 py-6 text-white">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center">
                <Users className="w-6 h-6" />
              </div>
              <div>
                <h2 className="text-2xl font-bold">User Management</h2>
                <p className="text-sky-100">Manage guest accounts and technician access</p>
              </div>
            </div>
            <Button 
              variant="ghost" 
              size="icon" 
              onClick={onClose}
              className="text-white hover:bg-white/20"
            >
              <X className="w-6 h-6" />
            </Button>
          </div>
        </div>

        <div className="p-8 overflow-y-auto max-h-[calc(90vh-120px)]">
          {/* Error Display */}
          {error && (
            <Alert className="mb-6 border-red-200 bg-red-50 dark:bg-red-950/20 dark:border-red-800">
              <AlertDescription className="text-red-800 dark:text-red-200">{error}</AlertDescription>
            </Alert>
          )}

          {/* Loading State */}
          {isLoading && (
            <div className="mb-6 text-center text-muted-foreground">Loading...</div>
          )}

          {/* Guest Accounts Section */}
          <div className="mb-8">
            <div className="flex items-center gap-3 mb-6">
              <div className="w-8 h-8 bg-sky-100 dark:bg-sky-900 rounded-lg flex items-center justify-center">
                <UserCircle className="w-5 h-5 text-sky-600 dark:text-sky-400" />
              </div>
              <h3 className="text-xl font-semibold text-foreground">Guest Accounts</h3>
              <Badge variant="secondary" className="bg-sky-100 text-sky-800 dark:bg-sky-900 dark:text-sky-200">
                {guests.length} accounts
              </Badge>
            </div>

            {/* Create Guest Form */}
            {isCreatingGuest ? (
              <Card className="p-6 mb-6 border-2 border-sky-200 bg-sky-50/50 dark:bg-sky-950/20 dark:border-sky-800">
                <div className="flex items-center gap-2 mb-4">
                  <div className="w-8 h-8 bg-sky-600 rounded-lg flex items-center justify-center">
                    <Plus className="w-4 h-4 text-white" />
                  </div>
                  <h4 className="text-lg font-semibold text-sky-900 dark:text-sky-100">Create Guest Account</h4>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="guest-name" className="text-sm font-medium">Guest Name</Label>
                    <Input
                      id="guest-name"
                      placeholder="Enter guest name"
                      value={newGuest.name}
                      onChange={(e) => setNewGuest({ ...newGuest, name: e.target.value })}
                      className="h-11"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="guest-pin" className="text-sm font-medium">PIN Code</Label>
                    <div className="relative">
                      <Key className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                      <Input
                        id="guest-pin"
                        type="password"
                        placeholder="4-digit PIN"
                        maxLength={4}
                        value={newGuest.pin}
                        onChange={(e) => setNewGuest({ ...newGuest, pin: e.target.value.replace(/\D/g, "") })}
                        className="h-11 pl-10"
                      />
                    </div>
                    <p className="text-xs text-muted-foreground">Enter a 4-digit PIN for guest access</p>
                  </div>
                </div>

                <div className="flex gap-3 mt-6">
                  <Button 
                    onClick={handleCreateGuest} 
                    className="flex items-center gap-2 bg-sky-600 hover:bg-sky-700"
                    disabled={isLoading || !newGuest.name.trim()}
                  >
                    <Save className="w-4 h-4" />
                    Create Guest Account
                  </Button>
                  <Button variant="outline" onClick={resetForms}>
                    Cancel
                  </Button>
                </div>
              </Card>
            ) : (
              <div className="mb-6">
                <Button 
                  onClick={() => setIsCreatingGuest(true)} 
                  className="w-full h-12 bg-sky-600 hover:bg-sky-700 text-white font-semibold"
                >
                  <Plus className="w-5 h-5 mr-2" />
                  Add Guest Account
                </Button>
              </div>
            )}

            {/* Guest List */}
            {guests.length === 0 ? (
              <Card className="p-8 text-center border-dashed border-2 border-border">
                <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
                  <UserCircle className="w-8 h-8 text-muted-foreground" />
                </div>
                <h4 className="text-lg font-medium text-foreground mb-2">No guest accounts</h4>
                <p className="text-muted-foreground mb-4">Create guest accounts to allow limited access to your thermostat</p>
                <Button onClick={() => setIsCreatingGuest(true)} className="bg-sky-600 hover:bg-sky-700">
                  <Plus className="w-4 h-4 mr-2" />
                  Create Guest Account
                </Button>
              </Card>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {guests.map((guest) => (
                  <Card key={guest.id} className="p-6 hover:shadow-lg transition-shadow border">
                    <div className="flex items-start justify-between mb-4">
                      <div className="flex items-center gap-3">
                        <div className="w-12 h-12 bg-gradient-to-br from-sky-500 to-blue-600 rounded-xl flex items-center justify-center">
                          <UserCircle className="w-6 h-6 text-white" />
                        </div>
                        <div>
                          <h4 className="font-semibold text-lg text-foreground">{guest.username}</h4>
                          <p className="text-sm text-muted-foreground">
                            Created {new Date(guest.created_at).toLocaleDateString()}
                          </p>
                        </div>
                      </div>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDeleteGuest(guest.id)}
                        disabled={isLoading}
                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>

                    <div className="flex items-center gap-2">
                      <Key className="w-4 h-4 text-muted-foreground" />
                      <span className="text-sm text-muted-foreground">PIN: ••••</span>
                      <Badge variant="secondary" className="bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                        Active
                      </Badge>
                    </div>
                  </Card>
                ))}
              </div>
            )}
          </div>

          {/* Technician Access Section (Homeowner only) */}
          {userRole === "homeowner" && (
            <div>
              <div className="flex items-center gap-3 mb-6">
                <div className="w-8 h-8 bg-orange-100 dark:bg-orange-900 rounded-lg flex items-center justify-center">
                  <Wrench className="w-5 h-5 text-orange-600 dark:text-orange-400" />
                </div>
                <h3 className="text-xl font-semibold text-foreground">Technician Access</h3>
                <Badge variant="secondary" className="bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200">
                  {technicians?.length || 0} access grants
                </Badge>
              </div>

              {/* Grant Access Form */}
              {isGrantingAccess ? (
                <Card className="p-6 mb-6 border-2 border-orange-200 bg-orange-50/50 dark:bg-orange-950/20 dark:border-orange-800">
                  <div className="flex items-center gap-2 mb-4">
                    <div className="w-8 h-8 bg-orange-600 rounded-lg flex items-center justify-center">
                      <Shield className="w-4 h-4 text-white" />
                    </div>
                    <h4 className="text-lg font-semibold text-orange-900 dark:text-orange-100">Grant Technician Access</h4>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="tech-name" className="text-sm font-medium">Technician</Label>
                      <Select
                        value={newTechnician.technicianUsername}
                        onValueChange={(value) => setNewTechnician({ ...newTechnician, technicianUsername: value })}
                      >
                        <SelectTrigger className="h-11">
                          <SelectValue placeholder="Select a technician" />
                        </SelectTrigger>
                        <SelectContent>
                          {availableTechnicians.map((tech) => (
                            <SelectItem key={tech.id} value={tech.username}>
                              {tech.username}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="space-y-2">
                      <Label className="text-sm font-medium">Access Duration</Label>
                      <div className="grid grid-cols-2 gap-2">
                        {[24, 48, 72, 168].map((hours) => (
                          <Button
                            key={hours}
                            variant={newTechnician.hours === hours ? "default" : "outline"}
                            className={`h-10 ${
                              newTechnician.hours === hours 
                                ? "bg-orange-600 hover:bg-orange-700" 
                                : "hover:bg-muted"
                            }`}
                            onClick={() => setNewTechnician({ ...newTechnician, hours })}
                          >
                            {getAccessDuration(hours)}
                          </Button>
                        ))}
                      </div>
                    </div>
                  </div>

                  <div className="flex gap-3 mt-6">
                    <Button 
                      onClick={handleGrantAccess} 
                      className="flex items-center gap-2 bg-orange-600 hover:bg-orange-700"
                      disabled={isLoading || !newTechnician.technicianUsername.trim()}
                    >
                      <Save className="w-4 h-4" />
                      Grant Access
                    </Button>
                    <Button variant="outline" onClick={resetForms}>
                      Cancel
                    </Button>
                  </div>
                </Card>
              ) : (
                <div className="mb-6">
                  <Button 
                    onClick={() => setIsGrantingAccess(true)} 
                    className="w-full h-12 bg-orange-600 hover:bg-orange-700 text-white font-semibold"
                  >
                    <Shield className="w-5 h-5 mr-2" />
                    Grant Technician Access
                  </Button>
                </div>
              )}

              {/* Technician List */}
              {!technicians || technicians.length === 0 ? (
                <Card className="p-8 text-center border-dashed border-2 border-border">
                  <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
                    <Wrench className="w-8 h-8 text-muted-foreground" />
                  </div>
                  <h4 className="text-lg font-medium text-foreground mb-2">No technician access granted</h4>
                  <p className="text-muted-foreground mb-4">Grant temporary access to technicians for system maintenance</p>
                  <Button onClick={() => setIsGrantingAccess(true)} className="bg-orange-600 hover:bg-orange-700">
                    <Shield className="w-4 h-4 mr-2" />
                    Grant Access
                  </Button>
                </Card>
              ) : (
                <div className="grid grid-cols-1 gap-4">
                  {(technicians || []).map((tech) => (
                    <Card key={tech.id} className="p-6 hover:shadow-lg transition-shadow border">
                      <div className="flex items-start justify-between mb-4">
                        <div className="flex items-center gap-3">
                          <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${
                            tech.is_active 
                              ? "bg-gradient-to-br from-orange-500 to-orange-600" 
                              : "bg-gradient-to-br from-muted to-muted-foreground"
                          }`}>
                            <Wrench className="w-6 h-6 text-white" />
                          </div>
                          <div>
                            <div className="flex items-center gap-2">
                              <h4 className="font-semibold text-lg text-foreground">{tech.technician_name}</h4>
                              {tech.is_active ? (
                                <CheckCircle className="w-5 h-5 text-green-500" />
                              ) : (
                                <XCircle className="w-5 h-5 text-muted-foreground" />
                              )}
                            </div>
                            <p className="text-sm text-muted-foreground">
                              Access granted • {new Date(tech.start_time).toLocaleDateString()}
                            </p>
                          </div>
                        </div>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleRevokeAccess(tech.id)}
                          disabled={isLoading}
                          className="h-8 w-8 text-muted-foreground hover:text-destructive"
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>

                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-4">
                          <div className="flex items-center gap-2">
                            <Calendar className="w-4 h-4 text-muted-foreground" />
                            <span className="text-sm text-muted-foreground">
                              {tech.is_active ? `Expires: ${formatDate(tech.end_time)}` : "Expired"}
                            </span>
                          </div>
                          <Badge className={`${
                            tech.is_active 
                              ? "bg-green-100 text-green-800 border-green-200 dark:bg-green-900 dark:text-green-200 dark:border-green-800" 
                              : "bg-muted text-muted-foreground border-border"
                          } border`}>
                            {tech.is_active ? "Active" : "Expired"}
                          </Badge>
                        </div>
                      </div>
                    </Card>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}