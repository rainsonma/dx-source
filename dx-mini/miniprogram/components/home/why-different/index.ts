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
      // WeChat animate keyframes take individual transform properties
      // (translateY, scale, etc.) — NOT a `transform:` string. translateY
      // values are in px.
      self.animate(
        '.arrow',
        [
          { offset: 0,   translateY: 0 },
          { offset: 0.4, translateY: -8 },
          { offset: 0.6, translateY: -8 },
          { offset: 1,   translateY: 0 },
        ],
        2000,
        () => {
          if (!self._stopped) self.shakeArrow()
        },
      )
    },
  },
})
