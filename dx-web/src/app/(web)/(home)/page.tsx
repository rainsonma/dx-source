import { LandingHeader } from "@/components/in/landing-header";
import { HeroSection } from "@/features/web/home/components/hero-section";
import { FeaturesSection } from "@/features/web/home/components/features-section";
import { AiFeaturesSection } from "@/features/web/home/components/ai-features-section";
import { CoursePlatformSection } from "@/features/web/home/components/course-platform-section";
import { SmartVocabularySection } from "@/features/web/home/components/smart-vocabulary-section";
import { SocialCommunitySection } from "@/features/web/home/components/social-community-section";
import { StatsSection } from "@/features/web/home/components/stats-section";
import { TestimonialsSection } from "@/features/web/home/components/testimonials-section";
import { FinalCtaSection } from "@/features/web/home/components/final-cta-section";
import { Footer } from "@/components/in/footer";

export default function HomePage() {
  return (
    <div className="flex min-h-screen w-full flex-col">
      {/* Hero wrapper with gradient background */}
      <div className="flex w-full flex-col bg-gradient-to-b from-teal-100 via-blue-100 via-violet-100 via-pink-100 to-white">
        <LandingHeader />
        <HeroSection />
      </div>
      <FeaturesSection />
      <AiFeaturesSection />
      <CoursePlatformSection />
      <SmartVocabularySection />
      <SocialCommunitySection />
      <StatsSection />
      <TestimonialsSection />
      <FinalCtaSection />
      <Footer />
    </div>
  );
}
