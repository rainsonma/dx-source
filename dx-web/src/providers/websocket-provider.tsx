"use client";

import { createContext, useContext, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "";

type EventHandler = (event: { type: string; data: unknown }) => void;

type WSContextValue = {
  status: "connecting" | "open" | "closed";
  subscribe: (topic: string, handler: EventHandler) => () => void;
};

const WSContext = createContext<WSContextValue | null>(null);

type Envelope = {
  op: "subscribe" | "unsubscribe" | "event" | "ack" | "error";
  topic?: string;
  type?: string;
  data?: unknown;
  id?: string;
  ok?: boolean;
  code?: number;
  message?: string;
};

function randomId(): string {
  return Math.random().toString(36).slice(2) + Date.now().toString(36);
}

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const [status, setStatus] = useState<"connecting" | "open" | "closed">("connecting");

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttempt = useRef(0);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const hasEverConnected = useRef(false);

  const subsRef = useRef<Map<string, Set<EventHandler>>>(new Map());
  const ackedRef = useRef<Set<string>>(new Set());

  const router = useRouter();

  useEffect(() => {
    let cancelled = false;

    const sendSubscribe = (ws: WebSocket, topic: string) => {
      const env: Envelope = { op: "subscribe", topic, id: randomId() };
      ws.send(JSON.stringify(env));
    };

    const scheduleReconnect = () => {
      const n = reconnectAttempt.current++;
      if (n >= 10) {
        toast.error("连接已断开，请刷新页面");
        return;
      }
      const base = Math.min(1000 * Math.pow(2, n), 30000);
      const jitter = Math.random() * Math.min(base * 0.3, 5000);
      const delay = base + jitter;
      reconnectTimer.current = setTimeout(connect, delay);
    };

    const routeIncoming = (env: Envelope) => {
      if (env.op === "event" && env.topic && env.type) {
        const handlers = subsRef.current.get(env.topic);
        handlers?.forEach((h) => h({ type: env.type!, data: env.data }));
        return;
      }
      if (env.op === "ack" && env.ok && env.topic) {
        ackedRef.current.add(env.topic);
        return;
      }
    };

    const connect = () => {
      if (cancelled) return;
      setStatus("connecting");
      const url = `${API_URL}/api/ws`;
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        setStatus("open");
        hasEverConnected.current = true;
        reconnectAttempt.current = 0;
        ackedRef.current.clear();
        for (const topic of subsRef.current.keys()) {
          sendSubscribe(ws, topic);
        }
      };

      ws.onmessage = (ev: MessageEvent) => {
        try {
          const env = JSON.parse(ev.data as string) as Envelope;
          routeIncoming(env);
        } catch {
          // Discard malformed frames
        }
      };

      ws.onclose = (ev: CloseEvent) => {
        wsRef.current = null;
        setStatus("closed");
        ackedRef.current.clear();

        if (cancelled) return;

        if (ev.code === 4001) {
          toast.error("您的账号已在其他设备登录");
          router.push("/auth/signin?reason=session_replaced");
          return;
        }
        if (ev.code === 4401) {
          router.push("/auth/signin?reason=session_expired");
          return;
        }
        if (ev.code === 1006 && !hasEverConnected.current) {
          router.push("/auth/signin?reason=session_expired");
          return;
        }
        if (ev.code === 1000) {
          return;
        }
        scheduleReconnect();
      };

      ws.onerror = () => {};
    };

    connect();

    return () => {
      cancelled = true;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      if (wsRef.current) {
        wsRef.current.close(1000, "unmount");
        wsRef.current = null;
      }
    };
  }, [router]);

  const subscribe = (topic: string, handler: EventHandler) => {
    let set = subsRef.current.get(topic);
    if (!set) {
      set = new Set();
      subsRef.current.set(topic, set);
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        const env: Envelope = { op: "subscribe", topic, id: randomId() };
        wsRef.current.send(JSON.stringify(env));
      }
    }
    set.add(handler);

    return () => {
      const current = subsRef.current.get(topic);
      if (!current) return;
      current.delete(handler);
      if (current.size === 0) {
        subsRef.current.delete(topic);
        ackedRef.current.delete(topic);
        if (wsRef.current?.readyState === WebSocket.OPEN) {
          const env: Envelope = { op: "unsubscribe", topic };
          wsRef.current.send(JSON.stringify(env));
        }
      }
    };
  };

  return (
    <WSContext.Provider value={{ status, subscribe }}>
      {children}
    </WSContext.Provider>
  );
}

export function useWS(): WSContextValue {
  const ctx = useContext(WSContext);
  if (!ctx) {
    throw new Error("useWS must be used within WebSocketProvider");
  }
  return ctx;
}
