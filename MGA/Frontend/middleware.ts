import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  const token = request.cookies.get("apiToken")?.value;

  if (!token && request.nextUrl.pathname.startsWith("/dashboard")) {
    const signInUrl = request.nextUrl.clone();
    signInUrl.pathname = "/";
    signInUrl.searchParams.set("reason", "auth");
    return NextResponse.redirect(signInUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*"],
};

