import { NextResponse, type NextRequest } from "next/server";

const API_URL =
  process.env.API_INTERNAL_URL ||
  process.env.NEXT_PUBLIC_API_URL ||
  "http://localhost:3001";

type Params = {
  params: Promise<{ code: string }>;
};

/**
 * Validate an invite code via the Go API, set the `ref` cookie on the redirect
 * response, and send the user to the sign-up page. A Route Handler is required
 * because Server Components cannot mutate cookies in Next.js.
 */
export async function GET(request: NextRequest, { params }: Params) {
  const { code } = await params;

  const signupUrl = new URL("/auth/signup", request.nextUrl.origin);
  const response = NextResponse.redirect(signupUrl);

  try {
    const res = await fetch(
      `${API_URL}/api/invite/validate?code=${encodeURIComponent(code)}`,
      { cache: "no-store" },
    );
    const data: { code: number; data: { valid: boolean } } = await res.json();

    if (data.code === 0 && data.data.valid) {
      response.cookies.set("ref", code, {
        httpOnly: true,
        secure: process.env.NODE_ENV === "production",
        sameSite: "lax",
        maxAge: 60 * 60 * 24 * 7, // 7 days
        path: "/",
      });
    }
  } catch {
    // Ignore errors — still redirect to sign-up with no cookie set.
  }

  return response;
}
