"use client"

import { useActionState, useEffect, useRef, useCallback } from "react"
import { swrMutate } from "@/lib/swr"

import {
  createGameLevelAction,
  type GameLevelActionResult,
} from "@/features/web/ai-custom/actions/course-game.action"

const initialState: GameLevelActionResult = {}

export function useCreateGameLevel(gameId: string, onSuccess?: () => void) {
  const onSuccessRef = useRef(onSuccess)
  onSuccessRef.current = onSuccess

  const boundAction = useCallback(
    createGameLevelAction.bind(null, gameId),
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
