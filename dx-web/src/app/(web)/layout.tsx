import { PkInvitationProvider } from "@/features/web/play-pk/components/pk-invitation-provider";

export default function WebLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <PkInvitationProvider>{children}</PkInvitationProvider>;
}
