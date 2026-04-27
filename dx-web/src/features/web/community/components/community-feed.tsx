"use client"

import { useState } from "react"
import { Plus } from "lucide-react"
import { Spinner } from "@/components/ui/spinner"

import type { FeedTab } from "../types/post"
import { usePostFeed } from "../hooks/use-post-feed"
import { FeedTabs } from "./feed-tabs"
import { PostCard } from "./post-card"
import { CreatePostDialog } from "./create-post-dialog"

interface CommunityFeedProps {
  currentUserId?: string
}

export function CommunityFeed({ currentUserId }: CommunityFeedProps) {
  const [tab, setTab] = useState<FeedTab>("latest")
  const [createOpen, setCreateOpen] = useState(false)
  const { posts, isLoading, hasMore, sentinelRef, mutate } = usePostFeed(tab)

  return (
    <>
      {/* Tab row */}
      <div className="flex items-center justify-between">
        <FeedTabs active={tab} onChange={setTab} />
        <button
          type="button"
          onClick={() => setCreateOpen(true)}
          className="flex items-center gap-2 rounded-[10px] bg-teal-600 px-5 py-2.5 text-sm font-semibold text-white hover:bg-teal-700"
        >
          <Plus className="h-4 w-4" />
          发帖
        </button>
      </div>

      {/* Feed */}
      <div className="flex flex-col gap-4">
        {isLoading && posts.length === 0 && (
          <div className="flex justify-center py-12">
            <Spinner className="h-6 w-6 text-teal-600" />
          </div>
        )}

        {!isLoading && posts.length === 0 && (
          <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
            <p className="text-sm">暂无帖子</p>
          </div>
        )}

        {posts.map((post) => (
          <PostCard
            key={post.id}
            post={post}
            currentUserId={currentUserId}
            onMutate={() => mutate()}
          />
        ))}

        {hasMore && <div ref={sentinelRef} className="h-1" />}
      </div>

      <CreatePostDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onCreated={() => mutate()}
      />
    </>
  )
}
