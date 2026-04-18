import { api } from '../../../utils/api'
import { config } from '../../../utils/config'
import { getToken } from '../../../utils/auth'

interface ProfileData {
  username: string; nickname: string | null; avatarUrl: string | null
  city: string | null; introduction: string | null
}

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    primaryColor: '#0d9488',
    loading: true,
    saving: false,
    uploadingAvatar: false,
    profile: null as ProfileData | null,
    nickname: '',
    city: '',
    introduction: '',
  },
  onLoad() {
    const theme = app.globalData.theme
    this.setData({ theme, primaryColor: theme === 'dark' ? '#14b8a6' : '#0d9488' })
    this.loadProfile()
  },
  async loadProfile() {
    try {
      const profile = await api.get<ProfileData>('/api/user/profile')
      this.setData({
        loading: false,
        profile,
        nickname: profile.nickname ?? '',
        city: profile.city ?? '',
        introduction: profile.introduction ?? '',
      })
    } catch {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },
  onNicknameChange(e: WechatMiniprogram.CustomEvent) {
    this.setData({ nickname: (e.detail as { value: string }).value })
  },
  onCityChange(e: WechatMiniprogram.CustomEvent) {
    this.setData({ city: (e.detail as { value: string }).value })
  },
  onIntroductionChange(e: WechatMiniprogram.CustomEvent) {
    this.setData({ introduction: (e.detail as { value: string }).value })
  },
  chooseAvatar() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const tempPath = res.tempFiles[0].tempFilePath
        this.setData({ uploadingAvatar: true })
        wx.uploadFile({
          url: config.apiBaseUrl + '/api/uploads/images',
          filePath: tempPath,
          name: 'file',
          header: { Authorization: `Bearer ${getToken()}` },
          success: (uploadRes) => {
            const body = JSON.parse(uploadRes.data) as { code: number; data: { url: string } }
            if (body.code === 0) {
              api.put('/api/user/avatar', { avatar_url: body.data.url })
                .then(() => {
                  this.setData({
                    uploadingAvatar: false,
                    profile: { ...this.data.profile!, avatarUrl: body.data.url },
                  })
                  wx.showToast({ title: '头像已更新', icon: 'none' })
                })
                .catch(() => {
                  this.setData({ uploadingAvatar: false })
                  wx.showToast({ title: '更新失败', icon: 'none' })
                })
            } else {
              this.setData({ uploadingAvatar: false })
              wx.showToast({ title: '上传失败', icon: 'none' })
            }
          },
          fail: () => {
            this.setData({ uploadingAvatar: false })
            wx.showToast({ title: '上传失败', icon: 'none' })
          },
        })
      },
    })
  },
  async save() {
    if (this.data.saving) return
    this.setData({ saving: true })
    try {
      await api.put('/api/user/profile', {
        nickname: this.data.nickname || null,
        city: this.data.city || null,
        introduction: this.data.introduction || null,
      })
      this.setData({ saving: false })
      wx.showToast({ title: '保存成功', icon: 'none' })
      wx.navigateBack()
    } catch (err) {
      this.setData({ saving: false })
      wx.showToast({ title: (err as Error).message || '保存失败', icon: 'none' })
    }
  },
})
