import type { GameMode } from "@/consts/game-mode";

const CJK_REGEX = /[\u4e00-\u9fff]/;

export const MAX_METAS_PER_LEVEL = 20;
export const MAX_CONTENT_LENGTH = 600;

export function maxPairsForMode(mode: GameMode): number {
  switch (mode) {
    case "vocab-match": return 5;
    case "vocab-elimination": return 8;
    case "vocab-battle": return 20;
    default: return 5;
  }
}

export function vocabBatchSize(mode: GameMode): number {
  switch (mode) {
    case "vocab-match": return 5;
    case "vocab-elimination": return 8;
    default: return 0;
  }
}

export type VocabPair = {
  sourceData: string;
  translation: string;
};

function isChinese(line: string): boolean {
  return CJK_REGEX.test(line);
}

export type ParseVocabResult =
  | { ok: true; pairs: VocabPair[] }
  | { ok: false; error: string };

export function parseVocabText(raw: string, maxPairs: number, batchSize: number): ParseVocabResult {
  const lines = raw
    .split("\n")
    .map((l) => l.trim())
    .filter((l) => l.length > 0);

  if (lines.length === 0) {
    return { ok: false, error: "未解析到有效内容，请检查输入" };
  }

  const pairs: VocabPair[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    if (isChinese(line)) {
      return { ok: false, error: `第 ${i + 1} 行是中文但缺少对应的英文词汇` };
    }

    if (i + 1 >= lines.length || !isChinese(lines[i + 1])) {
      return { ok: false, error: `词汇「${line}」缺少中文释义，请确保每个英文词汇下方都有对应的中文释义` };
    }

    const oversized = line.length > MAX_CONTENT_LENGTH || lines[i + 1].length > MAX_CONTENT_LENGTH;
    if (oversized) {
      return { ok: false, error: `单条内容或翻译超过 ${MAX_CONTENT_LENGTH} 字符限制` };
    }

    pairs.push({
      sourceData: line,
      translation: lines[i + 1],
    });
    i += 2;
  }

  if (pairs.length > maxPairs) {
    return { ok: false, error: `词汇数量（${pairs.length}）超过当前模式上限 ${maxPairs} 对，请精简后重试` };
  }

  if (batchSize > 0 && pairs.length % batchSize !== 0) {
    return { ok: false, error: `词汇数量必须是 ${batchSize} 的倍数（当前 ${pairs.length} 条）` };
  }

  return { ok: true, pairs };
}
