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
  methods: {
    goResume() {
      const session = (this.data as any).recentSession as RecentSession | null
      if (!session) return
      wx.navigateTo({ url: `/pages/games/detail/detail?id=${session.gameId}` })
    },
  },
})
