import { LandingHeader } from "@/components/in/landing-header";

interface StickyHeaderProps {
  isLoggedIn?: boolean;
}

export function StickyHeader({ isLoggedIn = false }: StickyHeaderProps) {
  return (
    <div className="sticky top-0 z-50 w-full border-b border-slate-200 bg-white/80 backdrop-blur-md">
      <LandingHeader isLoggedIn={isLoggedIn} />
    </div>
  );
}
