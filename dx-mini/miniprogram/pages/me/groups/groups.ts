import { api, PaginatedData } from '../../../utils/api'

interface GroupListItem {
  id: string; name: string; description: string | null
  ownerName: string; memberCount: number; inviteCode: string
  isMember: boolean; isOwner: boolean
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    loading: true,
    groups: [] as GroupListItem[],
    nextCursor: '',
    hasMore: false,
    showCreateDialog: false,
    showJoinDialog: false,
    createName: '',
    joinCode: '',
    creating: false,
    joining: false,
    statusBarHeight: 20,
  },
  onLoad() {
    const sys = wx.getSystemInfoSync()
    const statusBarHeight = sys.statusBarHeight || 20
    const theme = app.globalData.theme
    this.setData({ theme, primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488', statusBarHeight })
    this.loadGroups(true)
  },
  goBack() { wx.navigateBack() },
  onPullDownRefresh() {
    this.loadGroups(true).then(() => wx.stopPullDownRefresh())
  },
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) this.loadGroups(false)
  },
  async loadGroups(reset: boolean) {
    this.setData({ loading: true })
    const cursor = reset ? '' : this.data.nextCursor
    const qs = cursor ? `?cursor=${cursor}` : ''
    try {
      const res = await api.get<PaginatedData<GroupListItem>>(`/api/groups${qs}`)
      this.setData({
        loading: false,
        groups: reset ? res.items : [...this.data.groups, ...res.items],
        nextCursor: res.nextCursor,
        hasMore: res.hasMore,
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  goDetail(e: WechatMiniprogram.TouchEvent) {
    const id = (e.currentTarget.dataset as { id: string }).id
    wx.navigateTo({ url: `/pages/me/groups-detail/groups-detail?id=${id}` })
  },
  openCreateDialog() { this.setData({ showCreateDialog: true }) },
  closeCreateDialog() { this.setData({ showCreateDialog: false, createName: '' }) },
  openJoinDialog() { this.setData({ showJoinDialog: true }) },
  closeJoinDialog() { this.setData({ showJoinDialog: false, joinCode: '' }) },
  onCreateNameChange(e: WechatMiniprogram.CustomEvent) {
    this.setData({ createName: (e.detail as { value: string }).value })
  },
  onJoinCodeChange(e: WechatMiniprogram.CustomEvent) {
    this.setData({ joinCode: (e.detail as { value: string }).value })
  },
  async createGroup() {
    if (!this.data.createName.trim() || this.data.creating) return
    this.setData({ creating: true })
    try {
      await api.post('/api/groups', { name: this.data.createName.trim() })
      this.setData({ showCreateDialog: false, createName: '', creating: false })
      this.loadGroups(true)
      wx.showToast({ title: '创建成功', icon: 'none' })
    } catch (err) {
      this.setData({ creating: false })
      wx.showToast({ title: (err as Error).message || '创建失败', icon: 'none' })
    }
  },
  async joinGroup() {
    if (!this.data.joinCode.trim() || this.data.joining) return
    this.setData({ joining: true })
    try {
      await api.post(`/api/groups/join/${this.data.joinCode.trim()}`, {})
      this.setData({ showJoinDialog: false, joinCode: '', joining: false })
      this.loadGroups(true)
      wx.showToast({ title: '加入成功', icon: 'none' })
    } catch (err) {
      this.setData({ joining: false })
      wx.showToast({ title: (err as Error).message || '加入失败', icon: 'none' })
    }
  },
})
