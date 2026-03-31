"use client"

import { useState } from "react"
import { Heart, MessageCircle, Bookmark } from "lucide-react"
import { toast } from "sonner"
import { postApi } from "../actions/post.action"

interface PostActionsProps {
  postId: string
  likeCount: number
  commentCount: number
  isLiked: boolean
  isBookmarked: boolean
  onCommentClick?: () => void
  onMutate?: () => void
}

export function PostActions({
  postId,
  likeCount: initialLikeCount,
  commentCount,
  isLiked: initialIsLiked,
  isBookmarked: initialIsBookmarked,
  onCommentClick,
  onMutate,
}: PostActionsProps) {
  const [isLiked, setIsLiked] = useState(initialIsLiked)
  const [likeCount, setLikeCount] = useState(initialLikeCount)
  const [isBookmarked, setIsBookmarked] = useState(initialIsBookmarked)
  const [pending, setPending] = useState(false)

  async function handleLike() {
    if (pending) return
    const prevLiked = isLiked
    const prevCount = likeCount
    setIsLiked(!prevLiked)
    setLikeCount(prevLiked ? prevCount - 1 : prevCount + 1)
    setPending(true)
    try {
      const res = await postApi.toggleLike(postId)
      if (res.code !== 0) {
        setIsLiked(prevLiked)
        setLikeCount(prevCount)
        toast.error(res.message)
        return
      }
      setIsLiked(res.data.liked)
      setLikeCount(res.data.like_count)
      onMutate?.()
    } catch {
      setIsLiked(prevLiked)
      setLikeCount(prevCount)
      toast.error("操作失败")
    } finally {
      setPending(false)
    }
  }

  async function handleBookmark() {
    if (pending) return
    const prev = isBookmarked
    setIsBookmarked(!prev)
    setPending(true)
    try {
      const res = await postApi.toggleBookmark(postId)
      if (res.code !== 0) {
        setIsBookmarked(prev)
        toast.error(res.message)
        return
      }
      setIsBookmarked(res.data.bookmarked)
      onMutate?.()
    } catch {
      setIsBookmarked(prev)
      toast.error("操作失败")
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="flex w-full items-center justify-around">
      <button
        type="button"
        onClick={handleLike}
        className="flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm text-muted-foreground hover:bg-muted"
      >
        <Heart
          className={`h-4 w-4 ${isLiked ? "fill-red-500 text-red-500" : ""}`}
        />
        <span>{likeCount || ""}</span>
      </button>
      <button
        type="button"
        onClick={onCommentClick}
        className="flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm text-muted-foreground hover:bg-muted"
      >
        <MessageCircle className="h-4 w-4" />
        <span>{commentCount || ""}</span>
      </button>
      <button
        type="button"
        onClick={handleBookmark}
        className="flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm text-muted-foreground hover:bg-muted"
      >
        <Bookmark
          className={`h-4 w-4 ${isBookmarked ? "fill-teal-600 text-teal-600" : ""}`}
        />
      </button>
    </div>
  )
}
