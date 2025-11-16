"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  X,
  Plus,
  Clock,
  Calendar,
  Sun,
  Moon,
  Home,
  Briefcase,
  Trash2,
  Edit3,
  Thermometer,
} from "lucide-react";
import {
  listSchedules,
  createSchedule as apiCreateSchedule,
  deleteSchedule as apiDeleteSchedule,
} from "@/lib/api";

interface ScheduleItem {
  id: string;
  name: string;
  time: string;
  temperature: number;
  isActive: boolean;
}

interface ScheduleManagerProps {
  onClose: () => void;
}

export function ScheduleManager({ onClose }: ScheduleManagerProps) {
  const [schedules, setSchedules] = useState<ScheduleItem[]>([]);
  const [isCreating, setIsCreating] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [newSchedule, setNewSchedule] = useState({
    name: "",
    time: "07:00",
    temperature: 72,
  });

  useEffect(() => {
    void loadSchedules();
  }, []);

  async function loadSchedules() {
    const token =
      typeof window !== "undefined" ? localStorage.getItem("apiToken") : null;
    if (!token) {
      console.error("Not authenticated");
      setSchedules([]);
      return;
    }
    try {
      const rows = await listSchedules();
      // Map server DTOs to a simple UI schedule representation
      const mapped: ScheduleItem[] = (Array.isArray(rows) ? rows : []).map(
        (r) => ({
          id: String(r.id),
          name: r.name || `Schedule #${r.id}`,
          time: r.start_time,
          temperature: r.target_temp,
          isActive: true,
        }),
      );
      setSchedules(mapped);
    } catch (e) {
      console.error("Failed to load schedules:", e);
      setSchedules([]);
    }
  }

  const getScheduleIcon = (name: string | undefined) => {
    if (!name) return <Clock className="w-4 h-4 text-sky-500" />;
    const lowerName = name.toLowerCase();
    if (lowerName.includes("wake") || lowerName.includes("morning"))
      return <Sun className="w-4 h-4 text-yellow-500" />;
    if (lowerName.includes("sleep") || lowerName.includes("night"))
      return <Moon className="w-4 h-4 text-indigo-500" />;
    if (lowerName.includes("leave") || lowerName.includes("away"))
      return <Briefcase className="w-4 h-4 text-gray-500" />;
    if (lowerName.includes("return") || lowerName.includes("home"))
      return <Home className="w-4 h-4 text-green-500" />;
    return <Clock className="w-4 h-4 text-sky-500" />;
  };

  const handleCreateSchedule = async () => {
    if (!newSchedule.name.trim()) return;

    const token =
      typeof window !== "undefined" ? localStorage.getItem("apiToken") : null;
    if (!token) {
      console.error("Not authenticated");
      return;
    }

    try {
      await apiCreateSchedule({
        name: newSchedule.name,
        start_time: newSchedule.time,
        target_temp: newSchedule.temperature,
      });
      await loadSchedules();
      setIsCreating(false);
      setNewSchedule({ name: "", time: "07:00", temperature: 72 });
    } catch (e) {
      console.error("Failed to create schedule:", e);
    }
  };

  const handleEditSchedule = (schedule: ScheduleItem) => {
    setEditingId(schedule.id);
    setNewSchedule({
      name: schedule.name,
      time: schedule.time,
      temperature: schedule.temperature,
    });
  };

  const handleUpdateSchedule = async () => {
    if (!newSchedule.name.trim() || !editingId) return;
    // No update endpoint; leave as-is (future: implement server-side update)
    setEditingId(null);
    await loadSchedules();
    setNewSchedule({ name: "", time: "07:00", temperature: 72 });
  };

  const handleDeleteSchedule = async (id: string) => {
    if (confirm("Are you sure you want to delete this schedule?")) {
      const token =
        typeof window !== "undefined" ? localStorage.getItem("apiToken") : null;
      if (!token) {
        console.error("Not authenticated");
        return;
      }
      setSchedules((prev) => prev.filter((schedule) => schedule.id !== id));
      try {
        await apiDeleteSchedule(Number(id));
        await loadSchedules();
      } catch (e) {
        console.error("Failed to delete schedule:", e);
      }
    }
  };

  const toggleScheduleActive = (id: string) => {
    setSchedules(
      schedules.map((schedule) =>
        schedule.id === id
          ? { ...schedule, isActive: !schedule.isActive }
          : schedule,
      ),
    );
  };

  const resetForm = () => {
    setIsCreating(false);
    setEditingId(null);
    setNewSchedule({ name: "", time: "07:00", temperature: 72 });
  };

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden border">
        {/* Header */}
        <div className="bg-gradient-to-r from-sky-500 to-blue-600 px-8 py-6 text-white">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center">
                <Calendar className="w-6 h-6" />
              </div>
              <div>
                <h2 className="text-2xl font-bold">Schedule Management</h2>
                <p className="text-sky-100">
                  Set up automatic temperature schedules
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
                  {editingId ? "Edit Schedule" : "Create New Schedule"}
                </h3>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Schedule Name */}
                <div className="space-y-2">
                  <Label
                    htmlFor="schedule-name"
                    className="text-sm font-medium"
                  >
                    Schedule Name
                  </Label>
                  <Input
                    id="schedule-name"
                    placeholder="e.g., Wake Up, Sleep, Leave Home"
                    value={newSchedule.name}
                    onChange={(e) =>
                      setNewSchedule({ ...newSchedule, name: e.target.value })
                    }
                    className="h-11"
                  />
                </div>

                {/* Time */}
                <div className="space-y-2">
                  <Label
                    htmlFor="schedule-time"
                    className="text-sm font-medium"
                  >
                    Time
                  </Label>
                  <Input
                    id="schedule-time"
                    type="time"
                    value={newSchedule.time}
                    onChange={(e) =>
                      setNewSchedule({ ...newSchedule, time: e.target.value })
                    }
                    className="h-11"
                  />
                </div>

                {/* Temperature */}
                <div className="space-y-2">
                  <Label className="text-sm font-medium">
                    Target Temperature
                  </Label>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() =>
                        setNewSchedule({
                          ...newSchedule,
                          temperature: Math.max(
                            60,
                            newSchedule.temperature - 1,
                          ),
                        })
                      }
                      className="h-11 w-11"
                    >
                      −
                    </Button>
                    <div className="flex items-center gap-2 bg-background rounded-lg border px-4 py-2 min-w-[100px] justify-center">
                      <Thermometer className="w-4 h-4 text-muted-foreground" />
                      <span className="text-xl font-bold">
                        {newSchedule.temperature}°
                      </span>
                    </div>
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() =>
                        setNewSchedule({
                          ...newSchedule,
                          temperature: Math.min(
                            85,
                            newSchedule.temperature + 1,
                          ),
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
                    editingId ? handleUpdateSchedule : handleCreateSchedule
                  }
                  className="flex items-center gap-2 bg-sky-600 hover:bg-sky-700"
                  disabled={!newSchedule.name.trim()}
                >
                  {editingId ? "Update Schedule" : "Create Schedule"}
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
                Create New Schedule
              </Button>
            </div>
          )}

          {/* Schedules List */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-foreground flex items-center gap-2">
              <div className="w-6 h-6 bg-muted rounded flex items-center justify-center">
                <span className="text-xs font-bold text-muted-foreground">
                  {schedules.length}
                </span>
              </div>
              Scheduled Events
            </h3>
            {schedules.length === 0 ? (
              <Card className="p-8 text-center border-dashed border-2 border-border">
                <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
                  <Calendar className="w-8 h-8 text-muted-foreground" />
                </div>
                <h4 className="text-lg font-medium text-foreground mb-2">
                  No schedules created yet
                </h4>
                <p className="text-muted-foreground mb-4">
                  Create your first schedule to automate temperature changes
                </p>
                <Button
                  onClick={() => setIsCreating(true)}
                  className="bg-sky-600 hover:bg-sky-700"
                >
                  <Plus className="w-4 h-4 mr-2" />
                  Create Schedule
                </Button>
              </Card>
            ) : (
              <div className="grid grid-cols-1 gap-4">
                {schedules.map((schedule) => (
                  <Card
                    key={schedule.id}
                    className="p-6 hover:shadow-lg transition-shadow border"
                  >
                    <div className="flex items-start justify-between mb-4">
                      <div className="flex items-center gap-3">
                        <div className="w-12 h-12 bg-gradient-to-br from-sky-500 to-blue-600 rounded-xl flex items-center justify-center">
                          {getScheduleIcon(schedule.name)}
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <h4 className="font-semibold text-lg text-foreground">
                              {schedule.name}
                            </h4>
                            <Badge
                              variant={
                                schedule.isActive ? "default" : "secondary"
                              }
                              className={
                                schedule.isActive
                                  ? "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
                                  : ""
                              }
                            >
                              {schedule.isActive ? "Active" : "Inactive"}
                            </Badge>
                          </div>
                          <p className="text-sm text-muted-foreground">
                            {schedule.time}
                          </p>
                        </div>
                      </div>
                      <div className="flex gap-2">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => setEditingId(schedule.id)}
                          className="h-8 w-8 text-muted-foreground hover:text-sky-600"
                        >
                          <Edit3 className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => void handleDeleteSchedule(schedule.id)}
                          className="h-8 w-8 text-muted-foreground hover:text-destructive"
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </div>

                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-4">
                        <div className="flex items-center gap-2">
                          <Thermometer className="w-4 h-4 text-muted-foreground" />
                          <span className="text-2xl font-bold text-foreground">
                            {schedule.temperature}°
                          </span>
                        </div>
                      </div>
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
