import { cookies } from "next/headers";
import { GroupInviteContent } from "@/features/web/groups/components/group-invite-content";

export default async function GroupInvitePage({
  params,
}: {
  params: Promise<{ code: string }>;
}) {
  const { code } = await params;
  const cookieStore = await cookies();
  const isLoggedIn = !!cookieStore.get("dx_refresh")?.value;

  return (
    <div className="flex min-h-screen items-center justify-center px-4 py-12">
      <GroupInviteContent code={code} isLoggedIn={isLoggedIn} />
    </div>
  );
}
