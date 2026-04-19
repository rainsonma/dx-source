Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    name: { type: String, value: '' },
    size: { type: String, value: '' },
    color: { type: String, value: '' },
    customStyle: { type: String, value: '' },
    customClass: { type: String, value: '' },
  },
  methods: {
    onClick(e: WechatMiniprogram.CustomEvent) {
      this.triggerEvent('click', e.detail)
    },
  },
})
