interface RecentSession {
  gameId: string
  gameName: string
  completedLevels: number
}

// Per-card spotlight durations (ms). Each is long enough for that card's
// internal animations to run once and settle. Total rotation ≈ 10.6 s.
const SPOTLIGHTS_MS: number[] = [2200, 2400, 3600, 2600]

Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    recentSession: {
      type: null,
      value: null,
    },
  },
  lifetimes: {
    attached() {
      const self = this as any
      self._stopped = false
      self._cardIdx = 0
      self.rotate()
    },
    detached() {
      (this as any)._stopped = true
    },
  },
  methods: {
    goResume() {
      const session = (this.data as any).recentSession as RecentSession | null
      if (!session) return
      wx.navigateTo({ url: `/pages/games/detail/detail?id=${session.gameId}` })
    },

    // Rotation coordinator: spotlights one card, schedules next.
    rotate() {
      const self = this as any
      if (self._stopped) return
      const i = self._cardIdx
      if (i === 0) self.playSentence()
      else if (i === 1) self.playMatch()
      else if (i === 2) self.playElim()
      else if (i === 3) self.playBattle()
      setTimeout(() => {
        if (self._stopped) return
        self._cardIdx = (self._cardIdx + 1) % 4
        self.rotate()
      }, SPOTLIGHTS_MS[i])
    },

    // 连词成句: all 4 tiles bounce, 120ms stagger, once per spotlight
    playSentence() {
      const self = this as any
      const frames = [
        { offset: 0,   translateY: 0 },
        { offset: 0.5, translateY: -4 },
        { offset: 1,   translateY: 0 },
      ]
      ;[1, 2, 3, 4].forEach((n, i) => {
        setTimeout(() => {
          if (self._stopped || self._cardIdx !== 0) return
          self.animate(`.illus-tiles .tile-${n}`, frames, 1200)
        }, i * 120)
      })
    },

    // 词汇配对: 3 pairs pulse in sequence, 220ms stagger
    playMatch() {
      const self = this as any
      const frames = [
        { offset: 0,   translateY: 0,  opacity: 1 },
        { offset: 0.5, translateY: -5, opacity: 0.65 },
        { offset: 1,   translateY: 0,  opacity: 1 },
      ]
      ;[1, 2, 3].forEach((n, i) => {
        setTimeout(() => {
          if (self._stopped || self._cardIdx !== 1) return
          const a = 2 * n - 1
          const b = 2 * n
          self.animate(`.illus-match .m${a}`, frames, 1400)
          self.animate(`.illus-match .m${b}`, frames, 1400)
        }, i * 220)
      })
    },

    // 词汇消消乐: 4 pairs bounce-scale, 320ms stagger
    playElim() {
      const self = this as any
      const frames = [
        { offset: 0,    translateY: 0,  opacity: 1 },
        { offset: 0.35, translateY: -6, opacity: 0.55 },
        { offset: 0.55, translateY: 2,  opacity: 0.85 },
        { offset: 1,    translateY: 0,  opacity: 1 },
      ]
      const pairs: Record<number, [number, number]> = {
        0: [1, 5], 1: [2, 7], 2: [3, 6], 3: [4, 8],
      }
      ;[0, 1, 2, 3].forEach((p, i) => {
        setTimeout(() => {
          if (self._stopped || self._cardIdx !== 2) return
          const [a, b] = pairs[p]
          self.animate(`.illus-grid .e${a}`, frames, 2200)
          self.animate(`.illus-grid .e${b}`, frames, 2200)
        }, i * 320)
      })
    },

    // 词汇对轰: 3 concurrent bar + star animations
    playBattle() {
      const self = this as any
      if (self._stopped) return
      self.animate('.illus-battle .bar-teal', [
        { offset: 0,   scaleX: 1 },
        { offset: 0.5, scaleX: 0.9 },
        { offset: 1,   scaleX: 1 },
      ], 1400)
      setTimeout(() => {
        if (self._stopped || self._cardIdx !== 3) return
        self.animate('.illus-battle .bar-red', [
          { offset: 0,   scaleX: 1 },
          { offset: 0.5, scaleX: 0.6 },
          { offset: 0.8, scaleX: 0.6 },
          { offset: 1,   scaleX: 1 },
        ], 2200)
      }, 200)
      setTimeout(() => {
        if (self._stopped || self._cardIdx !== 3) return
        self.animate('.illus-battle .battle-star', [
          { offset: 0,   translateX: 0,   opacity: 0 },
          { offset: 0.5, translateX: 50,  opacity: 1 },
          { offset: 1,   translateX: 100, opacity: 0 },
        ], 1400)
      }, 400)
    },
  },
})
