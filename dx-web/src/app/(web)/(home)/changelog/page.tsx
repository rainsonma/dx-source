import { cookies } from "next/headers";
import { LandingHeader } from "@/components/in/landing-header";
import { ChangelogTimeline } from "@/features/web/home/components/changelog-timeline";
import { Footer } from "@/components/in/footer";

export default async function ChangelogPage() {
  const cookieStore = await cookies();
  const isLoggedIn = !!cookieStore.get("dx_token")?.value;

  return (
    <div className="flex min-h-screen flex-col bg-white">
      <LandingHeader isLoggedIn={isLoggedIn} />
      <div className="h-px w-full bg-slate-200" />
      <ChangelogTimeline />
      <Footer />
    </div>
  );
}
