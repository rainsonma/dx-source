Component({
  options: {
    addGlobalClass: true,
    multipleSlots: false,
  },
  properties: {
    theme: { type: String, value: 'light' },
    pinned: { type: Boolean, value: true },
    revealed: { type: Boolean, value: true },
    placeholder: { type: String, value: '搜索课程' },
    mode: { type: String, value: 'launcher' },
    showCancel: { type: Boolean, value: false },
  },
  data: {
    statusBarHeight: 20,
    rowHeight: 40,
    pillRight: 102,
  },
  lifetimes: {
    attached() {
      const sys = wx.getSystemInfoSync()
      const cap = wx.getMenuButtonBoundingClientRect()
      const statusBarHeight = sys.statusBarHeight || 20
      const rowHeight = Math.max(40, (cap.bottom - statusBarHeight) + 8)
      const pillRight = Math.max(102, sys.windowWidth - cap.left + 8)
      this.setData({ statusBarHeight, rowHeight, pillRight })
    },
  },
  methods: {
    onTap() {
      this.triggerEvent('tap', {})
    },
    onCancel() {
      this.triggerEvent('cancel', {})
    },
  },
})
