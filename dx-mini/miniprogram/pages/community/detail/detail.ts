import { api } from '../../../utils/api'
import { PaginatedData } from '../../../utils/api'
import type { Post, CommentWithReplies } from '../types'

const app = getApp<{ globalData: { theme: 'light' | 'dark' } }>()

Page({
  data: {
    theme: 'light' as 'light' | 'dark',
    statusBarHeight: 20,
    postId: '',
    post: null as Post | null,
    loading: false,
    followed: false,
    comments: [] as CommentWithReplies[],
    commentsCursor: '',
    commentsHasMore: false,
    commentsLoading: false,
  },
  onLoad(query: Record<string, string>) {
    const sys = wx.getSystemInfoSync()
    this.setData({
      theme: app.globalData.theme,
      statusBarHeight: sys.statusBarHeight || 20,
      postId: query.id || '',
    })
    if (query.id) this.loadPost()
    if (query.id) this.loadComments(true)
  },
  onReachBottom() {
    if (this.data.commentsHasMore && !this.data.commentsLoading) {
      this.loadComments(false)
    }
  },
  onShow() {
    this.setData({ theme: app.globalData.theme })
  },
  goBack() {
    wx.navigateBack()
  },
  async loadPost() {
    this.setData({ loading: true })
    try {
      const post = await api.get<Post>(`/api/posts/${this.data.postId}`)
      this.setData({ post, loading: false })
    } catch (err) {
      this.setData({ loading: false })
      wx.showToast({ title: (err as Error).message || '加载失败', icon: 'none' })
    }
  },
  async loadComments(reset: boolean) {
    if (this.data.commentsLoading) return
    this.setData({ commentsLoading: true })
    const cursor = reset ? '' : this.data.commentsCursor
    const parts = ['limit=20']
    if (cursor) parts.push(`cursor=${encodeURIComponent(cursor)}`)
    try {
      const res = await api.get<PaginatedData<CommentWithReplies>>(
        `/api/posts/${this.data.postId}/comments?${parts.join('&')}`
      )
      this.setData({
        comments: reset ? res.items : [...this.data.comments, ...res.items],
        commentsCursor: res.nextCursor,
        commentsHasMore: res.hasMore,
        commentsLoading: false,
      })
    } catch (err) {
      this.setData({ commentsLoading: false })
      wx.showToast({ title: (err as Error).message || '加载评论失败', icon: 'none' })
    }
  },
  emitUpdate(patch: Partial<Post>) {
    if (!this.data.post) return
    let channel: WechatMiniprogram.EventChannel | null = null
    try { channel = this.getOpenerEventChannel() } catch { channel = null }
    if (channel) channel.emit('post-updated', { id: this.data.post.id, patch })
  },
  async onToggleLike() {
    if (!this.data.post) return
    const before = this.data.post
    const optimistic: Post = {
      ...before,
      is_liked: !before.is_liked,
      like_count: before.is_liked ? Math.max(before.like_count - 1, 0) : before.like_count + 1,
    }
    this.setData({ post: optimistic })
    try {
      const res = await api.post<{ liked: boolean; like_count: number }>(`/api/posts/${before.id}/like`, {})
      const next = { ...optimistic, is_liked: res.liked, like_count: res.like_count }
      this.setData({ post: next })
      this.emitUpdate({ is_liked: res.liked, like_count: res.like_count })
    } catch (err) {
      this.setData({ post: before })
      wx.showToast({ title: (err as Error).message || '操作失败', icon: 'none' })
    }
  },
  async onToggleBookmark() {
    if (!this.data.post) return
    const before = this.data.post
    const optimistic: Post = { ...before, is_bookmarked: !before.is_bookmarked }
    this.setData({ post: optimistic })
    try {
      const res = await api.post<{ bookmarked: boolean }>(`/api/posts/${before.id}/bookmark`, {})
      const next = { ...optimistic, is_bookmarked: res.bookmarked }
      this.setData({ post: next })
      this.emitUpdate({ is_bookmarked: res.bookmarked })
    } catch (err) {
      this.setData({ post: before })
      wx.showToast({ title: (err as Error).message || '操作失败', icon: 'none' })
    }
  },
  async onToggleFollow() {
    if (!this.data.post) return
    const before = this.data.followed
    this.setData({ followed: !before })
    try {
      const res = await api.post<{ followed: boolean }>(`/api/users/${this.data.post.author.id}/follow`, {})
      this.setData({ followed: res.followed })
    } catch (err) {
      this.setData({ followed: before })
      wx.showToast({ title: (err as Error).message || '操作失败', icon: 'none' })
    }
  },
})
