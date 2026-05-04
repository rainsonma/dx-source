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

// Splits a block of text into one sentence per line. Walks each line and breaks
// on `.`/`!`/`?` (with any trailing closing-quote/paren) followed by whitespace
// or end-of-line. Lines without a terminator are kept intact.
//
// Used when importing AI-generated stories into the manual textarea — the
// model is *supposed* to put one sentence per line but sometimes packs several
// onto one line, which would otherwise be saved as a single record.
export function splitIntoSentences(text: string): string {
  const lines = text
    .replace(/\r\n?/g, "\n")
    .split("\n")
    .map((l) => l.trim())
    .filter((l) => l.length > 0);

  const out: string[] = [];
  for (const line of lines) {
    const matches = line.match(/[^.!?]+[.!?]+["')\]]?(?=\s|$)/g);
    if (matches && matches.length > 0) {
      for (const m of matches) {
        const t = m.trim();
        if (t) out.push(t);
      }
    } else {
      out.push(line);
    }
  }
  return out.join("\n");
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
