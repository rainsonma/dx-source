import { AuthGuard } from "@/components/in/auth-guard";
import { HallThemeProvider } from "@/features/web/hall/components/hall-theme-provider";

export default function HallLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <HallThemeProvider>
      <AuthGuard>{children}</AuthGuard>
    </HallThemeProvider>
  );
}
