export function formatDate(iso: string | null | undefined): string {
  if (!iso) return ''
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export function formatRelativeDate(iso: string | null | undefined): string {
  if (!iso) return ''
  const diff = Date.now() - new Date(iso).getTime()
  const days = Math.floor(diff / 86400000)
  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days}天前`
  if (days < 30) return `${Math.floor(days / 7)}周前`
  if (days < 365) return `${Math.floor(days / 30)}个月前`
  return `${Math.floor(days / 365)}年前`
}

export function formatNumber(n: number): string {
  if (n >= 10000) return `${(n / 10000).toFixed(1)}万`
  return String(n)
}

export function gradeLabel(grade: string): string {
  const map: Record<string, string> = {
    free: '免费',
    month: '月度会员',
    season: '季度会员',
    year: '年度会员',
    lifetime: '终身会员',
  }
  return map[grade] || grade
}

// Days between `now` and the ISO date string. Returns 0 for null/expired.
// Used by the membership section to render "续费 · 还剩 N 天" on the
// user's current tier button.
export function daysUntil(isoDate: string | null | undefined): number {
  if (!isoDate) return 0
  const target = new Date(isoDate).getTime()
  const now = Date.now()
  return Math.max(0, Math.ceil((target - now) / 86400000))
}
