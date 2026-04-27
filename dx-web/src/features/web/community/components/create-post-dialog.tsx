"use client"

import { useEffect, useState } from "react"
import { Loader2, X } from "lucide-react"
import { toast } from "sonner"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { postApi } from "../actions/post.action"
import { createPostSchema } from "../schemas/post.schema"
import type { Post } from "../types/post"

interface CreatePostDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated: () => void
  editPost?: Post
}

export function CreatePostDialog({ open, onOpenChange, onCreated, editPost }: CreatePostDialogProps) {
  const [content, setContent] = useState("")
  const [tagInput, setTagInput] = useState("")
  const [tags, setTags] = useState<string[]>([])
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)

  useEffect(() => {
    if (editPost) {
      setContent(editPost.content)
      setTags(editPost.tags)
      setTagInput("")
      setErrors({})
    }
  }, [editPost?.id])

  function resetForm() {
    setContent("")
    setTagInput("")
    setTags([])
    setErrors({})
  }

  function addTag() {
    const tag = tagInput.trim().replace(/^#/, "")
    if (!tag) return
    if (tags.length >= 5) {
      toast.error("标签不能超过5个")
      return
    }
    if (tag.length > 20) {
      toast.error("标签不能超过20个字符")
      return
    }
    if (!tags.includes(tag)) {
      setTags([...tags, tag])
    }
    setTagInput("")
  }

  function removeTag(tag: string) {
    setTags(tags.filter((t) => t !== tag))
  }

  function handleTagKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter") {
      e.preventDefault()
      addTag()
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setErrors({})

    const result = createPostSchema.safeParse({ content, tags: tags.length > 0 ? tags : undefined })
    if (!result.success) {
      const fieldErrors: Record<string, string> = {}
      for (const issue of result.error.issues) {
        const key = issue.path[0] as string
        if (!fieldErrors[key]) fieldErrors[key] = issue.message
      }
      setErrors(fieldErrors)
      return
    }

    setPending(true)
    try {
      if (editPost) {
        const res = await postApi.update(editPost.id, { content, tags: tags.length > 0 ? tags : undefined })
        if (res.code !== 0) {
          toast.error(res.message)
          return
        }
        toast.success("已保存")
        onOpenChange(false)
        onCreated()
      } else {
        const res = await postApi.create({ content, tags: tags.length > 0 ? tags : undefined })
        if (res.code !== 0) {
          toast.error(res.message)
          return
        }
        toast.success("发布成功")
        resetForm()
        onOpenChange(false)
        onCreated()
      }
    } catch {
      toast.error(editPost ? "保存失败" : "发布失败")
    } finally {
      setPending(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v && !editPost) resetForm(); onOpenChange(v) }}>
      <DialogContent className="sm:max-w-lg" aria-describedby={undefined}>
        <DialogHeader>
          <DialogTitle>{editPost ? "编辑帖子" : "发布帖子"}</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          {/* Content textarea */}
          <div className="flex flex-col gap-1.5">
            <div className="relative">
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="分享你的学习心得..."
                maxLength={2000}
                rows={6}
                className="flex w-full resize-none rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
              <span className="absolute bottom-2 right-3 text-xs text-muted-foreground">
                {content.length}/2000
              </span>
            </div>
            {errors.content && (
              <p className="text-xs text-red-500">{errors.content}</p>
            )}
          </div>

          {/* Tags */}
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium text-foreground">
              标签（选填，最多5个）
            </label>
            <div className="flex flex-wrap gap-2">
              {tags.map((tag) => (
                <span
                  key={tag}
                  className="flex items-center gap-1 rounded-md bg-teal-50 px-2.5 py-1 text-xs font-medium text-teal-700"
                >
                  #{tag}
                  <button
                    type="button"
                    onClick={() => removeTag(tag)}
                    className="ml-0.5 rounded-full hover:text-teal-900"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </span>
              ))}
            </div>
            {tags.length < 5 && (
              <input
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={handleTagKeyDown}
                onBlur={addTag}
                placeholder="输入标签后按 Enter"
                maxLength={20}
                className="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => { if (!editPost) resetForm(); onOpenChange(false) }}
            >
              取消
            </Button>
            <Button
              type="submit"
              disabled={pending}
              className="bg-teal-600 hover:bg-teal-700"
            >
              {pending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {editPost ? "保存" : "发布"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
