import { cookies } from "next/headers";
import { StickyHeader } from "@/components/in/sticky-header";
import { HeroSection } from "@/features/web/home/components/hero-section";
import { WhyDifferentSection } from "@/features/web/home/components/why-different-section";
import { FeaturesSection } from "@/features/web/home/components/features-section";
import { AiFeaturesSection } from "@/features/web/home/components/ai-features-section";
import { LearningLoopSection } from "@/features/web/home/components/learning-loop-section";
import { CommunitySection } from "@/features/web/home/components/community-section";
import { MembershipSection } from "@/features/web/home/components/membership-section";
import { FaqSection } from "@/features/web/home/components/faq-section";
import { FinalCtaSection } from "@/features/web/home/components/final-cta-section";
import { Footer } from "@/components/in/footer";

export default async function HomePage() {
  const cookieStore = await cookies();
  const isLoggedIn = !!cookieStore.get("dx_token")?.value;

  return (
    <div className="relative z-0 flex min-h-screen w-full flex-col">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[620px] bg-gradient-to-b from-teal-100 via-blue-100 via-violet-100 via-pink-100 to-white"
      />
      <StickyHeader isLoggedIn={isLoggedIn} transparent />
      <HeroSection isLoggedIn={isLoggedIn} />
      <WhyDifferentSection />
      <FeaturesSection />
      <AiFeaturesSection />
      <LearningLoopSection />
      <CommunitySection />
      <MembershipSection />
      <FaqSection />
      <FinalCtaSection isLoggedIn={isLoggedIn} />
      <Footer />
    </div>
  );
}
