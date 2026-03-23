import { cookies } from "next/headers";
import { redirect } from "next/navigation";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

type Props = {
  params: Promise<{ code: string }>;
};

/** Validate invite code via Go API, set ref cookie, redirect to sign-up */
export default async function InviteRedirectPage({ params }: Props) {
  const { code } = await params;

  try {
    const res = await fetch(
      `${API_URL}/api/invite/validate?code=${encodeURIComponent(code)}`
    );
    const data: { code: number; data: { valid: boolean } } = await res.json();

    if (data.code === 0 && data.data.valid) {
      const cookieStore = await cookies();
      cookieStore.set("ref", code, {
        httpOnly: true,
        secure: process.env.NODE_ENV === "production",
        sameSite: "lax",
        maxAge: 60 * 60 * 24 * 7, // 7 days
        path: "/",
      });
    }
  } catch {
    // Ignore errors — just redirect to sign-up
  }

  redirect("/auth/signup");
}
