import { cookies } from "next/headers";
import { LandingHeader } from "@/components/in/landing-header";
import { FeaturesContent } from "@/features/web/home/components/features-content";
import { Footer } from "@/components/in/footer";

export default async function FeaturesPage() {
  const cookieStore = await cookies();
  const isLoggedIn = !!cookieStore.get("dx_token")?.value;

  return (
    <div className="flex min-h-screen flex-col bg-white">
      <LandingHeader isLoggedIn={isLoggedIn} />
      <FeaturesContent />
      <Footer />
    </div>
  );
}
