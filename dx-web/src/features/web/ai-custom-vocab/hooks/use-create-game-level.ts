"use client"

import { useActionState, useEffect, useRef, useCallback } from "react"
import { swrMutate } from "@/lib/swr"

import {
  createGameLevelAction,
  type GameLevelActionResult,
} from "@/features/web/ai-custom-vocab/actions/course-game.action"

const initialState: GameLevelActionResult = {}

export function useCreateGameLevel(gameId: string, onSuccess?: () => void) {
  const onSuccessRef = useRef(onSuccess)
  useEffect(() => { onSuccessRef.current = onSuccess })

  const boundAction = useCallback(
    (prev: GameLevelActionResult, formData: FormData) => createGameLevelAction(gameId, prev, formData),
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

  return { state, formAction, isPending }
}
