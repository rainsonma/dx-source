const LINES: string[] = [
  `› negotiate · 谈判 — Let's negotiate the salary.`,
  `› résumé · 简历 — Please send me your résumé.`,
  `› confident · 自信 — Stay confident during the interview.`,
  `› leverage · 优势 — Leverage your strengths.`,
  `› follow up · 跟进 — I'll follow up next week.`,
]

const CHAR_MS = 30
const LINE_GAP_MS = 260

Component({
  options: {
    addGlobalClass: true,
  },
  data: {
    lines: [] as Array<Array<{ ch: string; d: number }>>,
  },
  lifetimes: {
    attached() {
      // Pre-compute each character's animation-delay once; one setData on
      // mount. After that, the typewriter is pure CSS — no further setData
      // ticks, no paint pressure on the rest of the page.
      let cum = 0
      const lines: Array<Array<{ ch: string; d: number }>> = []
      for (const line of LINES) {
        const chars: Array<{ ch: string; d: number }> = []
        for (const ch of line) {
          chars.push({ ch, d: cum })
          cum += CHAR_MS
        }
        cum += LINE_GAP_MS
        lines.push(chars)
      }
      this.setData({ lines })
    },
  },
})
