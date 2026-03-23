import { AiChatPanel } from "@/features/web/ai-practice/components/ai-chat-panel";

export default async function AiPracticeChatPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return <AiChatPanel id={id} />;
}
