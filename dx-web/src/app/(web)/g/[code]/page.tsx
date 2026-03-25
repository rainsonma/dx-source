import { GroupInviteContent } from "@/features/web/groups/components/group-invite-content";

export default async function GroupInvitePage({
  params,
}: {
  params: Promise<{ code: string }>;
}) {
  const { code } = await params;
  return (
    <div className="flex min-h-screen items-center justify-center px-4 py-12">
      <GroupInviteContent code={code} />
    </div>
  );
}
