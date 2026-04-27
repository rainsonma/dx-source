import { config } from '../../../../utils/config'
import { getToken } from '../../../../utils/auth'

Component({
  options: { addGlobalClass: true },
  properties: {
    show: { type: Boolean, value: false },
    theme: { type: String, value: 'light' },
  },
  data: {
    content: '',
    tagInput: '',
    tags: [] as string[],
    imageUrl: '',
    uploading: false,
  },
  methods: {
    onContentInput(e: WechatMiniprogram.Input) {
      this.setData({ content: (e.detail as { value: string }).value })
    },
    onTagInput(e: WechatMiniprogram.Input) {
      this.setData({ tagInput: (e.detail as { value: string }).value })
    },
    onTagConfirm() {
      const raw = (this.data as { tagInput: string }).tagInput.trim().replace(/^#/, '')
      if (!raw) return
      const tags = (this.data as { tags: string[] }).tags
      if (tags.length >= 5) {
        wx.showToast({ title: '最多5个标签', icon: 'none' })
        return
      }
      if (raw.length > 20) {
        wx.showToast({ title: '标签不超过20字', icon: 'none' })
        return
      }
      if (tags.indexOf(raw) >= 0) {
        this.setData({ tagInput: '' })
        return
      }
      this.setData({ tags: tags.concat([raw]), tagInput: '' })
    },
    onTagRemove(e: WechatMiniprogram.TouchEvent) {
      const tag = e.currentTarget.dataset['tag'] as string
      const tags = (this.data as { tags: string[] }).tags.filter((t) => t !== tag)
      this.setData({ tags })
    },
    onClose() {
      this.triggerEvent('close')
    },
    onSubmit() {
      // wired in Task 13
    },
    onPickImage() {
      const self = this
      wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        sizeType: ['compressed'],
        success(res) {
          const file = res.tempFiles[0]
          if (file.size > 2 * 1024 * 1024) {
            wx.showToast({ title: '图片不超过 2MB', icon: 'none' })
            return
          }
          const lower = file.tempFilePath.toLowerCase()
          if (!/\.(jpg|jpeg|png)$/.test(lower)) {
            wx.showToast({ title: '仅支持 JPG/PNG', icon: 'none' })
            return
          }
          self.setData({ uploading: true, imageUrl: file.tempFilePath })
          wx.uploadFile({
            url: config.apiBaseUrl + '/api/uploads/images',
            filePath: file.tempFilePath,
            name: 'file',
            formData: { role: 'post-image' },
            header: { Authorization: 'Bearer ' + getToken() },
            success(uploadRes) {
              try {
                const body = JSON.parse(uploadRes.data) as { code: number; message: string; data: { url: string } }
                if (body.code === 0) {
                  self.setData({ imageUrl: body.data.url, uploading: false })
                } else {
                  self.setData({ imageUrl: '', uploading: false })
                  wx.showToast({ title: body.message || '上传失败', icon: 'none' })
                }
              } catch {
                self.setData({ imageUrl: '', uploading: false })
                wx.showToast({ title: '上传失败', icon: 'none' })
              }
            },
            fail() {
              self.setData({ imageUrl: '', uploading: false })
              wx.showToast({ title: '上传失败', icon: 'none' })
            },
          })
        },
      })
    },
    onRemoveImage() {
      this.setData({ imageUrl: '' })
    },
  },
})
