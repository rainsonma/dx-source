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
      self.animate(
        `.illus-match .pair-${n}`,
        [
          { offset: 0,   backgroundColor: '#ffffff' },
          { offset: 0.5, backgroundColor: '#ccfbf1' },
          { offset: 1,   backgroundColor: '#ffffff' },
        ],
        1400,
        () => {
          setTimeout(() => {
            if (!self._stopped) self.loopMatchPair(n)
          }, 800)
        },
      )
    },

    loopElimPair(p: number) {
      const self = this as any
      if (self._stopped) return
      self.animate(
        `.illus-grid .cell-p${p}`,
        [
          { offset: 0,    backgroundColor: '#ffffff', scale: 1 },
          { offset: 0.35, backgroundColor: '#fce7f3', scale: 1.06 },
          { offset: 0.55, backgroundColor: '#f1f5f9', scale: 0.95 },
          { offset: 1,    backgroundColor: '#ffffff', scale: 1 },
        ],
        2400,
        () => {
          setTimeout(() => {
            if (!self._stopped) self.loopElimPair(p)
          }, 600)
        },
      )
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
