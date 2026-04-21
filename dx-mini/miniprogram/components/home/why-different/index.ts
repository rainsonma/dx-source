Component({
  options: {
    addGlobalClass: true,
  },
  lifetimes: {
    attached() {
      const self = this as any
      self._stopped = false
      self.shakeArrow()
    },
    detached() {
      const self = this as any
      self._stopped = true
    },
  },
  methods: {
    // Gentle, continuous up/down shake on all three down-arrows between
    // before/after rows. Uses WeChat's `this.animate` (advanced animation
    // API) rather than CSS @keyframes because this hook gives us a real
    // JS callback we can loop on, which is more forgiving across WeChat
    // renderer versions than CSS `animation-iteration-count: infinite`.
    shakeArrow() {
      const self = this as any
      if (self._stopped) return
      self.animate(
        '.arrow',
        [
          { offset: 0,   transform: 'translateY(0)' },
          { offset: 0.4, transform: 'translateY(-5rpx)' },
          { offset: 0.6, transform: 'translateY(-5rpx)' },
          { offset: 1,   transform: 'translateY(0)' },
        ],
        1200,
        () => {
          if (!self._stopped) self.shakeArrow()
        },
      )
    },
  },
})
