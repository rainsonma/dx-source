"use client";

import { useEffect, useState } from "react";
import { LandingHeader } from "@/components/in/landing-header";

interface StickyHeaderProps {
  isLoggedIn?: boolean;
  /** Start with a transparent background and transition to solid on scroll. */
  transparent?: boolean;
}

export function StickyHeader({
  isLoggedIn = false,
  transparent = false,
}: StickyHeaderProps) {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    if (!transparent) return;
    const onScroll = () => setScrolled(window.scrollY > 40);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, [transparent]);

  const showBackground = !transparent || scrolled;

  return (
    <div
      className={`sticky top-0 z-50 w-full transition-[background-color,border-color,backdrop-filter] duration-300 ${
        showBackground
          ? "border-b border-slate-200 bg-white/80 backdrop-blur-md"
          : "border-b border-transparent bg-transparent"
      }`}
    >
      <LandingHeader isLoggedIn={isLoggedIn} />
    </div>
  );
}
