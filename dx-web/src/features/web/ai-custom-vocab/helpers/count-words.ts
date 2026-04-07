/** Count English words in text by splitting on whitespace */
export function countWords(text: string): number {
  return text.trim().split(/\s+/).filter(Boolean).length;
}
