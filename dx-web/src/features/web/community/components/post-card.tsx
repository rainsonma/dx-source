"use client"

import { useState } from "react"
import Image from "next/image"
import { UserPlus } from "lucide-react"
import { toast } from "sonner"
import { getAvatarColor } from "@/lib/avatar"
import { postApi } from "../actions/post.action"
import type { Post } from "../types/post"
import { PostActions } from "./post-actions"
import { CommentSection } from "./comment-section"
import { PostActionsMenu } from "./post-actions-menu"
import { CreatePostDialog } from "./create-post-dialog"

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const minutes = Math.floor(diff / 60_000)
  if (minutes < 1) return "刚刚"
  if (minutes < 60) return `${minutes}分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}天前`
  return new Date(dateStr).toLocaleDateString("zh-CN")
}

interface PostCardProps {
  post: Post
  currentUserId?: string
  onMutate?: () => void
}

export function PostCard({ post, currentUserId, onMutate }: PostCardProps) {
  const [showComments, setShowComments] = useState(false)
  const [followed, setFollowed] = useState(false)
  const [followPending, setFollowPending] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const isOwner = currentUserId !== undefined && currentUserId === post.author.id

  const color = getAvatarColor(post.author.id)
  const letter = post.author.nickname.charAt(0)

  async function handleDelete() {
    if (!confirm("确认删除？")) return
    try {
      const res = await postApi.delete(post.id)
      if (res.code !== 0) { toast.error(res.message); return }
      toast.success("已删除")
      onMutate?.()
    } catch {
      toast.error("删除失败")
    }
  }

  async function handleFollow() {
    if (followPending) return
    const prev = followed
    setFollowed(!prev)
    setFollowPending(true)
    try {
      const res = await postApi.toggleFollow(post.author.id)
      if (res.code !== 0) {
        setFollowed(prev)
        toast.error(res.message)
        return
      }
      setFollowed(res.data.followed)
    } catch {
      setFollowed(prev)
      toast.error("操作失败")
    } finally {
      setFollowPending(false)
    }
  }

  return (
    <div className="flex flex-col gap-4 rounded-xl border border-border bg-card p-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div
            className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full text-sm font-semibold text-white"
            style={{ backgroundColor: color }}
          >
            {letter}
          </div>
          <div className="flex flex-col gap-0.5">
            <span className="text-sm font-semibold text-foreground">
              {post.author.nickname}
            </span>
            <span className="text-xs text-muted-foreground">
              {timeAgo(post.created_at)}
            </span>
          </div>
        </div>
        {isOwner ? (
          <PostActionsMenu
            onEdit={() => setEditOpen(true)}
            onDelete={handleDelete}
          />
        ) : followed ? (
          <span className="rounded-full bg-muted px-4 py-1.5 text-[13px] font-medium text-muted-foreground">
            已关注
          </span>
        ) : (
          <button
            type="button"
            onClick={handleFollow}
            disabled={followPending}
            className="flex items-center gap-1.5 rounded-full border border-teal-600 px-4 py-1.5 text-[13px] font-semibold text-teal-600 disabled:opacity-50"
          >
            <UserPlus className="h-3.5 w-3.5" />
            关注
          </button>
        )}
      </div>

      {/* Content */}
      <p className="text-sm leading-relaxed text-foreground">{post.content}</p>

      {/* Image */}
      {post.image_url && (
        <div className="relative h-[220px] w-full overflow-hidden rounded-[10px]">
          <Image
            src={post.image_url}
            alt="post image"
            fill
            className="object-cover"
          />
        </div>
      )}

      {/* Tags */}
      {post.tags.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {post.tags.map((tag) => (
            <span
              key={tag}
              className="rounded-md bg-teal-50 px-2.5 py-1 text-xs font-medium text-teal-700"
            >
              #{tag}
            </span>
          ))}
        </div>
      )}

      <div className="h-px w-full bg-border" />

      {/* Actions */}
      <PostActions
        postId={post.id}
        likeCount={post.like_count}
        commentCount={post.comment_count}
        isLiked={post.is_liked}
        isBookmarked={post.is_bookmarked}
        onCommentClick={() => setShowComments((v) => !v)}
        onMutate={onMutate}
      />

      {/* Comments */}
      {showComments && <CommentSection postId={post.id} currentUserId={currentUserId} />}

      {editOpen && (
        <CreatePostDialog
          open={editOpen}
          onOpenChange={setEditOpen}
          onCreated={() => onMutate?.()}
          editPost={post}
        />
      )}
    </div>
  )
}
