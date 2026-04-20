import { icons } from './icons'

Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    name: { type: String, value: '' },
    size: { type: String, value: '' },
    color: { type: String, value: '' },
    strokeWidth: { type: String, value: '1.25' },
    customStyle: { type: String, value: '' },
    customClass: { type: String, value: '' },
  },
  data: {
    src: '',
    hostStyle: '',
  },
  observers: {
    'name, size, color, strokeWidth, customStyle'(
      name: string,
      size: string,
      color: string,
      strokeWidth: string,
      customStyle: string,
    ) {
      const normalizedSize = /^\d+(\.\d+)?$/.test(size) ? `${size}px` : size
      const hostStyle = `width:${normalizedSize};height:${normalizedSize};${customStyle}`
      const raw = (icons as Record<string, string>)[name] || ''
      const svg = raw
        .replace(/currentColor/g, color || '#000')
        .replace(/stroke-width="2"/g, `stroke-width="${strokeWidth}"`)
      const src = svg ? `data:image/svg+xml;utf8,${encodeURIComponent(svg)}` : ''
      this.setData({ src, hostStyle })
    },
  },
  methods: {
    onClick(e: WechatMiniprogram.CustomEvent) {
      this.triggerEvent('click', e.detail)
    },
  },
})
