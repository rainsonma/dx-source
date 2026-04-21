interface RecentSession {
  gameId: string
  gameName: string
  completedLevels: number
}

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
      self.startAnimations()
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

    // Kick off every per-illustration loop in parallel. Each loop calls
    // `this.animate` and schedules itself again via setTimeout in the
    // callback — staying staggered across cycles after the initial offset.
    startAnimations() {
      const self = this as any
      // 连词成句: 4 word pills bounce with 120ms stagger
      ;[1, 2, 3, 4].forEach((n, i) => {
        setTimeout(() => self.loopSentenceTile(n), i * 120)
      })
      // 词汇配对: 3 pair rows pulse with 220ms stagger
      ;[1, 2, 3].forEach((n, i) => {
        setTimeout(() => self.loopMatchPair(n), i * 220)
      })
      // 词汇消消乐: 4 pairs scale+pulse with 320ms stagger
      ;[0, 1, 2, 3].forEach((p, i) => {
        setTimeout(() => self.loopElimPair(p), i * 320)
      })
      // 词汇对轰: 3 concurrent loops on bars + star
      self.loopBattleTeal()
      setTimeout(() => self.loopBattleRed(), 200)
      setTimeout(() => self.loopBattleStar(), 400)
    },

    loopSentenceTile(n: number) {
      const self = this as any
      if (self._stopped) return
      self.animate(
        `.illus-tiles .tile-${n}`,
        [
          { offset: 0,   translateY: 0 },
          { offset: 0.5, translateY: -4 },
          { offset: 1,   translateY: 0 },
        ],
        1200,
        () => {
          setTimeout(() => {
            if (!self._stopped) self.loopSentenceTile(n)
          }, 600)
        },
      )
    },

    loopMatchPair(n: number) {
      const self = this as any
      if (self._stopped) return
      // translateY + opacity pulse (both verified working on <text> in this
      // WeChat version). Pair tiles animate in sync to visually connect them.
      const frames = [
        { offset: 0,   translateY: 0,  opacity: 1 },
        { offset: 0.5, translateY: -5, opacity: 0.65 },
        { offset: 1,   translateY: 0,  opacity: 1 },
      ]
      // Pair n ↔ tiles m(2n-1) and m(2n): p1=[m1,m2], p2=[m3,m4], p3=[m5,m6]
      const a = 2 * n - 1
      const b = 2 * n
      self.animate(`.illus-match .m${a}`, frames, 1400)
      self.animate(`.illus-match .m${b}`, frames, 1400, () => {
        setTimeout(() => {
          if (!self._stopped) self.loopMatchPair(n)
        }, 800)
      })
    },

    loopElimPair(p: number) {
      const self = this as any
      if (self._stopped) return
      const frames = [
        { offset: 0,    translateY: 0,  opacity: 1 },
        { offset: 0.35, translateY: -6, opacity: 0.55 },
        { offset: 0.55, translateY: 2,  opacity: 0.85 },
        { offset: 1,    translateY: 0,  opacity: 1 },
      ]
      // Pair map (see WXML comment): p0=[e1,e5], p1=[e2,e7], p2=[e3,e6], p3=[e4,e8]
      const pairs: Record<number, [number, number]> = {
        0: [1, 5], 1: [2, 7], 2: [3, 6], 3: [4, 8],
      }
      const [a, b] = pairs[p]
      self.animate(`.illus-grid .e${a}`, frames, 2200)
      self.animate(`.illus-grid .e${b}`, frames, 2200, () => {
        setTimeout(() => {
          if (!self._stopped) self.loopElimPair(p)
        }, 600)
      })
    },

    loopBattleTeal() {
      const self = this as any
      if (self._stopped) return
      self.animate(
        '.illus-battle .bar-teal',
        [
          { offset: 0,   scaleX: 1 },
          { offset: 0.5, scaleX: 0.9 },
          { offset: 1,   scaleX: 1 },
        ],
        1400,
        () => {
          if (!self._stopped) self.loopBattleTeal()
        },
      )
    },

    loopBattleRed() {
      const self = this as any
      if (self._stopped) return
      self.animate(
        '.illus-battle .bar-red',
        [
          { offset: 0,   scaleX: 1 },
          { offset: 0.5, scaleX: 0.6 },
          { offset: 0.8, scaleX: 0.6 },
          { offset: 1,   scaleX: 1 },
        ],
        2200,
        () => {
          if (!self._stopped) self.loopBattleRed()
        },
      )
    },

    loopBattleStar() {
      const self = this as any
      if (self._stopped) return
      self.animate(
        '.illus-battle .battle-star',
        [
          { offset: 0,   translateX: 0,  opacity: 0 },
          { offset: 0.5, translateX: 50, opacity: 1 },
          { offset: 1,   translateX: 100, opacity: 0 },
        ],
        1400,
        () => {
          setTimeout(() => {
            if (!self._stopped) self.loopBattleStar()
          }, 600)
        },
      )
    },
  },
})
