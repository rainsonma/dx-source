"use client";

import { useEffect, useRef, useState } from "react";
import { Clock } from "lucide-react";

type Props = {
  minutes: number;
  onExpire: () => void;
};

export function LevelTimer({ minutes, onExpire }: Props) {
  const [remaining, setRemaining] = useState(minutes * 60);
  const onExpireRef = useRef(onExpire);
  onExpireRef.current = onExpire;

  useEffect(() => {
    setRemaining(minutes * 60);
    const interval = setInterval(() => {
      setRemaining((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          onExpireRef.current();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(interval);
  }, [minutes]);

  const mins = Math.floor(remaining / 60);
  const secs = remaining % 60;
  const isLow = remaining <= 60;

  return (
    <div className={`flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-mono tabular-nums ${
      isLow ? "bg-red-100 text-red-600" : "bg-teal-100 text-teal-700"
    }`}>
      <Clock className="h-3.5 w-3.5" />
      {String(mins).padStart(2, "0")}:{String(secs).padStart(2, "0")}
    </div>
  );
}
