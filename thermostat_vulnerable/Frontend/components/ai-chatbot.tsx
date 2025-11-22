"use client";

import { useState } from "react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

type Message = {
  role: "assistant" | "user";
  content: string;
};

type AIChatbotProps = {
  outdoorTemp?: number;
};

type PromptConfig = {
  label: string;
  keywords: string[];
  respond: (params: { prompt: string; outdoorTemp?: number }) => string;
};

const PROMPTS: PromptConfig[] = [
  {
    label: "What should I wear today?",
    keywords: ["wear", "clothes", "outfit", "weather"],
    respond: ({ outdoorTemp }) => {
      const temp = typeof outdoorTemp === "number" ? outdoorTemp : 60;
      if (temp <= 40) {
        return "It's pretty chilly outside. Consider a heavy coat, scarf, gloves, and warm boots.";
      }
      if (temp <= 60) {
        return "A light jacket or sweater should do the trick. Layers will keep you comfortable.";
      }
      if (temp <= 80) {
        return "Mild weather today. Comfortable jeans or shorts with a breathable top would work well.";
      }
      return "It's hot out there! Light fabrics, shorts, and breathable shirts will help you stay cool.";
    },
  },
  {
    label: "How is the HVAC system performing?",
    keywords: ["hvac", "system", "status", "performance"],
    respond: () =>
      "All monitored systems report normal operation. If you notice anything unusual, try running diagnostics or contact support.",
  },
  {
    label: "Any energy saving tips?",
    keywords: ["energy", "saving", "tips"],
    respond: () =>
      "Try setting schedules for when you're away, keep vents clear, and avoid drastic temperature swings to reduce energy use.",
  },
  {
    label: "Should I schedule maintenance?",
    keywords: ["maintenance", "service", "checkup"],
    respond: () =>
      "If it's been more than a year since your last tune-up or you hear unusual noises, it's a good time to schedule maintenance.",
  },
  {
    label: "What is the humidity level like?",
    keywords: ["humidity", "humid"],
    respond: () =>
      "Indoor humidity should stay between 30-50% for comfort. Use a dehumidifier or humidifier if you notice discomfort.",
  },
];

export function AIChatbot({ outdoorTemp }: AIChatbotProps) {
  const [messages, setMessages] = useState<Message[]>([
    {
      role: "assistant",
      content: "Hi! I'm your quick-assist bot. Choose a question below or type your own.",
    },
  ]);
  const [input, setInput] = useState("");

  const findResponse = (prompt: string): string => {
    const normalized = prompt.trim().toLowerCase();
    const match = PROMPTS.find((preset) =>
      preset.keywords.some((kw) => normalized.includes(kw))
    );
    if (!match) {
      return "I don't have a quick tip for that yet. Try one of the prompt buttons below.";
    }
    return match.respond({ prompt: normalized, outdoorTemp });
  };

  const handleAsk = (question: string) => {
    const trimmed = question.trim();
    if (!trimmed) return;
    setMessages((prev) => [
      ...prev,
      { role: "user", content: trimmed },
      { role: "assistant", content: findResponse(trimmed) },
    ]);
  };

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    handleAsk(input);
    setInput("");
  };

  return (
    <Card className="p-4 space-y-4">
      <div>
        <h3 className="font-semibold text-foreground">AI Chatbot</h3>
        <p className="text-sm text-muted-foreground">
          Quick suggestions based on preset questions—no external services required.
        </p>
      </div>

      <div className="space-y-2 max-h-64 overflow-y-auto rounded-lg border border-border/50 bg-muted/30 p-3">
        {messages.map((message, index) => (
          <div
            key={index}
            className={`text-sm leading-relaxed ${
              message.role === "assistant" ? "text-muted-foreground" : "text-foreground font-medium"
            }`}
          >
            {message.role === "assistant" ? "Assistant: " : "You: "}
            {message.content}
          </div>
        ))}
      </div>

      <div className="flex flex-wrap gap-2">
        {PROMPTS.map((prompt) => (
          <Button
            key={prompt.label}
            variant="secondary"
            size="sm"
            onClick={() => handleAsk(prompt.label)}
          >
            {prompt.label}
          </Button>
        ))}
      </div>

      <form onSubmit={handleSubmit} className="flex gap-2">
        <Input
          placeholder="Ask a quick question..."
          value={input}
          onChange={(event) => setInput(event.target.value)}
        />
        <Button type="submit">Ask</Button>
      </form>
    </Card>
  );
}

