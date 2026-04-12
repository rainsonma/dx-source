import useSWR from "swr"
import type { HallMenuSection } from "@/features/web/hall/types/hall-menu.types"

export function useHallMenu() {
  return useSWR<HallMenuSection[]>("/api/hall/menus")
}

export function useHallMenuItem(href: string) {
  const { data: sections } = useHallMenu()
  if (!sections) return null
  for (const section of sections) {
    const item = section.items.find((i) => i.href === href)
    if (item) return item
  }
  return null
}
