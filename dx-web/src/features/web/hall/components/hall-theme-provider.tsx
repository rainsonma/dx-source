"use client";

import { ThemeProvider } from "next-themes";

/** Scoped theme provider for all hall pages */
export function HallThemeProvider({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="light" storageKey="hall-theme">
      {children}
    </ThemeProvider>
  );
}
