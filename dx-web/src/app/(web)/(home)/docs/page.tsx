import { LandingHeader } from "@/components/in/landing-header";
import { DocsPageContent } from "@/features/web/home/components/docs/docs-content";
import { Footer } from "@/components/in/footer";

export default function DocsPage() {
  return (
    <div className="flex min-h-screen flex-col bg-white">
      <LandingHeader />
      <div className="h-px w-full bg-slate-200" />
      <DocsPageContent />
      <Footer />
    </div>
  );
}
