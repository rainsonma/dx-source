import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const authRoutes = ["/auth/signin", "/auth/signup"];
const protectedRoutes = ["/hall"];

export default function middleware(request: NextRequest) {
  const hasToken = !!request.cookies.get("dx_token")?.value;
  const { pathname } = request.nextUrl;

  // Users with token should not see signin/signup pages
  if (hasToken && authRoutes.some((r) => pathname.startsWith(r))) {
    return NextResponse.redirect(new URL("/hall", request.url));
  }

  // Users without token cannot access protected routes
  if (!hasToken && protectedRoutes.some((r) => pathname.startsWith(r))) {
    return NextResponse.redirect(new URL("/auth/signin", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/hall/:path*", "/auth/signin", "/auth/signup"],
};
