Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    unknownTotal: {
      type: null,
      value: null,
    },
    reviewPending: {
      type: null,
      value: null,
    },
    masterTotal: {
      type: null,
      value: null,
    },
  },
  methods: {
    goUnknown() { wx.navigateTo({ url: '/pages/learn/unknown/unknown' }) },
    goReview()  { wx.navigateTo({ url: '/pages/learn/review/review' }) },
    goMastered(){ wx.navigateTo({ url: '/pages/learn/mastered/mastered' }) },
  },
})
