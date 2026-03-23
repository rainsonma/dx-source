import { HallMainShell } from "@/features/web/hall/components/hall-main-shell";

export default function HallMainLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <HallMainShell>{children}</HallMainShell>;
}
