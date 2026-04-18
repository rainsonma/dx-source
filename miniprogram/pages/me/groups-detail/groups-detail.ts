import { api } from '../../../utils/api'

interface GroupMember { id: string; username: string; nickname: string | null; role: string; avatarChar: string }
interface GroupDetail {
  id: string; name: string; description: string | null
  ownerName: string; memberCount: number; inviteCode: string; isOwner: boolean
  currentGameId: string | null; currentGameName: string
  isPlaying: boolean; startGameLevelId: string | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    loading: true,
    group: null as GroupDetail | null,
    members: [] as GroupMember[],
    starting: false,
  },
  onLoad(options: { id?: string }) {
    const theme = app.globalData.theme
    this.setData({ theme, primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488' })
    if (options.id) this.loadGroup(options.id)
  },
  async loadGroup(id: string) {
    try {
      const [group, members] = await Promise.all([
        api.get<GroupDetail>(`/api/groups/${id}`),
        api.get<GroupMember[]>(`/api/groups/${id}/members`),
      ])
      const membersWithChar = members.map(m => ({
        ...m,
        avatarChar: (m.nickname || m.username).charAt(0),
      }))
      this.setData({ loading: false, group, members: membersWithChar })
      wx.setNavigationBarTitle({ title: group.name })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  copyInviteCode() {
    wx.setClipboardData({ data: this.data.group!.inviteCode })
    wx.showToast({ title: '邀请码已复制', icon: 'none' })
  },
  async startGroupGame() {
    if (!this.data.group || !this.data.group.currentGameId || this.data.starting) return
    this.setData({ starting: true })
    try {
      await api.post(`/api/groups/${this.data.group.id}/start-game`, {})
      wx.showToast({ title: '游戏已开始', icon: 'none' })
      this.setData({ starting: false })
    } catch (err) {
      this.setData({ starting: false })
      wx.showToast({ title: (err as Error).message || '开始失败', icon: 'none' })
    }
  },
})
