import { create } from "zustand";

type GameSearchTextStore = {
  q: string;
  setQ: (q: string) => void;
  clearQ: () => void;
};

export const useGameSearchText = create<GameSearchTextStore>((set) => ({
  q: "",
  setQ: (q) => set({ q }),
  clearQ: () => set({ q: "" }),
}));
