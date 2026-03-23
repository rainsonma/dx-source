"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getAccessToken } from "@/lib/token";
import { refreshAccessToken } from "@/lib/api-client";
import { SWRProvider } from "@/lib/swr";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [status, setStatus] = useState<"loading" | "authenticated" | "unauthenticated">("loading");

  useEffect(() => {
    async function checkAuth() {
      if (getAccessToken()) {
        setStatus("authenticated");
        return;
      }

      try {
        await refreshAccessToken();
        setStatus("authenticated");
      } catch {
        setStatus("unauthenticated");
        router.replace("/auth/signin");
      }
    }

    checkAuth();
  }, [router]);

  if (status === "loading") {
    return null;
  }

  if (status === "unauthenticated") {
    return null;
  }

  return <SWRProvider>{children}</SWRProvider>;
}
