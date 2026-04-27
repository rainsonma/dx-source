import { getAvatarColor, getAvatarLetter } from '../../../../utils/avatar'
import { formatRelativeDate } from '../../../../utils/format'
import type { Comment, CommentWithReplies } from '../../types'

interface DisplayComment {
  comment: Comment
  replies: Comment[]
}

Component({
  options: { addGlobalClass: true },
  properties: {
    item: { type: Object, value: null },
    theme: { type: String, value: 'light' },
    isOwner: { type: Boolean, value: false },
  },
  data: {
    parentColor: '#999',
    parentLetter: '?',
    parentTime: '',
    replyDecor: [] as { color: string; letter: string; time: string }[],
  },
  observers: {
    item(item: DisplayComment | null) {
      if (!item) return
      this.setData({
        parentColor: getAvatarColor(item.comment.author.id),
        parentLetter: getAvatarLetter(item.comment.author.nickname),
        parentTime: formatRelativeDate(item.comment.created_at),
        replyDecor: item.replies.map((r) => ({
          color: getAvatarColor(r.author.id),
          letter: getAvatarLetter(r.author.nickname),
          time: formatRelativeDate(r.created_at),
        })),
      })
    },
  },
  methods: {
    onReply() {
      const item = (this.data as { item: CommentWithReplies }).item
      this.triggerEvent('reply', {
        commentId: item.comment.id,
        nickname: item.comment.author.nickname,
      })
    },
    onMore() {
      this.triggerEvent('open-actions', {
        commentId: (this.data as { item: CommentWithReplies }).item.comment.id,
      })
    },
  },
})
