Component({
  options: {
    addGlobalClass: true,
  },
  methods: {
    goLeaderboard() { wx.switchTab({ url: '/pages/leaderboard/leaderboard' }) },
    goCommunity()   { wx.switchTab({ url: '/pages/me/community/community' }) },
    goGroups()      { wx.navigateTo({ url: '/pages/me/groups/groups' }) },
  },
})
