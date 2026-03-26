import { GroupGameRoom } from "@/features/web/groups/components/group-game-room";

export default async function GroupGameRoomPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return (
    <div className="flex min-h-full flex-col items-center justify-center px-4 py-12">
      <GroupGameRoom groupId={id} />
    </div>
  );
}
