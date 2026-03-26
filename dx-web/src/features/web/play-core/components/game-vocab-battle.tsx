import { Zap } from "lucide-react";

const oppShields = [true, true, false, false, false];
const myShields = [true, true, true, true, false];
const oppLetters = [
  { letter: "h", revealed: false },
  { letter: "e", revealed: false },
  { letter: "l", revealed: false },
  { letter: "l", revealed: false },
  { letter: "o", revealed: false },
];
const myLetters = [
  { letter: "h", revealed: true },
  { letter: "e", revealed: true },
  { letter: "l", revealed: true },
  { letter: "l", revealed: true },
  { letter: "o", revealed: false },
];
const keyboardLetters = ["H", "O", "L", "E", "L"];

export function GameVocabBattle() {
  return (
    <div className="flex w-full max-w-[760px] flex-col rounded-[20px] border border-border bg-card shadow-sm">
      {/* Opponent zone */}
      <div className="flex flex-col items-center gap-4 px-6 py-7 md:px-8">
        <div className="flex items-center gap-2.5">
          <span className="text-xs text-muted-foreground">🤖 对手</span>
        </div>
        <div className="flex items-center justify-center gap-2">
          {oppShields.map((active, i) => (
            <div
              key={i}
              className={`h-6 w-6 rounded-full border-2 ${
                active
                  ? "border-red-400 bg-red-400"
                  : "border-border bg-muted"
              }`}
            />
          ))}
        </div>
        <div className="flex items-center justify-center gap-2.5">
          {oppLetters.map((l, i) => (
            <div
              key={i}
              className="flex h-10 w-10 items-center justify-center rounded-lg border border-border bg-muted"
            >
              <span className="text-sm font-medium text-slate-300">
                {l.revealed ? l.letter : "?"}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Translation zone */}
      <div className="flex flex-col items-center gap-2.5 bg-gradient-to-b from-red-50/0 via-red-50 to-red-50/0 px-6 py-4 md:px-8">
        <p className="text-center text-2xl font-extrabold tracking-wider text-foreground md:text-[32px]">
          你好
        </p>
        <div className="h-0.5 w-full rounded-full bg-gradient-to-r from-red-500/0 via-red-500/30 via-30% via-teal-500/30 via-70% to-teal-500/0" />
      </div>

      {/* My zone */}
      <div className="flex flex-col items-center gap-4 px-6 py-5 md:px-8">
        <div className="flex items-center justify-center gap-2.5">
          {myLetters.map((l, i) => (
            <div
              key={i}
              className={`flex h-10 w-10 items-center justify-center rounded-lg border ${
                l.revealed
                  ? "border-teal-300 bg-teal-50"
                  : "border-border bg-muted"
              }`}
            >
              <span
                className={`text-sm font-semibold ${
                  l.revealed ? "text-teal-600" : "text-slate-300"
                }`}
              >
                {l.revealed ? l.letter : "_"}
              </span>
            </div>
          ))}
        </div>
        <div className="flex items-center justify-center gap-2">
          {myShields.map((active, i) => (
            <div
              key={i}
              className={`h-6 w-6 rounded-full border-2 ${
                active
                  ? "border-teal-400 bg-teal-400"
                  : "border-border bg-muted"
              }`}
            />
          ))}
        </div>
        <div className="flex items-center gap-2.5">
          <span className="text-xs text-muted-foreground">🎯 我</span>
        </div>
      </div>

      <div className="h-px w-full bg-muted" />

      {/* Combo row */}
      <div className="flex items-center justify-center gap-3 px-6 py-2 md:px-8">
        <span className="text-[13px] font-medium text-muted-foreground">连击</span>
        <div className="flex items-center gap-1.5 rounded-lg bg-red-500 px-3 py-1">
          <Zap className="h-3 w-3 text-white" />
          <span className="text-xs font-bold text-white">×3 🔥</span>
        </div>
        <span className="text-[13px] font-semibold text-red-500">
          多重打击！
        </span>
      </div>

      {/* Hint + Letters */}
      <div className="flex flex-col items-center gap-3 px-6 pb-6 pt-3 md:px-8">
        <span className="text-xs font-medium text-muted-foreground">
          点击字母发射炮弹击碎对手护盾
        </span>
        <div className="flex items-center justify-center gap-2.5">
          {keyboardLetters.map((letter, i) => (
            <button
              key={i}
              type="button"
              className="flex h-12 w-12 items-center justify-center rounded-xl bg-slate-800 shadow-md md:h-14 md:w-14"
            >
              <span className="text-lg font-bold text-white">{letter}</span>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
