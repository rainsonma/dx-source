"use client";

import { useState, useTransition } from "react";
import { PenLine, X, Plus, Trash2, Save, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { VisuallyHidden } from "@radix-ui/react-visually-hidden";
import { toast } from "sonner";
import type { ContentVocabData, VocabInput, PosKey, DefinitionEntry } from "@/lib/api-client";
import { updateVocabAction } from "@/features/web/ai-custom/actions/content-vocab.action";

const ALL_POS: PosKey[] = ["n", "v", "adj", "adv", "prep", "conj", "pron", "art", "num", "int", "aux", "det"];
const POS_LABELS: Record<PosKey, string> = {
  n: "名词 n", v: "动词 v", adj: "形容词 adj", adv: "副词 adv",
  prep: "介词 prep", conj: "连词 conj", pron: "代词 pron", art: "冠词 art",
  num: "数词 num", int: "感叹词 int", aux: "助动词 aux", det: "限定词 det",
};

function parseDefinition(raw: string | null | undefined): Array<{ pos: PosKey; gloss: string }> {
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return (parsed as DefinitionEntry[]).flatMap((entry) =>
      Object.entries(entry).map(([pos, gloss]) => ({ pos: pos as PosKey, gloss: gloss as string }))
    );
  } catch {
    return [];
  }
}

type PosRow = { pos: PosKey; gloss: string };

type VocabEditDialogProps = {
  vocab: ContentVocabData;
  onClose: () => void;
  onSaved: (updated: ContentVocabData) => void;
};

export function VocabEditDialog({ vocab, onClose, onSaved }: VocabEditDialogProps) {
  const [content, setContent] = useState(vocab.content);
  const [posRows, setPosRows] = useState<PosRow[]>(() => {
    const parsed = parseDefinition(vocab.definition);
    return parsed.length > 0 ? parsed : [{ pos: "n", gloss: "" }];
  });
  const [ukPhonetic, setUkPhonetic] = useState(vocab.ukPhonetic ?? "");
  const [usPhonetic, setUsPhonetic] = useState(vocab.usPhonetic ?? "");
  const [ukAudioUrl, setUkAudioUrl] = useState(vocab.ukAudioUrl ?? "");
  const [usAudioUrl, setUsAudioUrl] = useState(vocab.usAudioUrl ?? "");
  const [explanation, setExplanation] = useState(vocab.explanation ?? "");
  const [isPending, startTransition] = useTransition();

  function addPosRow() {
    setPosRows((prev) => [...prev, { pos: "n", gloss: "" }]);
  }

  function removePosRow(index: number) {
    setPosRows((prev) => prev.filter((_, i) => i !== index));
  }

  function updatePosRow(index: number, field: "pos" | "gloss", value: string) {
    setPosRows((prev) => prev.map((row, i) => (i === index ? { ...row, [field]: value } : row)));
  }

  function handleSave() {
    const trimmedContent = content.trim();
    if (!trimmedContent) { toast.error("词条内容不能为空"); return; }

    const validDefs = posRows.filter((r) => r.gloss.trim().length > 0);
    const input: VocabInput = {
      content: trimmedContent,
      definition: validDefs.map((r) => ({ [r.pos]: r.gloss.trim() } as DefinitionEntry)),
      ukPhonetic: ukPhonetic.trim() || null,
      usPhonetic: usPhonetic.trim() || null,
      ukAudioUrl: ukAudioUrl.trim() || null,
      usAudioUrl: usAudioUrl.trim() || null,
      explanation: explanation.trim() || null,
    };

    startTransition(async () => {
      const res = await updateVocabAction(vocab.id, input);
      if (res.code !== 0) {
        if (res.code === 40302 || res.message.toLowerCase().includes("duplicate")) {
          toast.error("该词汇已存在于您的词库");
        } else {
          toast.error(res.message);
        }
        return;
      }
      if (res.data) {
        toast.success("词条已更新");
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
        <VisuallyHidden><DialogTitle>编辑词条</DialogTitle></VisuallyHidden>

        <div className="flex flex-col max-h-[90vh] overflow-y-auto">
          <div className="flex shrink-0 items-center justify-between px-6 py-4">
            <div className="flex items-center gap-2.5">
              <PenLine className="h-5 w-5 text-blue-600" />
              <h2 className="text-lg font-bold text-foreground">
                编辑 <span className="text-blue-600">{vocab.content}</span>
              </h2>
            </div>
            <button type="button" onClick={onClose} aria-label="关闭"
              className="flex h-7 w-7 items-center justify-center rounded-lg bg-muted">
              <X className="h-3.5 w-3.5 text-muted-foreground" />
            </button>
          </div>

          <div className="px-6 pb-6 flex flex-col gap-5">
            <section>
              <p className="mb-2 text-sm font-semibold text-foreground">词条</p>
              <input type="text" value={content} onChange={(e) => setContent(e.target.value)}
                placeholder="词条内容"
                className="h-9 w-full rounded-lg border border-border bg-muted/50 px-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
            </section>

            <section>
              <div className="mb-2 flex items-center justify-between">
                <p className="text-sm font-semibold text-foreground">释义</p>
                <button type="button" onClick={addPosRow}
                  className="flex items-center gap-1 rounded-lg bg-teal-50 px-2 py-1 text-xs font-semibold text-teal-700 hover:bg-teal-100">
                  <Plus className="h-3 w-3" />添加词性
                </button>
              </div>
              <div className="flex flex-col gap-2">
                {posRows.map((row, i) => (
                  <div key={i} className="flex items-center gap-2">
                    <select value={row.pos} onChange={(e) => updatePosRow(i, "pos", e.target.value)}
                      className="h-9 rounded-lg border border-border bg-muted px-2 text-xs font-medium text-foreground focus:outline-none focus:ring-1 focus:ring-blue-500">
                      {ALL_POS.map((pos) => <option key={pos} value={pos}>{POS_LABELS[pos]}</option>)}
                    </select>
                    <input type="text" value={row.gloss} onChange={(e) => updatePosRow(i, "gloss", e.target.value)}
                      placeholder="释义"
                      className="h-9 flex-1 rounded-lg border border-border bg-muted/50 px-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
                    {posRows.length > 1 && (
                      <button type="button" onClick={() => removePosRow(i)}
                        className="flex h-7 w-7 items-center justify-center rounded-lg bg-red-50 text-red-500 hover:bg-red-100">
                        <Trash2 className="h-3 w-3" />
                      </button>
                    )}
                  </div>
                ))}
              </div>
            </section>

            <section>
              <p className="mb-2 text-sm font-semibold text-foreground">音标</p>
              <div className="flex flex-col gap-2">
                {(["UK", "US"] as const).map((region) => (
                  <div key={region} className="flex items-center gap-2">
                    <span className="w-8 text-xs font-medium text-foreground">{region}</span>
                    <input type="text"
                      value={region === "UK" ? ukPhonetic : usPhonetic}
                      onChange={(e) => region === "UK" ? setUkPhonetic(e.target.value) : setUsPhonetic(e.target.value)}
                      placeholder="/fæst/"
                      className="h-9 flex-1 rounded-lg border border-border bg-muted/50 px-3 font-mono text-sm text-foreground placeholder:text-muted-foreground focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
                  </div>
                ))}
              </div>
            </section>

            <section>
              <p className="mb-2 text-sm font-semibold text-foreground">音频地址</p>
              <div className="flex flex-col gap-2">
                {(["UK", "US"] as const).map((region) => (
                  <div key={region} className="flex items-center gap-2">
                    <span className="w-8 text-xs font-medium text-foreground">{region}</span>
                    <input type="text"
                      value={region === "UK" ? ukAudioUrl : usAudioUrl}
                      onChange={(e) => region === "UK" ? setUkAudioUrl(e.target.value) : setUsAudioUrl(e.target.value)}
                      placeholder="/audio/uk/fast.mp3"
                      className="h-9 flex-1 rounded-lg border border-border bg-muted/50 px-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
                  </div>
                ))}
              </div>
            </section>

            <section>
              <p className="mb-2 text-sm font-semibold text-foreground">说明 / 用法</p>
              <textarea value={explanation} onChange={(e) => setExplanation(e.target.value)}
                placeholder="用法说明、例句等" rows={3}
                className="w-full resize-none rounded-xl border border-border bg-muted/50 px-3 py-2.5 text-sm text-foreground placeholder:text-muted-foreground focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
            </section>

            <div className="flex justify-end">
              <div className="flex overflow-hidden rounded-xl border border-border">
                <button type="button" onClick={onClose}
                  className="flex h-11 items-center justify-center gap-1.5 border-r border-border bg-muted px-4">
                  <span className="text-xs font-semibold text-muted-foreground">取消</span>
                </button>
                <button type="button" disabled={isPending} onClick={handleSave}
                  className="flex h-11 items-center justify-center gap-1.5 bg-blue-600 px-5 disabled:opacity-50">
                  {isPending ? <Loader2 className="h-4 w-4 animate-spin text-white" /> : <Save className="h-4 w-4 text-white" />}
                  <span className="text-sm font-semibold text-white">保存</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
