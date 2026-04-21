Component({
  options: {
    addGlobalClass: true,
  },
  methods: {
    goLeaderboard() { wx.navigateTo({ url: '/pages/leaderboard/leaderboard' }) },
    goCommunity()   { wx.navigateTo({ url: '/pages/me/community/community' }) },
    goGroups()      { wx.navigateTo({ url: '/pages/me/groups/groups' }) },
  },
})
