interface TabItem {
  icon: string
  activeIcon: string
  text: string
  path: string
}

Component({
  data: {
    active: 0,
    theme: 'light' as 'light' | 'dark',
    tabs: [
      { icon: 'wap-home-o', activeIcon: 'wap-home', text: '首页', path: '/pages/home/home' },
      { icon: 'column', activeIcon: 'column', text: '课程', path: '/pages/games/games' },
      { icon: 'chart-trending-o', activeIcon: 'chart-trending-o', text: '排行榜', path: '/pages/leaderboard/leaderboard' },
      { icon: 'records', activeIcon: 'records', text: '学习', path: '/pages/learn/learn' },
      { icon: 'contact', activeIcon: 'contact', text: '我的', path: '/pages/me/me' },
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
