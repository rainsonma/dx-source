"use client";

import { useCallback, useEffect, useState } from "react";

type WebkitDocument = Document & {
  webkitExitFullscreen?: () => Promise<void>;
  webkitFullscreenElement?: Element | null;
};

type WebkitHTMLElement = HTMLElement & {
  webkitRequestFullscreen?: () => Promise<void>;
};

/** Manage browser fullscreen state with cross-browser support. */
export function useFullscreen() {
  const [isFullscreen, setIsFullscreen] = useState(false);

  /** Toggle between entering and exiting fullscreen mode. */
  const toggleFullscreen = useCallback(() => {
    const doc = document as WebkitDocument;
    const el = document.documentElement as WebkitHTMLElement;

    try {
      if (doc.fullscreenElement || doc.webkitFullscreenElement) {
        if (doc.exitFullscreen) {
          doc.exitFullscreen();
        } else if (doc.webkitExitFullscreen) {
          doc.webkitExitFullscreen();
        }
      } else {
        if (el.requestFullscreen) {
          el.requestFullscreen();
        } else if (el.webkitRequestFullscreen) {
          el.webkitRequestFullscreen();
        }
      }
    } catch {
      // Silently ignore — fullscreen may not be supported
    }
  }, []);

  useEffect(() => {
    /** Sync React state when fullscreen changes (including ESC key exit). */
    const handleChange = () => {
      const doc = document as WebkitDocument;
      setIsFullscreen(!!(doc.fullscreenElement || doc.webkitFullscreenElement));
    };

    document.addEventListener("fullscreenchange", handleChange);
    document.addEventListener("webkitfullscreenchange", handleChange);

    return () => {
      document.removeEventListener("fullscreenchange", handleChange);
      document.removeEventListener("webkitfullscreenchange", handleChange);
    };
  }, []);

  /** Exit fullscreen when component unmounts (e.g. navigating away). */
  useEffect(() => {
    return () => {
      const doc = document as WebkitDocument;
      if (doc.fullscreenElement || doc.webkitFullscreenElement) {
        doc.exitFullscreen?.().catch(() => {});
      }
    };
  }, []);

  return { isFullscreen, toggleFullscreen };
}
