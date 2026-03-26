import { Sparkles } from "lucide-react";

type BlockState = "default" | "eliminated" | "selected";

const elimGrid: { text: string; state: BlockState }[][] = [
  [
    { text: "apple", state: "eliminated" },
    { text: "篮", state: "default" },
    { text: "rain", state: "default" },
    { text: "苹果", state: "eliminated" },
  ],
  [
    { text: "鱼", state: "default" },
    { text: "book", state: "selected" },
    { text: "鸡", state: "eliminated" },
    { text: "太阳", state: "default" },
  ],
  [
    { text: "sun", state: "default" },
    { text: "鸡", state: "eliminated" },
    { text: "鸟", state: "default" },
    { text: "水", state: "default" },
  ],
  [
    { text: "cat", state: "default" },
    { text: "雨", state: "eliminated" },
    { text: "水", state: "default" },
    { text: "狗", state: "eliminated" },
  ],
];

export function VocabEliminationGame() {
  return (
    <div className="flex w-full max-w-[700px] flex-col items-center gap-5">
      {/* Status row */}
      <div className="flex w-full flex-col items-center gap-3 sm:flex-row sm:justify-between">
        <span className="text-sm font-semibold text-foreground">
          已消除 3/8 对
        </span>
        <div className="h-1.5 w-full max-w-[300px] rounded-full bg-border">
          <div className="h-1.5 w-[113px] rounded-full bg-gradient-to-r from-pink-500 to-teal-500" />
        </div>
        <div className="flex items-center gap-1.5 rounded-lg border border-pink-500/20 bg-pink-50 px-3 py-1">
          <Sparkles className="h-3.5 w-3.5 text-pink-500" />
          <span className="text-xs font-bold text-pink-500">连击 ×2</span>
        </div>
      </div>

      {/* Grid */}
      <div className="flex w-full flex-col gap-2.5 rounded-[20px] border border-border bg-card p-4 shadow-sm md:p-6">
        {elimGrid.map((row, ri) => (
          <div key={ri} className="flex gap-2.5">
            {row.map((block, ci) => (
              <button
                key={ci}
                type="button"
                className={`flex h-14 flex-1 items-center justify-center rounded-xl border md:h-16 ${
                  block.state === "eliminated"
                    ? "border-border/20 bg-muted opacity-40"
                    : block.state === "selected"
                      ? "border-2 border-pink-500 bg-pink-50"
                      : "border-[1.5px] border-border bg-card"
                }`}
              >
                <span
                  className={`text-sm font-medium ${
                    block.state === "eliminated"
                      ? "text-muted-foreground line-through"
                      : block.state === "selected"
                        ? "text-pink-600"
                        : "text-foreground"
                  }`}
                >
                  {block.text}
                </span>
              </button>
            ))}
          </div>
        ))}
      </div>

      <p className="text-xs text-muted-foreground">
        点击两个匹配的方块进行消除
      </p>
    </div>
  );
}
