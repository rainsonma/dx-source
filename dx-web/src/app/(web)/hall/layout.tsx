import { HallThemeProvider } from "@/features/web/hall/components/hall-theme-provider";
import { SWRProvider } from "@/lib/swr";
import { PkInvitationProvider } from "@/features/web/play-pk/components/pk-invitation-provider";

export default function HallLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <HallThemeProvider>
      <SWRProvider>
        <PkInvitationProvider>{children}</PkInvitationProvider>
      </SWRProvider>
    </HallThemeProvider>
  );
}
