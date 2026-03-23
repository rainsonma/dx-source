import {
  ArrowLeft,
  Settings,
  RotateCcw,
  Square,
  Bot,
  Send,
  BookOpen,
  MessageCircle,
  Coffee,
} from "lucide-react";

const aiMessages = [
  {
    greeting: "Hi there! Welcome to the coffee shop! ☕",
    body: "I'm your AI barista today. What can I get for you? We have some great specials — would you like to hear about them, or do you already know what you'd like to order?",
    hints: ["What specials do you have?", "I'd like a latte."],
  },
  {
    body: "Great choice! We have a few oat milk options today:\n\n• Oat Milk Vanilla Latte — smooth and slightly sweet\n• Iced Oat Cappuccino — perfect for a warm day\n• Oat Caramel Macchiato — our most popular!\n\nWould you like to try any of these? I can also make any drink on our menu with oat milk. 😊",
    vocabNote: "🌟 New vocab: macchiato — 玛奇朵（一种咖啡饮品）",
  },
];

const userMessage = "Hi! Could you tell me about today's specials? I'm looking for something with oat milk.";

const vocabulary = [
  { word: "specials", meaning: "n. 特价商品；特色菜" },
  { word: "oat milk", meaning: "n. 燕麦奶" },
  { word: "macchiato", meaning: "n. 玛奇朵咖啡" },
];

export function AiChatPanel({ id }: { id: string }) {
  return (
    <div className="flex h-full flex-col">
      {/* Top bar */}
      <div className="flex w-full items-center justify-between border-b border-border bg-card px-4 py-3.5 md:px-6">
        <div className="flex items-center gap-3">
          <button
            type="button"
            className="flex items-center gap-1 rounded-lg border border-border px-2.5 py-1.5 text-xs font-medium text-muted-foreground"
          >
            <ArrowLeft className="h-4 w-4" />
            返回
          </button>
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-[10px] bg-amber-100">
              <Coffee className="h-[18px] w-[18px] text-amber-600" />
            </div>
            <div className="flex flex-col gap-0.5">
              <span className="text-sm font-semibold text-foreground">
                咖啡店点餐
              </span>
              <span className="text-[11px] text-muted-foreground">
                话题 #{id} · 初级
              </span>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            aria-label="设置"
            className="hidden h-9 w-9 items-center justify-center rounded-lg border border-border bg-muted sm:flex"
          >
            <Settings className="h-4 w-4 text-muted-foreground" />
          </button>
          <button
            type="button"
            aria-label="重置对话"
            className="hidden h-9 w-9 items-center justify-center rounded-lg border border-border bg-muted sm:flex"
          >
            <RotateCcw className="h-4 w-4 text-muted-foreground" />
          </button>
          <button
            type="button"
            className="flex items-center gap-1.5 rounded-lg bg-red-100 px-3.5 py-2 text-xs font-semibold text-red-600"
          >
            <Square className="h-3.5 w-3.5" />
            结束对话
          </button>
        </div>
      </div>

      {/* Chat body */}
      <div className="flex flex-1 flex-col overflow-hidden lg:flex-row">
        {/* Chat main */}
        <div className="flex flex-1 flex-col justify-between">
          {/* Messages */}
          <div className="flex flex-col gap-5 overflow-y-auto px-4 py-6 md:px-8">
            {/* AI message 1 */}
            <div className="flex gap-3">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[10px] bg-gradient-to-br from-teal-600 to-teal-800">
                <Bot className="h-5 w-5 text-white" />
              </div>
              <div className="flex max-w-[520px] flex-col gap-3 rounded-[2px_14px_14px_14px] border border-border bg-card p-4">
                <p className="text-sm font-semibold leading-relaxed text-foreground">
                  {aiMessages[0].greeting}
                </p>
                <p className="text-sm leading-relaxed text-foreground">
                  {aiMessages[0].body}
                </p>
                <div className="flex flex-wrap gap-2">
                  {aiMessages[0].hints?.map((hint) => (
                    <button
                      key={hint}
                      type="button"
                      className="rounded-md border border-teal-200 bg-teal-50 px-3 py-1.5 text-xs font-medium text-teal-600"
                    >
                      {hint}
                    </button>
                  ))}
                </div>
              </div>
            </div>

            {/* User message */}
            <div className="flex justify-end">
              <div className="max-w-[400px] rounded-[14px_2px_14px_14px] bg-teal-600 p-4">
                <p className="text-sm leading-relaxed text-white">
                  {userMessage}
                </p>
              </div>
            </div>

            {/* AI message 2 */}
            <div className="flex gap-3">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[10px] bg-gradient-to-br from-teal-600 to-teal-800">
                <Bot className="h-5 w-5 text-white" />
              </div>
              <div className="flex max-w-[520px] flex-col gap-2.5 rounded-[2px_14px_14px_14px] border border-border bg-card p-4">
                <p className="whitespace-pre-line text-sm leading-relaxed text-foreground">
                  {aiMessages[1].body}
                </p>
                {aiMessages[1].vocabNote && (
                  <div className="rounded-lg bg-amber-50 px-3 py-2">
                    <span className="text-xs text-amber-800">
                      {aiMessages[1].vocabNote}
                    </span>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Input area */}
          <div className="border-t border-border bg-card px-4 py-4 md:px-8">
            <div className="flex items-center gap-2.5">
              <div className="flex h-12 flex-1 items-center rounded-xl border border-border bg-muted px-4">
                <span className="text-sm text-muted-foreground">
                  Type your reply in English...
                </span>
              </div>
              <button
                type="button"
                aria-label="发送"
                className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[10px] bg-teal-600"
              >
                <Send className="h-4 w-4 text-white" />
              </button>
            </div>
            <div className="mt-2.5 flex items-center gap-4">
              <span className="text-[11px] font-medium text-muted-foreground">💡 提示：</span>
              <span className="text-[11px] text-muted-foreground">
                尝试用完整句子回答，说出你想点的饮品
              </span>
            </div>
          </div>
        </div>

        {/* Right panel */}
        <div className="flex w-full shrink-0 flex-col gap-5 border-t border-border bg-card p-5 lg:w-[300px] lg:border-l lg:border-t-0">
          {/* Vocabulary */}
          <div className="flex flex-col gap-3">
            <div className="flex items-center gap-2">
              <BookOpen className="h-4 w-4 text-teal-600" />
              <span className="text-sm font-bold text-foreground">本次词汇</span>
            </div>
            {vocabulary.map((v) => (
              <div
                key={v.word}
                className="flex flex-col gap-1 rounded-[10px] bg-muted px-3.5 py-2.5"
              >
                <span className="text-[13px] font-semibold text-foreground">{v.word}</span>
                <span className="text-[11px] text-muted-foreground">{v.meaning}</span>
              </div>
            ))}
          </div>

          <div className="h-px w-full bg-border" />

          {/* Expression tips */}
          <div className="flex flex-col gap-3">
            <div className="flex items-center gap-2">
              <MessageCircle className="h-4 w-4 text-teal-600" />
              <span className="text-sm font-bold text-foreground">表达建议</span>
            </div>
            <div className="flex flex-col gap-1.5 rounded-[10px] border border-teal-200 bg-teal-50 p-3.5">
              <span className="text-xs font-semibold text-teal-600">
                &quot;I&apos;d like to try the...&quot;
              </span>
              <span className="text-[11px] leading-snug text-muted-foreground">
                用 &quot;I&apos;d like&quot; 比 &quot;I want&quot; 更礼貌哦
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
