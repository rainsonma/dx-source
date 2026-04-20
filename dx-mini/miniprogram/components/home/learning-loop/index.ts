Component({
  options: {
    addGlobalClass: true,
  },
  methods: {
    goUnknown() { wx.navigateTo({ url: '/pages/learn/unknown/unknown' }) },
    goReview()  { wx.navigateTo({ url: '/pages/learn/review/review' }) },
    goMastered(){ wx.navigateTo({ url: '/pages/learn/mastered/mastered' }) },
  },
})
