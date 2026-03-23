import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const authRoutes = ["/auth/signin", "/auth/signup"];
const protectedRoutes = ["/hall"];

export default function middleware(request: NextRequest) {
  const hasRefreshToken = !!request.cookies.get("dx_refresh")?.value;
  const { pathname } = request.nextUrl;

  // Clear legacy dx_token cookie if present
  const legacyToken = request.cookies.get("dx_token")?.value;

  // Users with refresh token should not see signin/signup pages
  if (hasRefreshToken && authRoutes.some((r) => pathname.startsWith(r))) {
    const response = NextResponse.redirect(new URL("/hall", request.url));
    if (legacyToken) response.cookies.delete("dx_token");
    return response;
  }

  // Users without refresh token cannot access protected routes
  if (!hasRefreshToken && protectedRoutes.some((r) => pathname.startsWith(r))) {
    const response = NextResponse.redirect(new URL("/auth/signin", request.url));
    if (legacyToken) response.cookies.delete("dx_token");
    return response;
  }

  const response = NextResponse.next();
  if (legacyToken) response.cookies.delete("dx_token");
  return response;
}

export const config = {
  matcher: ["/hall/:path*", "/auth/signin", "/auth/signup"],
};
