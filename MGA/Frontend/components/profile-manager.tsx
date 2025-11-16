"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { ThermostatProfile } from "@/lib/types";
import {
  listProfiles,
  createProfile as apiCreateProfile,
  deleteProfile as apiDeleteProfile,
} from "@/lib/api";
import { Plus, Trash2, X, Edit3, Thermometer, Settings } from "lucide-react";

interface ProfileManagerProps {
  onClose: () => void;
  onProfileChange?: () => void;
}

export function ProfileManager({ onClose, onProfileChange }: ProfileManagerProps) {
  const [profiles, setProfiles] = useState<ThermostatProfile[]>([]);
  const [isCreating, setIsCreating] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [newProfile, setNewProfile] = useState({
    name: "",
    targetTemp: 72,
  });

  useEffect(() => {
    loadProfiles();
  }, []);

  const loadProfiles = async () => {
    const token =
      typeof window !== "undefined" ? localStorage.getItem("apiToken") : null;
    if (!token) {
      console.error("Not authenticated");
      setProfiles([]);
      return;
    }
    try {
      const rows = await listProfiles();
      const mapped = (Array.isArray(rows) ? rows : []).map(
        (r): ThermostatProfile => ({
          id: String(r.id),
          name: r.name,
          targetTemp: Math.round(r.target_temp),
          createdAt: r.created_at,
        }),
      );
      setProfiles(mapped);
    } catch (e) {
      console.error("Failed to load profiles:", e);
      setProfiles([]);
    }
  };

  const handleCreateProfile = async () => {
    if (!newProfile.name.trim()) return;

    const token =
      typeof window !== "undefined" ? localStorage.getItem("apiToken") : null;
    if (!token) {
      console.error("Not authenticated");
      return;
    }

    try {
      await apiCreateProfile({
        name: newProfile.name,
        target_temp: newProfile.targetTemp,
      });
      await loadProfiles();
      setIsCreating(false);
      setNewProfile({ name: "", targetTemp: 72 });
      // Notify parent that profiles have changed
      if (onProfileChange) {
        onProfileChange();
      }
    } catch (e) {
      console.error("Failed to create profile:", e);
    }
  };

  const handleEditProfile = (profile: ThermostatProfile) => {
    setEditingId(profile.id);
    setNewProfile({
      name: profile.name,
      targetTemp: profile.targetTemp,
    });
  };

  const handleUpdateProfile = async () => {
    if (!newProfile.name.trim() || !editingId) return;

    const token =
      typeof window !== "undefined" ? localStorage.getItem("apiToken") : null;
    if (!token) {
      console.error("Not authenticated");
      return;
    }

    // No update endpoint; emulate by delete+create
    try {
      const current = profiles.find((p) => p.id === editingId);
      if (current) {
        await apiDeleteProfile(Number(editingId));
        await apiCreateProfile({
          name: newProfile.name,
          target_temp: newProfile.targetTemp,
        });
      }
      await loadProfiles();
      setEditingId(null);
      setNewProfile({ name: "", targetTemp: 72 });
      // Notify parent that profiles have changed
      if (onProfileChange) {
        onProfileChange();
      }
    } catch (e) {
      console.error("Failed to update profile:", e);
    }
  };

  const handleDeleteProfile = async (id: string) => {
    if (confirm("Are you sure you want to delete this profile?")) {
      const token =
        typeof window !== "undefined" ? localStorage.getItem("apiToken") : null;
      if (!token) {
        console.error("Not authenticated");
        return;
      }
      const numericId = Number.parseInt(id, 10);
      if (!Number.isFinite(numericId) || numericId <= 0) {
        console.error("Invalid profile id", id);
        return;
      }
      setProfiles((prev) => prev.filter((p) => p.id !== id));
      try {
        await apiDeleteProfile(numericId);
        await loadProfiles();
        // Notify parent that profiles have changed
        if (onProfileChange) {
          onProfileChange();
        }
      } catch (e) {
        console.error("Failed to delete profile:", e);
      }
    }
  };

  const resetForm = () => {
    setIsCreating(false);
    setEditingId(null);
    setNewProfile({ name: "", targetTemp: 72 });
  };

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden border">
        {/* Header */}
        <div className="bg-gradient-to-r from-sky-500 to-blue-600 px-8 py-6 text-white">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center">
                <Settings className="w-6 h-6" />
              </div>
              <div>
                <h2 className="text-2xl font-bold">Temperature Profiles</h2>
                <p className="text-sky-100">
                  Manage your custom temperature settings
                </p>
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
          {/* Create/Edit Form */}
          {(isCreating || editingId) && (
            <Card className="p-6 mb-6 border-2 border-sky-200 bg-sky-50/50 dark:bg-sky-950/20 dark:border-sky-800">
              <div className="flex items-center gap-2 mb-4">
                <div className="w-8 h-8 bg-sky-600 rounded-lg flex items-center justify-center">
                  {editingId ? (
                    <Edit3 className="w-4 h-4 text-white" />
                  ) : (
                    <Plus className="w-4 h-4 text-white" />
                  )}
                </div>
                <h3 className="text-lg font-semibold text-sky-900 dark:text-sky-100">
                  {editingId ? "Edit Profile" : "Create New Profile"}
                </h3>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {/* Profile Name */}
                <div className="space-y-2">
                  <Label htmlFor="profile-name" className="text-sm font-medium">
                    Profile Name
                  </Label>
                  <Input
                    id="profile-name"
                    placeholder="e.g., Vacation, Night, Work"
                    value={newProfile.name}
                    onChange={(e) =>
                      setNewProfile({ ...newProfile, name: e.target.value })
                    }
                    className="h-11"
                  />
                </div>

                {/* Target Temperature */}
                <div className="space-y-2">
                  <Label className="text-sm font-medium">
                    Target Temperature
                  </Label>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() =>
                        setNewProfile({
                          ...newProfile,
                          targetTemp: Math.max(60, newProfile.targetTemp - 1),
                        })
                      }
                      className="h-11 w-11"
                    >
                      −
                    </Button>
                    <div className="flex items-center gap-2 bg-white rounded-lg border px-4 py-2 min-w-[100px] justify-center">
                      <Thermometer className="w-4 h-4 text-gray-500" />
                      <span className="text-xl font-bold">
                        {newProfile.targetTemp}°
                      </span>
                    </div>
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() =>
                        setNewProfile({
                          ...newProfile,
                          targetTemp: Math.min(85, newProfile.targetTemp + 1),
                        })
                      }
                      className="h-11 w-11"
                    >
                      +
                    </Button>
                  </div>
                </div>
              </div>

              <div className="flex gap-3 mt-6">
                <Button
                  onClick={
                    editingId ? handleUpdateProfile : handleCreateProfile
                  }
                  className="flex items-center gap-2 bg-sky-600 hover:bg-sky-700"
                  disabled={!newProfile.name.trim()}
                >
                  {editingId ? "Update Profile" : "Create Profile"}
                </Button>
                <Button variant="outline" onClick={resetForm}>
                  Cancel
                </Button>
              </div>
            </Card>
          )}

          {/* Create Button */}
          {!isCreating && !editingId && (
            <div className="mb-6">
              <Button
                onClick={() => setIsCreating(true)}
                className="w-full h-14 bg-sky-600 hover:bg-sky-700 text-white font-semibold"
              >
                <Plus className="w-5 h-5 mr-2" />
                Create New Profile
              </Button>
            </div>
          )}

          {/* Profiles List */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-foreground flex items-center gap-2">
              <div className="w-6 h-6 bg-muted rounded flex items-center justify-center">
                <span className="text-xs font-bold text-muted-foreground">
                  {profiles.length}
                </span>
              </div>
              Existing Profiles
            </h3>

            {profiles.length === 0 ? (
              <Card className="p-8 text-center border-dashed border-2 border-border">
                <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
                  <Settings className="w-8 h-8 text-muted-foreground" />
                </div>
                <h4 className="text-lg font-medium text-foreground mb-2">
                  No profiles created yet
                </h4>
                <p className="text-muted-foreground mb-4">
                  Create your first temperature profile to get started
                </p>
                <Button
                  onClick={() => setIsCreating(true)}
                  className="bg-sky-600 hover:bg-sky-700"
                >
                  <Plus className="w-4 h-4 mr-2" />
                  Create Profile
                </Button>
              </Card>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {profiles.map((profile) => (
                  <Card
                    key={profile.id}
                    className="p-6 hover:shadow-lg transition-shadow border"
                  >
                    <div className="flex items-start justify-between mb-4">
                      <div>
                        <h4 className="font-semibold text-lg text-foreground">
                          {profile.name}
                        </h4>
                        <p className="text-sm text-muted-foreground">
                          Created{" "}
                          {new Date(profile.createdAt).toLocaleDateString()}
                        </p>
                      </div>
                      <div className="flex gap-2">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleEditProfile(profile)}
                          className="h-8 w-8 text-muted-foreground hover:text-sky-600"
                        >
                          <Edit3 className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDeleteProfile(profile.id)}
                          className="h-8 w-8 text-muted-foreground hover:text-destructive"
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </div>

                    <div className="flex items-center gap-2 mt-2">
                      <Thermometer className="w-4 h-4 text-muted-foreground" />
                      <span className="text-2xl font-bold text-foreground">
                        {profile.targetTemp}°
                      </span>
                    </div>
                  </Card>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
