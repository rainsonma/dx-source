"use client";

import { useEffect, useRef, useState } from "react";

type Props = {
  seconds: number;
  onExpire: () => void;
  resetKey: string | number;
};

export function AnswerTimer({ seconds, onExpire, resetKey }: Props) {
  const [remaining, setRemaining] = useState(seconds);
  const onExpireRef = useRef(onExpire);
  onExpireRef.current = onExpire;

  useEffect(() => {
    setRemaining(seconds);
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
  }, [seconds, resetKey]);

  const ratio = remaining / seconds;

  return (
    <div className="flex items-center gap-1.5">
      <div className="h-1.5 w-16 overflow-hidden rounded-full bg-muted">
        <div
          className={`h-full rounded-full transition-all duration-1000 ${
            ratio > 0.3 ? "bg-teal-500" : "bg-red-500"
          }`}
          style={{ width: `${ratio * 100}%` }}
        />
      </div>
      <span className={`text-xs font-mono tabular-nums ${remaining <= 3 ? "text-red-500" : "text-muted-foreground"}`}>
        {remaining}s
      </span>
    </div>
  );
}
