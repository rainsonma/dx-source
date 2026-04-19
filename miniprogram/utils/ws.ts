import { config } from './config'

type EventHandler = (payload: unknown) => void

let socket: WechatMiniprogram.SocketTask | null = null
let isOpen = false
let isAuthed = false
const pendingSends: string[] = []
const handlers = new Map<string, EventHandler[]>()

function flushPending(): void {
  if (!socket || !isAuthed) return
  while (pendingSends.length > 0) {
    const data = pendingSends.shift() as string
    socket.send({ data })
  }
}

export const ws = {
  connect(token: string): void {
    isOpen = false
    isAuthed = false
    pendingSends.length = 0
    const wsUrl = config.apiBaseUrl.replace(/^http/, 'ws') + '/api/ws'
    socket = wx.connectSocket({
      url: wsUrl,
      success() {},
      fail() {},
    })
    socket.onOpen(() => {
      isOpen = true
      // First frame MUST be the auth envelope — the server rejects anything
      // else and closes. pendingSends stays buffered until auth_success.
      socket?.send({ data: JSON.stringify({ type: 'auth', token }) })
    })
    socket.onClose(() => {
      isOpen = false
      isAuthed = false
    })
    socket.onError(() => {
      isOpen = false
      isAuthed = false
    })
    socket.onMessage(({ data }) => {
      try {
        const msg = JSON.parse(data as string) as { event?: string; payload?: unknown }
        if (msg.event === 'auth_success') {
          isAuthed = true
          flushPending()
          return
        }
        if (msg.event === 'auth_failed' || msg.event === 'session_replaced') {
          isAuthed = false
          // Fall through so registered handlers still get session_replaced.
        }
        const key = msg.event ?? ''
        const cbs = handlers.get(key) ?? []
        cbs.forEach(cb => cb(msg.payload))
      } catch {
        // ignore malformed messages
      }
    })
  },
  subscribe(topic: string): void {
    const data = JSON.stringify({ type: 'subscribe', topic })
    if (socket && isAuthed) {
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
    isAuthed = false
    pendingSends.length = 0
    handlers.clear()
  },
}
