import { getAvatarColor, getAvatarLetter } from '../../../../utils/avatar'
import { formatRelativeDate, formatNumber } from '../../../../utils/format'
import { config } from '../../../../utils/config'
import type { Post } from '../../types'

Component({
  options: { addGlobalClass: true },
  properties: {
    post: { type: Object, value: null },
    theme: { type: String, value: 'light' },
  },
  data: {
    avatarColor: '#999',
    avatarLetter: '?',
    timeText: '',
    likeText: '',
    commentText: '',
    imageAbsoluteUrl: '',
  },
  observers: {
    post(post: Post | null) {
      if (!post) return
      this.setData({
        avatarColor: getAvatarColor(post.author.id),
        avatarLetter: getAvatarLetter(post.author.nickname),
        timeText: formatRelativeDate(post.created_at),
        likeText: post.like_count > 0 ? formatNumber(post.like_count) : '',
        commentText: post.comment_count > 0 ? formatNumber(post.comment_count) : '',
        imageAbsoluteUrl: post.image_url ? config.apiBaseUrl + post.image_url : '',
      })
    },
  },
  methods: {
    onCardTap() {
      this.triggerEvent('opendetail', { id: (this.data as { post: Post }).post.id })
    },
    onLikeTap() {
      this.triggerEvent('toggle-like', { id: (this.data as { post: Post }).post.id })
    },
    onCommentTap() {
      this.triggerEvent('opendetail', { id: (this.data as { post: Post }).post.id })
    },
    onBookmarkTap() {
      this.triggerEvent('toggle-bookmark', { id: (this.data as { post: Post }).post.id })
    },
    onFollowTap(e: WechatMiniprogram.TouchEvent) {
      void e
    },
  },
})
