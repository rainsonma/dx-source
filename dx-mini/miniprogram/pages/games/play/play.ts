import { api } from '../../../utils/api'

interface ContentItemData {
  id: string; content: string; contentType: string
  translation: string | null; definition: string | null
  items: string | null
}

interface StartSessionResult {
  id: string; gameLevelId: string; degree: string
  score: number; exp: number; maxCombo: number
  correctCount: number; wrongCount: number
  currentContentItemId: string | null
}

interface Choice { text: string; correct: boolean }

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    loading: true,
    sessionId: '',
    gameLevelId: '',
    gameId: '',
    contentItems: [] as ContentItemData[],
    currentIndex: 0,
    currentItem: null as ContentItemData | null,
    choices: [] as Choice[],
    score: 0,
    combo: 0,
    maxCombo: 0,
    correctCount: 0,
    wrongCount: 0,
    answered: false,
    selectedChoice: -1,
    showResult: false,
    startTime: 0,
  },
  onLoad(options: { gameId?: string; levelId?: string; degree?: string }) {
    this.setData({ theme: app.globalData.theme })
    if (options.gameId && options.levelId) {
      this.initSession(options.gameId, options.levelId, options.degree || 'normal')
    }
  },
  async initSession(gameId: string, levelId: string, degree: string) {
    try {
      const session = await api.post<StartSessionResult>('/api/play-single/start', {
        game_id: gameId,
        game_level_id: levelId,
        degree,
        pattern: null,
      })
      const items = await api.get<ContentItemData[]>(
        `/api/games/${gameId}/levels/${levelId}/content?degree=${degree}`
      )
      let startIndex = 0
      if (session.currentContentItemId) {
        const idx = items.findIndex(i => i.id === session.currentContentItemId)
        if (idx >= 0) startIndex = idx
      }
      this.setData({
        loading: false,
        sessionId: session.id,
        gameLevelId: levelId,
        gameId,
        contentItems: items,
        currentIndex: startIndex,
        score: session.score,
        combo: 0,
        maxCombo: session.maxCombo,
        correctCount: session.correctCount,
        wrongCount: session.wrongCount,
        startTime: Date.now(),
      })
      this.showCurrentItem()
    } catch (err) {
      wx.showToast({ title: (err as Error).message || '启动失败', icon: 'none' })
      wx.navigateBack()
    }
  },
  showCurrentItem() {
    const item = this.data.contentItems[this.data.currentIndex]
    if (!item) {
      this.endSession()
      return
    }
    const choices = this.buildChoices(item)
    this.setData({ currentItem: item, choices, answered: false, selectedChoice: -1, startTime: Date.now() })
  },
  buildChoices(item: ContentItemData): Choice[] {
    if (!item.items) return []
    try {
      const parsed = JSON.parse(item.items) as unknown[]
      if (Array.isArray(parsed)) {
        return parsed.map((c: unknown) => {
          if (typeof c === 'string') return { text: c, correct: c === item.content }
          const obj = c as { text?: string; correct?: boolean }
          return { text: obj.text || '', correct: obj.correct || false }
        })
      }
    } catch {}
    return []
  },
  selectChoice(e: WechatMiniprogram.TouchEvent) {
    if (this.data.answered) return
    const idx = e.currentTarget.dataset['idx'] as number
    const choice = this.data.choices[idx]
    const isCorrect = choice.correct
    const duration = Date.now() - this.data.startTime
    const baseScore = isCorrect ? 10 : 0
    const newCombo = isCorrect ? this.data.combo + 1 : 0
    const comboScore = isCorrect ? Math.min(newCombo - 1, 5) * 2 : 0
    const score = this.data.score + baseScore + comboScore
    const maxCombo = Math.max(this.data.maxCombo, newCombo)
    const nextIndex = this.data.currentIndex + 1
    const nextItem = this.data.contentItems[nextIndex] || null
    this.setData({
      answered: true,
      selectedChoice: idx,
      score,
      combo: newCombo,
      maxCombo,
      correctCount: this.data.correctCount + (isCorrect ? 1 : 0),
      wrongCount: this.data.wrongCount + (isCorrect ? 0 : 1),
    })
    api.post('/api/play-single/' + this.data.sessionId + '/answers', {
      game_session_id: this.data.sessionId,
      game_level_id: this.data.gameLevelId,
      content_item_id: this.data.currentItem!.id,
      is_correct: isCorrect,
      user_answer: choice.text,
      source_answer: this.data.currentItem!.content,
      base_score: baseScore,
      combo_score: comboScore,
      score: baseScore + comboScore,
      max_combo: maxCombo,
      play_time: Math.floor(duration / 1000),
      duration: Math.floor(duration / 1000),
      next_content_item_id: nextItem ? nextItem.id : null,
    }).catch(() => {})
    setTimeout(() => {
      this.setData({ currentIndex: nextIndex })
      this.showCurrentItem()
    }, isCorrect ? 600 : 1200)
  },
  markWord(e: WechatMiniprogram.TouchEvent) {
    const type = e.currentTarget.dataset['type'] as string
    const item = this.data.currentItem
    if (!item) return
    const path = `/api/tracking/${type}`
    const body: Record<string, string> = { content_item_id: item.id, game_id: this.data.gameId, game_level_id: this.data.gameLevelId }
    if (type === 'review') { delete body['game_id']; delete body['game_level_id'] }
    api.post(path, body).then(() => wx.showToast({ title: '已标记', icon: 'none' })).catch(() => {})
  },
  async endSession() {
    try {
      await api.post('/api/play-single/' + this.data.sessionId + '/end', {
        score: this.data.score,
        exp: Math.floor(this.data.score / 10),
        max_combo: this.data.maxCombo,
        correct_count: this.data.correctCount,
        wrong_count: this.data.wrongCount,
        skip_count: 0,
      })
    } catch {}
    this.setData({ showResult: true })
  },
  goBack() {
    wx.navigateBack()
  },
})
