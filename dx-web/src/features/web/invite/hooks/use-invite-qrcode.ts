"use client";

import { useEffect, useRef } from "react";

/** Render a QR code into a container element using easyqrcodejs */
export function useInviteQrcode(url: string) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!containerRef.current || !url) return;

    let cancelled = false;

    const renderQrcode = async () => {
      const QRCode = (await import("easyqrcodejs")).default;

      if (cancelled || !containerRef.current) return;

      containerRef.current.innerHTML = "";

      new QRCode(containerRef.current, {
        text: url,
        width: 100,
        height: 100,
        colorDark: "#0d9488",
        colorLight: "#ffffff",
        correctLevel: QRCode.CorrectLevel.M,
        quietZone: 4,
        quietZoneColor: "#ffffff",
      });
    };

    renderQrcode();

    return () => {
      cancelled = true;
      if (containerRef.current) {
        containerRef.current.innerHTML = "";
      }
    };
  }, [url]);

  return containerRef;
}
