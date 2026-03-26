import { Headphones, CheckCircle2 } from "lucide-react";

const listeningOptions = [
  { label: "A", text: "It will rain tomorrow", correct: false, selected: false },
  { label: "B", text: "It's sunny today", correct: true, selected: true },
  { label: "C", text: "The weather is cold", correct: false, selected: false },
  { label: "D", text: "It was cloudy yesterday", correct: false, selected: false },
];

const waveformHeights = [14, 20, 10, 22, 12, 18, 24, 11, 16, 21, 13, 19, 15, 23, 10, 17, 22, 12, 20, 14, 18, 11, 16, 24];

export function GameListening() {
  return (
    <div className="flex w-full max-w-[720px] flex-col gap-7 rounded-[20px] border border-border bg-card p-6 shadow-sm md:p-8">
      {/* Progress */}
      <div className="flex flex-col gap-2.5">
        <div className="flex items-center justify-between">
          <span className="text-sm font-semibold text-foreground">
            第 4 题 | 共 10 题
          </span>
          <button
            type="button"
            className="flex items-center gap-1 rounded-lg bg-teal-600/10 px-2.5 py-1 text-[11px] font-semibold text-teal-600"
          >
            📝 记录
          </button>
        </div>
        <div className="h-1.5 w-full rounded-full bg-border">
          <div className="h-1.5 w-2/5 rounded-full bg-amber-500" />
        </div>
      </div>

      <div className="h-px w-full bg-muted" />

      {/* Audio player */}
      <div className="flex flex-col items-center gap-4">
        <button
          type="button"
          aria-label="播放音频"
          className="flex h-[72px] w-[72px] items-center justify-center rounded-full bg-amber-500 shadow-lg shadow-amber-500/30"
        >
          <Headphones className="h-8 w-8 text-white" />
        </button>
        <div className="flex h-8 items-center justify-center gap-[3px]">
          {waveformHeights.map((h, i) => (
            <div
              key={i}
              className="w-[3px] rounded-full bg-amber-400"
              style={{ height: `${h}px` }}
            />
          ))}
        </div>
        <div className="flex items-center gap-3">
          <span className="text-[11px] text-muted-foreground">已播放次数</span>
          <span className="text-[11px] font-medium text-amber-600">
            已播放 1 次
          </span>
        </div>
      </div>

      <div className="h-px w-full bg-muted" />

      {/* Question */}
      <div className="flex flex-col items-center gap-5">
        <p className="text-center text-lg font-semibold text-foreground">
          What did the speaker say about the weather?
        </p>

        {/* Options */}
        <div className="flex w-full flex-col gap-3">
          {listeningOptions.map((opt) => (
            <button
              key={opt.label}
              type="button"
              className={`flex items-center gap-3 rounded-2xl border px-5 py-3.5 ${
                opt.selected && opt.correct
                  ? "border-amber-400 bg-amber-50"
                  : "border-border bg-card"
              }`}
            >
              <span
                className={`text-sm font-bold ${
                  opt.selected && opt.correct
                    ? "text-amber-600"
                    : "text-muted-foreground"
                }`}
              >
                {opt.label}
              </span>
              <span
                className={`flex-1 text-sm ${
                  opt.selected && opt.correct
                    ? "font-medium text-amber-700"
                    : "text-muted-foreground"
                }`}
              >
                {opt.text}
              </span>
              {opt.selected && opt.correct && (
                <CheckCircle2 className="h-5 w-5 text-amber-500" />
              )}
            </button>
          ))}
        </div>
      </div>

      <p className="text-center text-xs text-muted-foreground">
        听完语音后选择正确的答案
      </p>
    </div>
  );
}
