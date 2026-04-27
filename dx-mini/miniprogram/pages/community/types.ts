export interface PostAuthor {
  id: string
  nickname: string
  avatar_url: string | null
}

export interface Post {
  id: string
  content: string
  image_url: string | null
  tags: string[]
  like_count: number
  comment_count: number
  is_liked: boolean
  is_bookmarked: boolean
  author: PostAuthor
  created_at: string
}

export interface Comment {
  id: string
  content: string
  author: PostAuthor
  parent_id: string | null
  created_at: string
}

export interface CommentWithReplies {
  comment: Comment
  replies: Comment[]
}

export type FeedTab = 'latest' | 'hot' | 'following' | 'bookmarked'

export const FEED_TABS: { name: FeedTab; title: string }[] = [
  { name: 'latest',     title: '最新' },
  { name: 'hot',        title: '热门' },
  { name: 'following',  title: '关注' },
  { name: 'bookmarked', title: '收藏' },
]
