const ALLOWED = new Set([
  'message-circle-more', 'swords', 'bell', 'megaphone', 'trophy', 'gift',
  'rocket', 'star', 'shield', 'book-open', 'calendar', 'user-plus',
  'heart', 'zap', 'party-popper', 'info', 'alert-triangle', 'check-circle-2',
  'sparkles', 'crown',
])

export function resolveNoticeIcon(name: string | null | undefined): string {
  if (!name) return 'message-circle-more'
  return ALLOWED.has(name) ? name : 'message-circle-more'
}
