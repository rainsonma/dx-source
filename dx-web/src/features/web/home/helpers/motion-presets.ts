// dx-web/src/features/web/home/helpers/motion-presets.ts
import type { Variants } from "motion/react";

export const revealEase = { duration: 0.45, ease: [0.22, 1, 0.36, 1] } as const;
export const revealSpring = {
  type: "spring",
  stiffness: 160,
  damping: 24,
  mass: 0.6,
} as const;

export const revealVariants: Variants = {
  hidden: { opacity: 0, y: 24 },
  show: { opacity: 1, y: 0, transition: revealEase },
};

export const staggerContainerVariants: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
};

export const staggerChildVariants: Variants = {
  hidden: { opacity: 0, y: 16 },
  show: { opacity: 1, y: 0, transition: revealEase },
};
