"use client";

import { useState, useTransition } from "react";
import {
  Sparkles,
  X,
  Plus,
  Trash2,
  Save,
  Loader2,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { VisuallyHidden } from "@radix-ui/react-visually-hidden";
import { toast } from "sonner";
import type {
  ContentVocabData,
  ContentVocabComplementPatch,
  PosKey,
  DefinitionEntry,
} from "@/lib/api-client";
import { complementVocabAction } from "@/features/web/ai-custom/actions/content-vocab.action";

const ALL_POS: PosKey[] = ["n", "v", "adj", "adv", "prep", "conj", "pron", "art", "num", "int", "aux", "det"];
const POS_LABELS: Record<PosKey, string> = {
  n: "名词 n",
  v: "动词 v",
  adj: "形容词 adj",
  adv: "副词 adv",
  prep: "介词 prep",
  conj: "连词 conj",
  pron: "代词 pron",
  art: "冠词 art",
  num: "数词 num",
  int: "感叹词 int",
  aux: "助动词 aux",
  det: "限定词 det",
};

function parseDefinition(raw: string | null | undefined): DefinitionEntry[] {
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? (parsed as DefinitionEntry[]) : [];
  } catch {
    return [];
  }
}

type NewPosRow = {
  pos: PosKey;
  gloss: string;
};

type ComplementVocabDialogProps = {
  vocab: ContentVocabData;
  onClose: () => void;
  onSaved: (updated: ContentVocabData) => void;
};

export function ComplementVocabDialog({ vocab, onClose, onSaved }: ComplementVocabDialogProps) {
  const existingDefs = parseDefinition(vocab.definition);

  const [newPosRows, setNewPosRows] = useState<NewPosRow[]>([{ pos: "n", gloss: "" }]);
  const [ukPhonetic, setUkPhonetic] = useState("");
  const [usPhonetic, setUsPhonetic] = useState("");
  const [ukAudioUrl, setUkAudioUrl] = useState("");
  const [usAudioUrl, setUsAudioUrl] = useState("");
  const [explanation, setExplanation] = useState("");
  const [isPending, startTransition] = useTransition();

  function addPosRow() {
    setNewPosRows((prev) => [...prev, { pos: "n", gloss: "" }]);
  }

  function removePosRow(index: number) {
    setNewPosRows((prev) => prev.filter((_, i) => i !== index));
  }

  function updatePosRow(index: number, field: "pos" | "gloss", value: string) {
    setNewPosRows((prev) =>
      prev.map((row, i) => (i === index ? { ...row, [field]: value } : row))
    );
  }

  function handleSave() {
    const patch: ContentVocabComplementPatch = {};
    let changed = false;
    const summary: string[] = [];

    // New POS definition entries (only non-empty rows)
    const validNewDefs = newPosRows.filter((r) => r.gloss.trim().length > 0);
    if (validNewDefs.length > 0) {
      patch.definition = validNewDefs.map((r) => ({ [r.pos]: r.gloss.trim() } as DefinitionEntry));
      changed = true;
      summary.push(`添加了 ${validNewDefs.length} 条释义`);
    }

    // Only include null fields
    if (!vocab.ukPhonetic && ukPhonetic.trim()) {
      patch.ukPhonetic = ukPhonetic.trim();
      changed = true;
      summary.push("补全了英式音标");
    }
    if (!vocab.usPhonetic && usPhonetic.trim()) {
      patch.usPhonetic = usPhonetic.trim();
      changed = true;
      summary.push("补全了美式音标");
    }
    if (!vocab.ukAudioUrl && ukAudioUrl.trim()) {
      patch.ukAudioUrl = ukAudioUrl.trim();
      changed = true;
      summary.push("补全了英式音频");
    }
    if (!vocab.usAudioUrl && usAudioUrl.trim()) {
      patch.usAudioUrl = usAudioUrl.trim();
      changed = true;
      summary.push("补全了美式音频");
    }
    if (!vocab.explanation && explanation.trim()) {
      patch.explanation = explanation.trim();
      changed = true;
      summary.push("补全了说明");
    }

    if (!changed) {
      toast.info("没有需要补全的内容");
      return;
    }

    startTransition(async () => {
      const res = await complementVocabAction(vocab.id, patch);
      if (res.code !== 0) {
        toast.error(res.message);
        return;
      }
      if (res.data) {
        toast.success(summary.join("；"));
        onSaved(res.data);
      }
    });
  }

  return (
    <Dialog open onOpenChange={(open) => { if (!open) onClose(); }}>
      <DialogContent
        aria-describedby={undefined}
        showCloseButton={false}
        className="sm:max-w-lg overflow-hidden rounded-[20px] border-0 p-0 shadow-[0_12px_40px_rgba(15,23,42,0.19)]"
      >
        <VisuallyHidden>
          <DialogTitle>补全词汇字段</DialogTitle>
        </VisuallyHidden>

        <div className="flex flex-col max-h-[90vh] overflow-y-auto">
          {/* Header */}
          <div className="flex shrink-0 items-center justify-between px-6 py-4">
            <div className="flex items-center gap-2.5">
              <Sparkles className="h-5 w-5 text-violet-600" />
              <h2 className="text-lg font-bold text-foreground">
                补全 <span className="text-violet-600">{vocab.content}</span>
              </h2>
            </div>
            <button
              type="button"
              onClick={onClose}
              aria-label="关闭"
              className="flex h-7 w-7 items-center justify-center rounded-lg bg-muted"
            >
              <X className="h-3.5 w-3.5 text-muted-foreground" />
            </button>
          </div>

          <div className="px-6 pb-6 flex flex-col gap-5">
            {/* Existing definition entries (read-only) */}
            <section>
              <p className="mb-2 text-sm font-semibold text-foreground">现有释义</p>
              {existingDefs.length === 0 ? (
                <p className="text-xs text-muted-foreground">暂无</p>
              ) : (
                <div className="flex flex-wrap gap-1.5">
                  {existingDefs.map((entry, i) =>
                    Object.entries(entry).map(([pos, gloss]) => (
                      <span
                        key={`${i}-${pos}`}
                        className="flex items-center gap-1 rounded-full bg-teal-50 px-2.5 py-0.5 text-xs"
                      >
                        <span className="font-semibold text-teal-700">{pos}</span>
                        <span className="text-muted-foreground">{gloss as string}</span>
                      </span>
                    ))
                  )}
                </div>
              )}
            </section>

            {/* New definition rows */}
            <section>
              <div className="mb-2 flex items-center justify-between">
                <p className="text-sm font-semibold text-foreground">新增释义</p>
                <button
                  type="button"
                  onClick={addPosRow}
                  className="flex items-center gap-1 rounded-lg bg-teal-50 px-2 py-1 text-xs font-semibold text-teal-700 hover:bg-teal-100"
                >
                  <Plus className="h-3 w-3" />
                  添加词性
                </button>
              </div>
              <div className="flex flex-col gap-2">
                {newPosRows.map((row, i) => (
                  <div key={i} className="flex items-center gap-2">
                    <select
                      value={row.pos}
                      onChange={(e) => updatePosRow(i, "pos", e.target.value)}
                      className="h-9 rounded-lg border border-border bg-muted px-2 text-xs font-medium text-foreground focus:outline-none focus:ring-1 focus:ring-teal-500"
                    >
                      {ALL_POS.map((pos) => (
                        <option key={pos} value={pos}>{POS_LABELS[pos]}</option>
                      ))}
                    </select>
                    <input
                      type="text"
                      value={row.gloss}
                      onChange={(e) => updatePosRow(i, "gloss", e.target.value)}
                      placeholder="释义"
                      className="h-9 flex-1 rounded-lg border border-border bg-muted/50 px-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
                    />
                    {newPosRows.length > 1 && (
                      <button
                        type="button"
                        onClick={() => removePosRow(i)}
                        className="flex h-7 w-7 items-center justify-center rounded-lg bg-red-50 text-red-500 hover:bg-red-100"
                      >
                        <Trash2 className="h-3 w-3" />
                      </button>
                    )}
                  </div>
                ))}
              </div>
            </section>

            {/* Phonetics — only show if null */}
            {(!vocab.ukPhonetic || !vocab.usPhonetic) && (
              <section>
                <p className="mb-2 text-sm font-semibold text-foreground">音标</p>
                <div className="flex flex-col gap-2">
                  {!vocab.ukPhonetic && (
                    <div className="flex items-center gap-2">
                      <span className="w-8 text-xs font-medium text-foreground">UK</span>
                      <input
                        type="text"
                        value={ukPhonetic}
                        onChange={(e) => setUkPhonetic(e.target.value)}
                        placeholder="fæst"
                        className="h-9 flex-1 rounded-lg border border-border bg-muted/50 px-3 font-mono text-sm text-foreground placeholder:text-muted-foreground focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
                      />
                    </div>
                  )}
                  {!vocab.usPhonetic && (
                    <div className="flex items-center gap-2">
                      <span className="w-8 text-xs font-medium text-foreground">US</span>
                      <input
                        type="text"
                        value={usPhonetic}
                        onChange={(e) => setUsPhonetic(e.target.value)}
                        placeholder="fæst"
                        className="h-9 flex-1 rounded-lg border border-border bg-muted/50 px-3 font-mono text-sm text-foreground placeholder:text-muted-foreground focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
                      />
                    </div>
                  )}
                </div>
                {(vocab.ukPhonetic || vocab.usPhonetic) && (
                  <p className="mt-1.5 text-xs text-muted-foreground">
                    已有音标：
                    {vocab.ukPhonetic && <span className="ml-1 text-foreground">UK /{vocab.ukPhonetic}/</span>}
                    {vocab.usPhonetic && <span className="ml-1 text-foreground">US /{vocab.usPhonetic}/</span>}
                    （不可覆盖）
                  </p>
                )}
              </section>
            )}

            {/* Audio URLs — only show if null */}
            {(!vocab.ukAudioUrl || !vocab.usAudioUrl) && (
              <section>
                <p className="mb-2 text-sm font-semibold text-foreground">音频地址</p>
                <div className="flex flex-col gap-2">
                  {!vocab.ukAudioUrl && (
                    <div className="flex items-center gap-2">
                      <span className="w-8 text-xs font-medium text-foreground">UK</span>
                      <input
                        type="text"
                        value={ukAudioUrl}
                        onChange={(e) => setUkAudioUrl(e.target.value)}
                        placeholder="/audio/uk/fast.mp3"
                        className="h-9 flex-1 rounded-lg border border-border bg-muted/50 px-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
                      />
                    </div>
                  )}
                  {!vocab.usAudioUrl && (
                    <div className="flex items-center gap-2">
                      <span className="w-8 text-xs font-medium text-foreground">US</span>
                      <input
                        type="text"
                        value={usAudioUrl}
                        onChange={(e) => setUsAudioUrl(e.target.value)}
                        placeholder="/audio/us/fast.mp3"
                        className="h-9 flex-1 rounded-lg border border-border bg-muted/50 px-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
                      />
                    </div>
                  )}
                </div>
              </section>
            )}

            {/* Explanation — only show if null */}
            {!vocab.explanation && (
              <section>
                <p className="mb-2 text-sm font-semibold text-foreground">说明 / 用法</p>
                <textarea
                  value={explanation}
                  onChange={(e) => setExplanation(e.target.value)}
                  placeholder="用法说明、例句等"
                  rows={3}
                  className="w-full resize-none rounded-xl border border-border bg-muted/50 px-3 py-2.5 text-sm text-foreground placeholder:text-muted-foreground focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
                />
              </section>
            )}

            {/* Footer */}
            <div className="flex justify-end">
              <div className="flex overflow-hidden rounded-xl border border-border">
                <button
                  type="button"
                  onClick={onClose}
                  className="flex h-11 items-center justify-center gap-1.5 border-r border-border bg-muted px-4"
                >
                  <span className="text-xs font-semibold text-muted-foreground">取消</span>
                </button>
                <button
                  type="button"
                  disabled={isPending}
                  onClick={handleSave}
                  className="flex h-11 items-center justify-center gap-1.5 bg-teal-600 px-5 disabled:opacity-50"
                >
                  {isPending ? (
                    <Loader2 className="h-4 w-4 animate-spin text-white" />
                  ) : (
                    <Save className="h-4 w-4 text-white" />
                  )}
                  <span className="text-sm font-semibold text-white">保存补全</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
