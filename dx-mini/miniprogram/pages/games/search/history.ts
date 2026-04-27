const KEY = 'dx_recent_searches'
const MAX = 10

export function loadHistory(): string[] {
  const raw = wx.getStorageSync(KEY)
  if (!Array.isArray(raw)) return []
  const arr = raw as unknown[]
  const out: string[] = []
  for (let i = 0; i < arr.length && out.length < MAX; i++) {
    const v = arr[i]
    if (typeof v === 'string' && v.length > 0) out.push(v)
  }
  return out
}

export function pushHistory(term: string): string[] {
  const trimmed = (term || '').trim()
  if (!trimmed) return loadHistory()
  const cur = loadHistory().filter((t) => t !== trimmed)
  const next = [trimmed, ...cur].slice(0, MAX)
  wx.setStorageSync(KEY, next)
  return next
}

export function clearHistory(): void {
  wx.removeStorageSync(KEY)
}
