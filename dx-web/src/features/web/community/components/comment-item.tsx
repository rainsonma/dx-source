"use client"

import { useState } from "react"
import { toast } from "sonner"
import { getAvatarColor } from "@/lib/avatar"
import type { Comment } from "../types/post"
import { postApi } from "../actions/post.action"
import { CommentInput } from "./comment-input"
import { PostActionsMenu } from "./post-actions-menu"

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

interface CommentItemProps {
  postId: string
  comment: Comment
  showReply?: boolean
  currentUserId?: string
  onMutate?: () => void
}

export function CommentItem({ postId, comment, showReply = true, currentUserId, onMutate }: CommentItemProps) {
  const [replying, setReplying] = useState(false)
  const [editing, setEditing] = useState(false)
  const color = getAvatarColor(comment.author.id)
  const letter = comment.author.nickname.charAt(0)
  const isOwner = currentUserId !== undefined && currentUserId === comment.author.id

  async function handleDelete() {
    if (!confirm("确认删除？")) return
    try {
      const res = await postApi.deleteComment(postId, comment.id)
      if (res.code !== 0) { toast.error(res.message); return }
      toast.success("已删除")
      onMutate?.()
    } catch {
      toast.error("删除失败")
    }
  }

  return (
    <div className="flex gap-2.5">
      <div
        className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold text-white"
        style={{ backgroundColor: color }}
      >
        {letter}
      </div>
      <div className="flex flex-1 flex-col gap-1.5">
        <div className="flex items-center gap-2">
          <span className="text-[13px] font-semibold text-foreground">
            {comment.author.nickname}
          </span>
          <span className="text-xs text-muted-foreground">{timeAgo(comment.created_at)}</span>
          {isOwner && (
            <div className="ml-auto">
              <PostActionsMenu
                size="sm"
                onEdit={() => { setEditing(true); setReplying(false) }}
                onDelete={handleDelete}
              />
            </div>
          )}
        </div>
        {editing ? (
          <div className="mt-1">
            <CommentInput
              postId={postId}
              commentId={comment.id}
              initialContent={comment.content}
              onSuccess={() => {
                setEditing(false)
                onMutate?.()
              }}
              onCancel={() => setEditing(false)}
            />
          </div>
        ) : (
          <p className="text-[13px] leading-relaxed text-muted-foreground">{comment.content}</p>
        )}
        {!editing && showReply && (
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => setReplying((v) => !v)}
              className="text-xs text-muted-foreground hover:text-teal-600"
            >
              回复
            </button>
          </div>
        )}
        {replying && (
          <div className="mt-1">
            <CommentInput
              postId={postId}
              parentId={comment.id}
              placeholder={`回复 ${comment.author.nickname}...`}
              onSuccess={() => {
                setReplying(false)
                onMutate?.()
              }}
              onCancel={() => setReplying(false)}
            />
          </div>
        )}
      </div>
    </div>
  )
}
