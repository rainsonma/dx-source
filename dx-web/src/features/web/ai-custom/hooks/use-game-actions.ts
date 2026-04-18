"use client"

import { useCallback, useState, useTransition } from "react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import { swrMutate } from "@/lib/swr"

import {
  deleteGameAction,
  deleteGameLevelAction,
  publishGameAction,
  withdrawGameAction,
} from "@/features/web/ai-custom/actions/course-game.action"

export function useDeleteGame(gameId: string) {
  const router = useRouter()
  const [isPending, setIsPending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function execute(onSuccess?: () => void) {
    setIsPending(true)
    setError(null)
    const result = await deleteGameAction(gameId)
    if (result.error) {
      setError(result.error)
      setIsPending(false)
      return
    }
    await swrMutate("/api/course-games")
    onSuccess?.()
    router.replace("/hall/ai-custom")
  }

  const clearError = useCallback(() => setError(null), [])

  return { execute, isPending, error, clearError }
}

export function useDeleteGameLevel(gameId: string) {
  const [isPending, startTransition] = useTransition()
  const [error, setError] = useState<string | null>(null)

  function execute(levelId: string) {
    startTransition(async () => {
      const result = await deleteGameLevelAction(gameId, levelId)
      if (result.error) {
        setError(result.error)
      } else {
        await swrMutate("/api/course-games")
      }
    })
  }

  return { execute, isPending, error }
}

export function usePublishGame(gameId: string) {
  const [isPending, startTransition] = useTransition()

  function execute() {
    startTransition(async () => {
      const result = await publishGameAction(gameId)
      if (result.error) {
        toast.error(result.error)
      } else {
        await swrMutate("/api/course-games")
      }
    })
  }

  return { execute, isPending }
}

export function useWithdrawGame(gameId: string) {
  const [isPending, startTransition] = useTransition()
  const [error, setError] = useState<string | null>(null)

  function execute(onSuccess?: () => void) {
    startTransition(async () => {
      setError(null)
      const result = await withdrawGameAction(gameId)
      if (result.error) {
        setError(result.error)
        return
      }
      await swrMutate("/api/course-games")
      onSuccess?.()
    })
  }

  const clearError = useCallback(() => setError(null), [])

  return { execute, isPending, error, clearError }
}
