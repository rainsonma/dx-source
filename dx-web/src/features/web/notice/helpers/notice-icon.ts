import type { LucideIcon } from "lucide-react";
import {
  MessageCircleMore,
  Swords,
  Bell,
  Megaphone,
  Trophy,
  Gift,
  Rocket,
  Star,
  Shield,
  BookOpen,
  Calendar,
  UserPlus,
  Heart,
  Zap,
  PartyPopper,
  Info,
  AlertTriangle,
  CheckCircle2,
  Sparkles,
  Crown,
} from "lucide-react";

/** Map of supported Lucide icon names to components */
const iconMap: Record<string, LucideIcon> = {
  "message-circle-more": MessageCircleMore,
  swords: Swords,
  bell: Bell,
  megaphone: Megaphone,
  trophy: Trophy,
  gift: Gift,
  rocket: Rocket,
  star: Star,
  shield: Shield,
  "book-open": BookOpen,
  calendar: Calendar,
  "user-plus": UserPlus,
  heart: Heart,
  zap: Zap,
  "party-popper": PartyPopper,
  info: Info,
  "alert-triangle": AlertTriangle,
  "check-circle-2": CheckCircle2,
  sparkles: Sparkles,
  crown: Crown,
};

/** Resolve a Lucide icon name string to its component, fallback to MessageCircleMore */
export function resolveNoticeIcon(name?: string | null): LucideIcon {
  if (!name) return MessageCircleMore;
  return iconMap[name] ?? MessageCircleMore;
}
