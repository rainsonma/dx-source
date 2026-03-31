"use client"

import { useState } from "react"
import { Loader2, Send } from "lucide-react"
import { toast } from "sonner"
import { postApi } from "../actions/post.action"
import { createCommentSchema } from "../schemas/post.schema"

interface CommentInputProps {
  postId: string
  parentId?: string
  placeholder?: string
  onSuccess?: () => void
  onCancel?: () => void
}

export function CommentInput({
  postId,
  parentId,
  placeholder = "写下你的评论...",
  onSuccess,
  onCancel,
}: CommentInputProps) {
  const [content, setContent] = useState("")
  const [pending, setPending] = useState(false)

  async function handleSubmit() {
    const result = createCommentSchema.safeParse({ content, parent_id: parentId })
    if (!result.success) {
      toast.error(result.error.issues[0]?.message ?? "内容不能为空")
      return
    }
    setPending(true)
    try {
      const res = await postApi.createComment(postId, { content, parent_id: parentId })
      if (res.code !== 0) {
        toast.error(res.message)
        return
      }
      setContent("")
      onSuccess?.()
    } catch {
      toast.error("评论失败")
    } finally {
      setPending(false)
    }
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    }
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-2.5">
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          rows={1}
          maxLength={500}
          className="flex-1 resize-none rounded-full border border-border bg-card px-3.5 py-2 text-[13px] placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
        <button
          type="button"
          onClick={handleSubmit}
          disabled={pending || !content.trim()}
          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-teal-600 disabled:opacity-50"
        >
          {pending ? (
            <Loader2 className="h-4 w-4 animate-spin text-white" />
          ) : (
            <Send className="h-4 w-4 text-white" />
          )}
        </button>
      </div>
      {onCancel && (
        <button
          type="button"
          onClick={onCancel}
          className="self-end text-xs text-muted-foreground hover:text-foreground"
        >
          取消
        </button>
      )}
    </div>
  )
}
