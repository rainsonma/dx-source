import { api } from '../../../utils/api'

interface GameLevelData { id: string; name: string; order: number }
interface GameDetailData {
  id: string; name: string; description: string | null; mode: string
  coverUrl: string | null; author: string | null; categoryName: string | null
  pressName: string | null; levels: GameLevelData[]; levelCount: number
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    game: null as GameDetailData | null,
    favorited: false,
  },
  onLoad(options: { id?: string }) {
    this.setData({ theme: app.globalData.theme })
    if (options.id) this.loadGame(options.id)
  },
  async loadGame(id: string) {
    try {
      const [game, favRes] = await Promise.all([
        api.get<GameDetailData>(`/api/games/${id}`),
        api.get<{ favorited: boolean }>(`/api/games/${id}/favorited`),
      ])
      this.setData({ loading: false, game, favorited: favRes.favorited })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  async toggleFavorite() {
    if (!this.data.game) return
    try {
      const res = await api.post<{ favorited: boolean }>('/api/favorites/toggle', { gameId: this.data.game.id })
      this.setData({ favorited: res.favorited })
      wx.showToast({ title: res.favorited ? '已收藏' : '已取消收藏', icon: 'none' })
    } catch {
      wx.showToast({ title: '操作失败', icon: 'none' })
    }
  },
  startLevel(e: WechatMiniprogram.TouchEvent) {
    const levelId = e.currentTarget.dataset['levelId'] as string
    const gameId = this.data.game!.id
    wx.navigateTo({ url: `/pages/games/play/play?gameId=${gameId}&levelId=${levelId}&degree=normal` })
  },
})
