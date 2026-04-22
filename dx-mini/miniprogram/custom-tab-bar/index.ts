interface TabItem {
  icon: string
  text: string
  path: string
}

Component({
  data: {
    active: 0,
    theme: 'light' as 'light' | 'dark',
    tabs: [
      { icon: 'home',          text: '首页',   path: '/pages/home/home' },
      { icon: 'book-text',     text: '课程',   path: '/pages/games/games' },
      { icon: 'trophy',        text: '排行榜', path: '/pages/leaderboard/leaderboard' },
      { icon: 'message-square', text: '社区',   path: '/pages/me/community/community' },
      { icon: 'user',          text: '我的',   path: '/pages/me/me' },
    ] as TabItem[],
  },
  lifetimes: {
    attached() {
      const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()
      this.setData({ theme: app.globalData.theme })
    },
  },
  methods: {
    switchTab(e: WechatMiniprogram.TouchEvent) {
      const path = e.currentTarget.dataset['path'] as string
      wx.switchTab({ url: path })
    },
  },
})
