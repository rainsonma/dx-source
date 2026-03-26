import { Zap, Circle, CheckCircle2 } from "lucide-react";

const matchWords = [
  { word: "apple", matched: true },
  { word: "banana", selected: true },
  { word: "orange", matched: false },
  { word: "grape", matched: false },
  { word: "milk", matched: false },
];

const matchDefs = [
  { def: "苹果", matched: true },
  { def: "牛奶", matched: false },
  { def: "橙子", matched: false },
  { def: "葡萄", matched: false },
  { def: "橘子", matched: false },
];

export function GameVocabMatch() {
  return (
    <div className="flex w-full max-w-3xl flex-col gap-7 rounded-[20px] border border-border bg-card p-6 shadow-sm md:p-8">
      {/* Progress */}
      <div className="flex flex-col gap-2.5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-foreground">
              进度 4/10
            </span>
          </div>
          <div className="flex items-center gap-1.5 rounded-lg bg-teal-600/10 px-3 py-1">
            <Zap className="h-3.5 w-3.5 text-teal-600" />
            <span className="text-xs font-bold text-teal-600">连击 ×3</span>
          </div>
        </div>
        <div className="h-1.5 w-full rounded-full bg-border">
          <div className="h-1.5 w-2/5 rounded-full bg-gradient-to-r from-blue-500 to-teal-500" />
        </div>
      </div>

      {/* Match area */}
      <div className="flex flex-col gap-4 sm:flex-row sm:gap-6">
        {/* English words */}
        <div className="flex flex-1 flex-col gap-2.5">
          <span className="text-xs font-semibold text-muted-foreground">
            英文单词
          </span>
          {matchWords.map((item) => (
            <button
              key={item.word}
              type="button"
              className={`flex items-center gap-2.5 rounded-xl border px-4 py-3 ${
                item.matched
                  ? "border-emerald-300 bg-emerald-50"
                  : item.selected
                    ? "border-blue-400 bg-blue-50"
                    : "border-border bg-card"
              }`}
            >
              {item.matched ? (
                <CheckCircle2 className="h-4 w-4 text-emerald-500" />
              ) : (
                <Circle
                  className={`h-4 w-4 ${item.selected ? "text-blue-400" : "text-slate-300"}`}
                />
              )}
              <span
                className={`text-sm font-medium ${
                  item.matched
                    ? "text-emerald-600"
                    : item.selected
                      ? "text-blue-600"
                      : "text-foreground"
                }`}
              >
                {item.word}
              </span>
            </button>
          ))}
        </div>

        {/* Chinese definitions */}
        <div className="flex flex-1 flex-col gap-2.5">
          <span className="text-xs font-semibold text-muted-foreground">
            中文释义
          </span>
          {matchDefs.map((item) => (
            <button
              key={item.def}
              type="button"
              className={`flex items-center justify-center rounded-xl border px-4 py-3 ${
                item.matched
                  ? "border-emerald-300 bg-emerald-50"
                  : "border-border bg-card"
              }`}
            >
              <span
                className={`text-sm font-medium ${
                  item.matched ? "text-emerald-600" : "text-foreground"
                }`}
              >
                {item.def}
              </span>
            </button>
          ))}
        </div>
      </div>

      <p className="text-center text-xs text-muted-foreground">
        点击左侧单词，再点击右侧匹配的释义
      </p>
    </div>
  );
}
