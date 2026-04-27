"use client"

import { MoreHorizontal } from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface Props {
  onEdit: () => void
  onDelete: () => void
  size?: "sm" | "md"
}

export function PostActionsMenu({ onEdit, onDelete, size = "md" }: Props) {
  const dim = size === "sm" ? "h-3.5 w-3.5" : "h-4 w-4"
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className="rounded-full p-1.5 text-muted-foreground hover:bg-muted"
          aria-label="more"
        >
          <MoreHorizontal className={dim} />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-28">
        <DropdownMenuItem onClick={onEdit}>编辑</DropdownMenuItem>
        <DropdownMenuItem onClick={onDelete} className="text-red-600 focus:text-red-600">
          删除
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
