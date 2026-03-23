export type PublicGameCard = {
  id: string
  name: string
  description: string | null
  mode: string
  createdAt: Date
  cover: { url: string } | null
  user: { username: string } | null
  category: { name: string } | null
  _count: { levels: number }
}
