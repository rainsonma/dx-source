import { LandingHeader } from "@/components/in/landing-header";
import { FeaturesContent } from "@/features/web/home/components/features-content";
import { Footer } from "@/components/in/footer";

export default function FeaturesPage() {
  return (
    <div className="flex min-h-screen flex-col bg-white">
      <LandingHeader />
      <FeaturesContent />
      <Footer />
    </div>
  );
}
