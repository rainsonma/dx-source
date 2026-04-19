import { config } from './config'

type EventHandler = (payload: unknown) => void

let socket: WechatMiniprogram.SocketTask | null = null
let isOpen = false
const pendingSends: string[] = []
const handlers = new Map<string, EventHandler[]>()

function flushPending(): void {
  if (!socket || !isOpen) return
  while (pendingSends.length > 0) {
    const data = pendingSends.shift() as string
    socket.send({ data })
  }
}

export const ws = {
  connect(token: string): void {
    isOpen = false
    pendingSends.length = 0
    const base = config.apiBaseUrl.replace(/^http/, 'ws')
    const wsUrl = `${base}/api/ws?token=${encodeURIComponent(token)}`
    socket = wx.connectSocket({
      url: wsUrl,
      header: { Authorization: `Bearer ${token}` },
      success() {},
      fail() {},
    })
    socket.onOpen(() => {
      isOpen = true
      flushPending()
    })
    socket.onClose(() => {
      isOpen = false
    })
    socket.onError(() => {
      isOpen = false
    })
    socket.onMessage(({ data }) => {
      try {
        const msg = JSON.parse(data as string) as { event: string; payload: unknown }
        const cbs = handlers.get(msg.event) ?? []
        cbs.forEach(cb => cb(msg.payload))
      } catch {
        // ignore malformed messages
      }
    })
  },
  subscribe(topic: string): void {
    const data = JSON.stringify({ type: 'subscribe', topic })
    if (socket && isOpen) {
      socket.send({ data })
    } else {
      pendingSends.push(data)
    }
  },
  on(event: string, cb: EventHandler): void {
    if (!handlers.has(event)) handlers.set(event, [])
    handlers.get(event)!.push(cb)
  },
  disconnect(): void {
    if (socket) {
      socket.close({})
    }
    socket = null
    isOpen = false
    pendingSends.length = 0
    handlers.clear()
  },
}
