"use client"

import { Spinner } from "@/components/ui/spinner"
import { useComments } from "../hooks/use-comments"
import { CommentItem } from "./comment-item"
import { CommentInput } from "./comment-input"

interface CommentSectionProps {
  postId: string
  currentUserId?: string
}

export function CommentSection({ postId, currentUserId }: CommentSectionProps) {
  const { comments, isLoading, hasMore, sentinelRef, mutate } = useComments(postId)

  return (
    <div className="flex flex-col gap-3.5 rounded-[10px] bg-muted/50 p-4">
      {isLoading && (
        <div className="flex justify-center py-2">
          <Spinner />
        </div>
      )}

      {!isLoading && comments.length === 0 && (
        <p className="text-center text-xs text-muted-foreground">暂无评论</p>
      )}

      {comments.map(({ comment, replies }) => (
        <div key={comment.id} className="flex flex-col gap-2.5">
          <CommentItem postId={postId} comment={comment} currentUserId={currentUserId} onMutate={mutate} />
          {replies.length > 0 && (
            <div className="ml-10 flex flex-col gap-2.5 border-l-2 border-border pl-3">
              {replies.map((reply) => (
                <CommentItem
                  key={reply.id}
                  postId={postId}
                  comment={reply}
                  showReply={false}
                  currentUserId={currentUserId}
                  onMutate={mutate}
                />
              ))}
            </div>
          )}
        </div>
      ))}

      {hasMore && (
        <div ref={sentinelRef} className="flex justify-center py-1">
          <Spinner className="h-3 w-3" />
        </div>
      )}

      <div className="mt-1">
        <CommentInput postId={postId} onSuccess={mutate} />
      </div>
    </div>
  )
}
