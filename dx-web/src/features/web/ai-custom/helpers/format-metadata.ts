const CJK_REGEX = /[\u4e00-\u9fff]/;

export const MAX_ENTRIES = 600;
export const MAX_CONTENT_LENGTH = 600;
export const MAX_SENTENCES = 20;
export const MAX_VOCAB = 200;
export const MAX_ITEMS_PER_META = 50;

export function countEnglishLines(text: string): number {
  return text
    .split("\n")
    .map((l) => l.trim())
    .filter((l) => l.length > 0 && !CJK_REGEX.test(l)).length;
}

export type MetadataEntry = {
  sourceData: string;
  translation?: string;
};

function isChinese(line: string): boolean {
  return CJK_REGEX.test(line);
}

export function parseMetadataText(raw: string): MetadataEntry[] {
  const lines = raw
    .split("\n")
    .map((l) => l.trim())
    .filter((l) => l.length > 0);

  const entries: MetadataEntry[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    if (isChinese(line)) {
      // Orphan Chinese line — skip
      i++;
      continue;
    }

    const entry: MetadataEntry = {
      sourceData: line,
    };

    if (i + 1 < lines.length && isChinese(lines[i + 1])) {
      entry.translation = lines[i + 1];
      i += 2;
    } else {
      i++;
    }

    entries.push(entry);
  }

  return entries;
}
