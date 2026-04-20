Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    streak: {
      type: Number,
      value: 0,
    },
  },
  data: {
    heatCells: [false, false, false, false, false, false, false],
  },
  observers: {
    streak(streak: number) {
      const filled = Math.max(0, Math.min(streak || 0, 7))
      const cells = [false, false, false, false, false, false, false]
      for (let i = 0; i < filled; i++) cells[i] = true
      this.setData({ heatCells: cells })
    },
  },
  methods: {
    goLeaderboard() { wx.navigateTo({ url: '/pages/leaderboard/leaderboard' }) },
    goCommunity()   { wx.navigateTo({ url: '/pages/me/community/community' }) },
    goGroups()      { wx.navigateTo({ url: '/pages/me/groups/groups' }) },
  },
})
