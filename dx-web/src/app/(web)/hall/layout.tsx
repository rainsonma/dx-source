import { HallThemeProvider } from "@/features/web/hall/components/hall-theme-provider";
import { SWRProvider } from "@/lib/swr";
import { WebSocketProvider } from "@/providers/websocket-provider";
import { PkInvitationProvider } from "@/features/web/play-pk/components/pk-invitation-provider";

export default function HallLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <HallThemeProvider>
      <SWRProvider>
        <WebSocketProvider>
          <PkInvitationProvider>{children}</PkInvitationProvider>
        </WebSocketProvider>
      </SWRProvider>
    </HallThemeProvider>
  );
}
