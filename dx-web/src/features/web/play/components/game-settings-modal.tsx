"use client";

import { Volume2, Mic, X } from "lucide-react";
import { useGameStore } from "@/features/web/play/hooks/use-game-store";
import { useGameSettings } from "@/features/web/play/hooks/use-game-settings";

/** Game settings modal with sound and pronunciation toggles */
export function GameSettingsModal() {
  const closeOverlay = useGameStore((s) => s.closeOverlay);
  const typingSoundEnabled = useGameSettings((s) => s.typingSoundEnabled);
  const autoPlayPronunciation = useGameSettings((s) => s.autoPlayPronunciation);
  const toggleTypingSound = useGameSettings((s) => s.toggleTypingSound);
  const toggleAutoPlayPronunciation = useGameSettings(
    (s) => s.toggleAutoPlayPronunciation
  );

  const settingsRows = [
    {
      icon: Volume2,
      label: "打字音效",
      enabled: typingSoundEnabled,
      onToggle: toggleTypingSound,
    },
    {
      icon: Mic,
      label: "自动播放发音",
      enabled: autoPlayPronunciation,
      onToggle: toggleAutoPlayPronunciation,
    },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 px-4">
      <div className="flex w-full max-w-[420px] flex-col gap-6 rounded-[20px] bg-card p-6 md:p-8">
        {/* Header */}
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-bold text-foreground">设置</h2>
          <button
            type="button"
            aria-label="关闭"
            onClick={closeOverlay}
            className="flex h-8 w-8 items-center justify-center rounded-lg bg-muted"
          >
            <X className="h-4 w-4 text-muted-foreground" />
          </button>
        </div>

        <div className="h-px w-full bg-border" />

        {/* Settings rows */}
        <div className="flex flex-col gap-5">
          {settingsRows.map((row) => (
            <div
              key={row.label}
              className="flex items-center justify-between"
            >
              <div className="flex items-center gap-2.5">
                <row.icon className="h-[18px] w-[18px] text-muted-foreground" />
                <span className="text-[15px] font-medium text-foreground">
                  {row.label}
                </span>
              </div>
              <button
                type="button"
                role="switch"
                aria-checked={row.enabled}
                aria-label={row.label}
                onClick={row.onToggle}
                className={`flex h-6 w-11 items-center rounded-xl p-0.5 transition-colors ${
                  row.enabled
                    ? "justify-end bg-teal-600"
                    : "justify-start bg-border"
                }`}
              >
                <div className="h-5 w-5 rounded-full bg-white" />
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
