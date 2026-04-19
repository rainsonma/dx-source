import { config } from './config'

type EventHandler = (payload: unknown) => void

let socket: WechatMiniprogram.SocketTask | null = null
let isOpen = false
let isAuthed = false
const pendingSends: string[] = []
const handlers = new Map<string, EventHandler[]>()

function randomId(): string {
  return Math.random().toString(36).slice(2) + Date.now().toString(36)
}

function flushPending(): void {
  if (!socket || !isAuthed) return
  while (pendingSends.length > 0) {
    const data = pendingSends.shift() as string
    socket.send({ data })
  }
}

type Incoming = {
  op?: string
  topic?: string
  type?: string
  data?: unknown
  id?: string
  ok?: boolean
  code?: number
  message?: string
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
      // First frame MUST be the auth envelope. Server rejects anything else
      // and closes. pendingSends stays buffered until auth_success arrives.
      if (socket) {
        socket.send({ data: JSON.stringify({ op: 'auth', token }) })
      }
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
        const msg = JSON.parse(data as string) as Incoming
        // Auth-lifecycle events ride on the same Envelope shape:
        //   {op:"event", type:"auth_success" | "auth_failed" | "session_replaced"}
        if (msg.op === 'event') {
          if (msg.type === 'auth_success') {
            isAuthed = true
            flushPending()
            return
          }
          if (msg.type === 'auth_failed') {
            isAuthed = false
            return
          }
          if (msg.type === 'session_replaced') {
            isAuthed = false
            // fall through so any registered handler fires too
          }
          const key = msg.type || ''
          const cbs = handlers.get(key) || []
          cbs.forEach(cb => cb(msg.data))
        }
        // Ignore op:ack / op:error for now — the mini program doesn't use
        // subscribe ACKs; errors are surfaced via close codes.
      } catch {
        // ignore malformed messages
      }
    })
  },
  subscribe(topic: string): void {
    // Matches the server's realtime.Envelope schema.
    const data = JSON.stringify({ op: 'subscribe', topic, id: randomId() })
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
