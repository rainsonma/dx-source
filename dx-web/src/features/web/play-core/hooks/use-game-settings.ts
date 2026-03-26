import { create } from "zustand";
import { persist } from "zustand/middleware";

interface GameSettingsState {
  typingSoundEnabled: boolean;
  autoPlayPronunciation: boolean;
  toggleTypingSound: () => void;
  toggleAutoPlayPronunciation: () => void;
}

/** Persisted game settings for sound and pronunciation preferences */
export const useGameSettings = create<GameSettingsState>()(
  persist(
    (set) => ({
      typingSoundEnabled: true,
      autoPlayPronunciation: true,
      toggleTypingSound: () =>
        set((s) => ({ typingSoundEnabled: !s.typingSoundEnabled })),
      toggleAutoPlayPronunciation: () =>
        set((s) => ({ autoPlayPronunciation: !s.autoPlayPronunciation })),
    }),
    { name: "dx-game-settings" }
  )
);
