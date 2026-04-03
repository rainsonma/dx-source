import { HallThemeProvider } from "@/features/web/hall/components/hall-theme-provider";
import { SWRProvider } from "@/lib/swr";

export default function HallLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <HallThemeProvider>
      <SWRProvider>{children}</SWRProvider>
    </HallThemeProvider>
  );
}
