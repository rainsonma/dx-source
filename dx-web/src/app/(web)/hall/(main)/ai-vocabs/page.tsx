"use client";

import { useState } from "react";
import { Search, Sparkles, BookOpen } from "lucide-react";
import { useHallMenuItem } from "@/features/web/hall/hooks/use-hall-menu";
import { PageTopBar } from "@/features/web/hall/components/page-top-bar";
import { VocabList } from "@/features/web/ai-vocabs/components/vocab-list";
import { AddVocabFromAiDialog } from "@/features/web/ai-vocabs/components/add-vocab-from-ai-dialog";
import { AddVocabManualDialog } from "@/features/web/ai-vocabs/components/add-vocab-manual-dialog";

export default function AiVocabsPage() {
  const menu = useHallMenuItem("/hall/ai-vocabs");
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [aiDialogOpen, setAiDialogOpen] = useState(false);
  const [manualDialogOpen, setManualDialogOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  function handleSearchChange(value: string) {
    setSearch(value);
    clearTimeout((handleSearchChange as unknown as { timer?: ReturnType<typeof setTimeout> }).timer);
    const timer = setTimeout(() => setDebouncedSearch(value), 300);
    (handleSearchChange as unknown as { timer?: ReturnType<typeof setTimeout> }).timer = timer;
  }

  function handleAdded() {
    setRefreshKey((k) => k + 1);
  }

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title={menu?.label ?? "AI 词汇库"}
        subtitle={menu?.subtitle ?? ""}
      />

      {/* Title bar with controls */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="relative flex-1 sm:max-w-xs">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            value={search}
            onChange={(e) => handleSearchChange(e.target.value)}
            placeholder="搜索词汇..."
            className="h-9 w-full rounded-lg border border-border bg-muted/50 pl-9 pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
          />
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setAiDialogOpen(true)}
            className="flex items-center gap-1.5 rounded-lg bg-violet-600 px-3 py-2 text-sm font-semibold text-white hover:bg-violet-700"
          >
            <Sparkles className="h-3.5 w-3.5" />
            AI 生成
          </button>
          <button
            type="button"
            onClick={() => setManualDialogOpen(true)}
            className="flex items-center gap-1.5 rounded-lg bg-teal-600 px-3 py-2 text-sm font-semibold text-white hover:bg-teal-700"
          >
            <BookOpen className="h-3.5 w-3.5" />
            手动添加
          </button>
        </div>
      </div>

      <div className="flex flex-1 flex-col overflow-hidden rounded-[14px] border border-border bg-card">
        <div className="flex-1 overflow-y-auto px-4 pb-4 pt-3 lg:px-5 lg:pb-5">
          <VocabList search={debouncedSearch} refreshKey={refreshKey} />
        </div>
      </div>

      <AddVocabFromAiDialog
        open={aiDialogOpen}
        onOpenChange={setAiDialogOpen}
        onAdded={handleAdded}
      />
      <AddVocabManualDialog
        open={manualDialogOpen}
        onOpenChange={setManualDialogOpen}
        onAdded={handleAdded}
      />
    </div>
  );
}
