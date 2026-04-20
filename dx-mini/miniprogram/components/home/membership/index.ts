import { daysUntil } from '../../../utils/format'

type Grade = 'free' | 'month' | 'season' | 'year' | 'lifetime'

interface ButtonState {
  label: string
  disabled: boolean
  celebratory: boolean
}

interface TierStates {
  free: ButtonState
  month: ButtonState
  year: ButtonState
  lifetime: ButtonState
  currentExpiresIn: number  // 0 if not on a time-bounded tier
  showExpiryLine: boolean    // true for season/year, false otherwise
}

function computeStates(grade: Grade, vipDueAt: string | null): TierStates {
  const info: ButtonState = { label: '默认权益', disabled: true, celebratory: false }
  const open = (l: string): ButtonState => ({ label: l, disabled: false, celebratory: false })
  const off  = (l: string): ButtonState => ({ label: l, disabled: true,  celebratory: false })
  const done = (l: string): ButtonState => ({ label: l, disabled: true,  celebratory: true })

  const n = daysUntil(vipDueAt)

  if (grade === 'free') {
    return {
      free: info,
      month: open('立即开通'),
      year: open('立即开通'),
      lifetime: open('立即开通'),
      currentExpiresIn: 0,
      showExpiryLine: false,
    }
  }
  if (grade === 'month') {
    return {
      free: info,
      month: open(n > 0 ? `续费 · 还剩 ${n} 天` : '立即开通'),
      year: open('立即开通'),
      lifetime: open('立即开通'),
      currentExpiresIn: n,
      showExpiryLine: false,
    }
  }
  if (grade === 'season') {
    return {
      free: info,
      month: off('已包含'),
      year: open('升级到年度'),
      lifetime: open('升级到终身'),
      currentExpiresIn: n,
      showExpiryLine: true,
    }
  }
  if (grade === 'year') {
    return {
      free: info,
      month: off('已包含'),
      year: open(n > 0 ? `续费 · 还剩 ${n} 天` : '立即开通'),
      lifetime: open('升级到终身'),
      currentExpiresIn: n,
      showExpiryLine: true,
    }
  }
  if (grade === 'lifetime') {
    return {
      free: info,
      month: off('已包含'),
      year: off('已包含'),
      lifetime: done('✨ 已开通'),
      currentExpiresIn: 0,
      showExpiryLine: false,
    }
  }
  // Unknown grade: treat as free.
  return {
    free: info,
    month: open('立即开通'),
    year: open('立即开通'),
    lifetime: open('立即开通'),
    currentExpiresIn: 0,
    showExpiryLine: false,
  }
}

Component({
  options: {
    addGlobalClass: true,
  },
  properties: {
    grade: {
      type: String,
      value: 'free',
    },
    vipDueAt: {
      type: String,
      value: '',
    },
  },
  data: {
    states: computeStates('free', null),
  },
  observers: {
    'grade, vipDueAt'(grade: string, vipDueAt: string) {
      const g = (grade || 'free') as Grade
      const vip = vipDueAt ? vipDueAt : null
      this.setData({ states: computeStates(g, vip) })
    },
  },
  methods: {
    goPurchase(e: WechatMiniprogram.CustomEvent) {
      const disabled = e.currentTarget.dataset.disabled
      if (disabled) return
      wx.navigateTo({ url: '/pages/me/purchase/purchase' })
    },
  },
})
