"use client"

import { useActionState, useState, useEffect, useRef, useCallback } from "react"
import { swrMutate } from "@/lib/swr"

import {
  updateCourseGameAction,
  type UpdateGameResult,
} from "@/features/web/ai-custom/actions/course-game.action"

const initialState: UpdateGameResult = {}

export function useUpdateCourseGame(gameId: string, onSuccess?: () => void) {
  const [coverUrl, setCoverUrl] = useState<string | null>(null)
  const onSuccessRef = useRef(onSuccess)
  useEffect(() => { onSuccessRef.current = onSuccess })

  const boundAction = useCallback(
    (prev: UpdateGameResult, formData: FormData) => updateCourseGameAction(gameId, prev, formData),
    [gameId]
  )

  const [state, formAction, isPending] = useActionState(
    boundAction,
    initialState
  )

  useEffect(() => {
    if (state.success) {
      onSuccessRef.current?.()
      swrMutate("/api/course-games")
    }
  }, [state])

  return {
    state,
    formAction,
    isPending,
    coverUrl,
    setCoverUrl,
  }
}
