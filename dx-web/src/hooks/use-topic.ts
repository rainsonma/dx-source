"use client";

import { useEffect, useRef } from "react";

import { useWS } from "@/providers/websocket-provider";

export function useTopic(
  topic: string | null,
  handlers: Record<string, (data: unknown) => void>,
): void {
  const { subscribe } = useWS();

  const handlersRef = useRef(handlers);
  useEffect(() => {
    handlersRef.current = handlers;
  });

  useEffect(() => {
    if (!topic) return;
    const unsubscribe = subscribe(topic, (event) => {
      handlersRef.current[event.type]?.(event.data);
    });
    return unsubscribe;
  }, [topic, subscribe]);
}
