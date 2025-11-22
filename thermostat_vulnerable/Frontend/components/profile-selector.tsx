"use client";

import { useState, useEffect } from "react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import type { ThermostatProfile } from "@/lib/types";
import { listProfiles } from "@/lib/api";
import { Check, Thermometer } from "lucide-react";

interface ProfileSelectorProps {
  onSelectProfile: (profile: ThermostatProfile) => void;
  currentProfileId?: string;
  refreshTrigger?: number;
  onManage?: () => void;
}

export function ProfileSelector({
  onSelectProfile,
  currentProfileId,
  refreshTrigger,
  onManage,
}: ProfileSelectorProps) {
  const [profiles, setProfiles] = useState<ThermostatProfile[]>([]);

  useEffect(() => {
    const loadProfiles = async () => {
      const token =
        typeof window !== "undefined" ? localStorage.getItem("apiToken") : null;
      if (!token) return;

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
      } catch (error) {
        console.error("Failed to load profiles:", error);
        setProfiles([]);
      }
    };
    loadProfiles();
  }, [refreshTrigger]);

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="font-semibold text-sm text-muted-foreground uppercase tracking-wide">
          Select Profile
        </h3>
        {onManage && (
          <Button
            variant="outline"
            size="sm"
            onClick={onManage}
          >
            Manage
          </Button>
        )}
      </div>
      <div className="grid grid-cols-2 gap-3">
        {profiles.map((profile) => {
          const isSelected = currentProfileId === profile.id;

          return (
            <Card
              key={profile.id}
              className={`p-4 cursor-pointer transition-all border-2 ${
                isSelected
                  ? "border-primary bg-primary/5"
                  : "border-border hover:border-primary/50"
              }`}
              onClick={() => onSelectProfile(profile)}
            >
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <h4 className="font-semibold">{profile.name}</h4>
                  {isSelected && <Check className="w-4 h-4 text-primary" />}
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-2xl font-bold flex items-center gap-1">
                    <Thermometer className="w-4 h-4 text-muted-foreground" />
                    {profile.targetTemp}°
                  </span>
                </div>
              </div>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
