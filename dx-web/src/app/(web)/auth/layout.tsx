import { AuthHeader } from "@/components/in/auth-header";
import { Footer } from "@/components/in/footer";

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen w-full flex-col items-center bg-gradient-to-b from-teal-100 via-blue-100 via-violet-100 via-pink-100 to-white">
      <AuthHeader />
      <div className="flex flex-1 items-center justify-center py-24">
        {children}
      </div>
      <Footer />
    </div>
  );
}
