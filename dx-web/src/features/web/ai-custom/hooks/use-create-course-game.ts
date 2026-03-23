"use client"

import { useActionState, useState, useEffect, useRef } from "react"
import { swrMutate } from "@/lib/swr"

import {
  createCourseGameAction,
  type CreateCourseGameResult,
} from "@/features/web/ai-custom/actions/course-game.action"

const initialState: CreateCourseGameResult = {}

export function useCreateCourseGame(onSuccess?: () => void) {
  const [coverId, setCoverId] = useState<string | null>(null)
  const onSuccessRef = useRef(onSuccess)
  onSuccessRef.current = onSuccess

  const [state, formAction, isPending] = useActionState(
    createCourseGameAction,
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
    coverId,
    setCoverId,
  }
}
