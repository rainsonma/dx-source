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
      // wired in Task 12
    },
    onRemoveImage() {
      this.setData({ imageUrl: '' })
    },
  },
})
