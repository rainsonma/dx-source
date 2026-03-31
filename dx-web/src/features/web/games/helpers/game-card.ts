import type { PublicGameCard } from "@/features/web/games/actions/game.action"

/** Map Go API flat GameCardData to the nested PublicGameCard shape */
export function toPublicGameCard(item: Record<string, unknown>): PublicGameCard {
  return {
    id: item.id as string,
    name: item.name as string,
    description: (item.description as string | null) ?? null,
    mode: item.mode as string,
    createdAt: new Date(item.createdAt as string),
    cover: item.coverUrl ? { url: item.coverUrl as string } : null,
    user: item.author ? { username: item.author as string } : null,
    category: item.categoryName ? { name: item.categoryName as string } : null,
    _count: { levels: (item.levelCount as number) ?? 0 },
  }
}
