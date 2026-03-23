import { LandingHeader } from "@/components/in/landing-header";
import { ChangelogTimeline } from "@/features/web/home/components/changelog-timeline";
import { Footer } from "@/components/in/footer";

export default function ChangelogPage() {
  return (
    <div className="flex min-h-screen flex-col bg-white">
      <LandingHeader />
      <div className="h-px w-full bg-slate-200" />
      <ChangelogTimeline />
      <Footer />
    </div>
  );
}
