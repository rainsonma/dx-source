import type { PublicGameCard } from "@/features/web/games/actions/game.action"

/** Map Go API flat GameCardData to the nested PublicGameCard shape */
export function toPublicGameCard(item: any): PublicGameCard {
  return {
    id: item.id,
    name: item.name,
    description: item.description ?? null,
    mode: item.mode,
    createdAt: new Date(item.createdAt),
    cover: item.coverUrl ? { url: item.coverUrl } : null,
    user: item.author ? { username: item.author } : null,
    category: item.categoryName ? { name: item.categoryName } : null,
    _count: { levels: item.levelCount ?? 0 },
  }
}
