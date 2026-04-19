import { api } from '../../../utils/api'

interface GameCardData {
  id: string; name: string; mode: string
  coverUrl: string | null; categoryName: string | null; levelCount: number
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    games: [] as GameCardData[],
  },
  onLoad() {
    this.setData({ theme: app.globalData.theme })
    this.loadFavorites()
  },
  onPullDownRefresh() {
    this.loadFavorites().then(() => wx.stopPullDownRefresh())
  },
  async loadFavorites() {
    this.setData({ loading: true })
    try {
      const games = await api.get<GameCardData[]>('/api/favorites')
      this.setData({ loading: false, games })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  async unfavorite(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string
    try {
      await api.post('/api/favorites/toggle', { gameId: id })
      this.setData({ games: this.data.games.filter(g => g.id !== id) })
      wx.showToast({ title: '已取消收藏', icon: 'none' })
    } catch {
      wx.showToast({ title: '操作失败', icon: 'none' })
    }
  },
  goDetail(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset['id'] as string
    wx.navigateTo({ url: `/pages/games/detail/detail?id=${id}` })
  },
})
