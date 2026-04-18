import { config } from './config'

type EventHandler = (payload: unknown) => void

let socket: WechatMiniprogram.SocketTask | null = null
const handlers = new Map<string, EventHandler[]>()

export const ws = {
  connect(token: string): void {
    const wsUrl = config.apiBaseUrl.replace(/^http/, 'ws') + '/api/ws'
    socket = wx.connectSocket({
      url: wsUrl,
      header: { Authorization: `Bearer ${token}` },
      success() {},
      fail() {},
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
    if (socket) {
      socket.send({ data: JSON.stringify({ type: 'subscribe', topic }) })
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
    handlers.clear()
  },
}
