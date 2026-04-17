// dx-web/src/features/web/home/hooks/use-in-view.ts
"use client";

import { useCallback, useEffect, useState } from "react";

export function usePrefersReducedMotion(): boolean {
  const [reduced, setReduced] = useState(false);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    const apply = () => setReduced(mq.matches);
    apply();
    mq.addEventListener("change", apply);
    return () => mq.removeEventListener("change", apply);
  }, []);

  return reduced;
}

interface UseInViewOptions {
  threshold?: number;
  rootMargin?: string;
  root?: Element | Document | null;
}

export function useInView<T extends Element = HTMLElement>(
  options: UseInViewOptions = {},
): { ref: (node: T | null) => void; inView: boolean } {
  const { threshold = 0.2, rootMargin, root } = options;
  const [node, setNode] = useState<T | null>(null);
  const [inView, setInView] = useState(false);

  const ref = useCallback((next: T | null) => {
    setNode(next);
  }, []);

  useEffect(() => {
    if (!node) return;
    const obs = new IntersectionObserver(
      ([entry]) => {
        setInView(entry.isIntersecting);
      },
      { threshold, rootMargin, root },
    );
    obs.observe(node);
    return () => obs.disconnect();
  }, [node, threshold, rootMargin, root]);

  return { ref, inView };
}
