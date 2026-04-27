import { api } from '../../../utils/api'
import type { Post } from '../types'

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    statusBarHeight: 20,
    postId: '',
    post: null as Post | null,
    loading: false,
    followed: false,
  },
  onLoad(query: Record<string, string>) {
    const sys = wx.getSystemInfoSync()
    this.setData({
      theme: app.globalData.theme,
      statusBarHeight: sys.statusBarHeight || 20,
      postId: query.id || '',
    })
    if (query.id) this.loadPost()
  },
  onShow() {
    this.setData({ theme: app.globalData.theme })
  },
  goBack() {
    wx.navigateBack()
  },
  async loadPost() {
    this.setData({ loading: true })
    try {
      const post = await api.get<Post>(`/api/posts/${this.data.postId}`)
      this.setData({ post, loading: false })
    } catch (err) {
      this.setData({ loading: false })
      wx.showToast({ title: (err as Error).message || '加载失败', icon: 'none' })
    }
  },
})
