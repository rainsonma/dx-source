import { getAvatarColor, getAvatarLetter } from '../../../../utils/avatar'
import { formatDate, formatNumber } from '../../../../utils/format'
import { config } from '../../../../utils/config'
import type { Post } from '../../types'

Component({
  options: { addGlobalClass: true },
  properties: {
    post: { type: Object, value: null },
    theme: { type: String, value: 'light' },
    followed: { type: Boolean, value: false },
    isOwner: { type: Boolean, value: false },
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
        timeText: formatDate(post.created_at),
        likeText: post.like_count > 0 ? formatNumber(post.like_count) : '',
        commentText: post.comment_count > 0 ? formatNumber(post.comment_count) : '',
        imageAbsoluteUrl: post.image_url ? config.apiBaseUrl + post.image_url : '',
      })
    },
  },
  methods: {
    onPreviewImage() {
      const url = (this.data as { imageAbsoluteUrl: string }).imageAbsoluteUrl
      if (url) wx.previewImage({ urls: [url] })
    },
    onLikeTap() {
      this.triggerEvent('toggle-like')
    },
    onBookmarkTap() {
      this.triggerEvent('toggle-bookmark')
    },
    onFollowTap() {
      this.triggerEvent('toggle-follow')
    },
    onMoreTap() {
      this.triggerEvent('open-actions')
    },
  },
})
