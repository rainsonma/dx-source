"use client";

import { useState } from "react";
import { Copy, Check } from "lucide-react";

import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { SHARE_TABS } from "@/features/web/invite/helpers/share-snippets";

type ShareSnippetsModalProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  inviteUrl: string;
};

/** Modal displaying platform-specific share snippets with copy buttons */
export function ShareSnippetsModal({
  open,
  onOpenChange,
  inviteUrl,
}: ShareSnippetsModalProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton
        className="max-w-[520px] gap-0 rounded-[20px] border-none p-0"
      >
        <div className="flex flex-col gap-5 px-7 pt-7 pb-6">
          {/* Header */}
          <DialogTitle className="flex items-center gap-2.5 text-xl font-bold text-foreground">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              className="text-teal-600"
            >
              <circle cx="18" cy="5" r="3" />
              <circle cx="6" cy="12" r="3" />
              <circle cx="18" cy="19" r="3" />
              <line x1="8.59" x2="15.42" y1="13.51" y2="17.49" />
              <line x1="15.41" x2="8.59" y1="6.51" y2="10.49" />
            </svg>
            获取分享词
          </DialogTitle>
          <DialogDescription className="sr-only">
            选择平台并复制分享文案
          </DialogDescription>

          <div className="h-px bg-border" />

          {/* Tabs */}
          <Tabs defaultValue={SHARE_TABS[0].key}>
            <TabsList className="h-auto w-full bg-transparent p-0">
              {SHARE_TABS.map((tab) => (
                <TabsTrigger
                  key={tab.key}
                  value={tab.key}
                  className="rounded-lg px-4 py-2 text-[13px] font-medium text-muted-foreground data-[state=active]:bg-teal-600 data-[state=active]:font-semibold data-[state=active]:text-white data-[state=active]:shadow-none"
                >
                  {tab.label}
                </TabsTrigger>
              ))}
            </TabsList>

            {SHARE_TABS.map((tab) => (
              <TabsContent key={tab.key} value={tab.key}>
                <div className="flex max-h-[400px] flex-col gap-3 overflow-y-auto pr-1">
                  {tab.snippets.map((snippet, index) => (
                    <SnippetCard
                      key={index}
                      text={snippet.replace("{link}", inviteUrl)}
                    />
                  ))}
                </div>
              </TabsContent>
            ))}
          </Tabs>

          {/* Hint */}
          <p className="text-center text-xs text-muted-foreground">
            💡 复制后可直接粘贴到对应平台发布
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
}

/** Individual snippet card with copy button */
function SnippetCard({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  /** Copy snippet text to clipboard */
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Silently fail if clipboard access is denied
    }
  };

  return (
    <div className="flex flex-col gap-2.5 rounded-xl border border-border bg-muted px-4 py-3.5">
      <p className="whitespace-pre-line text-[13px] leading-relaxed text-foreground">
        {text}
      </p>
      <div className="flex">
        <button
          type="button"
          onClick={handleCopy}
          className="flex items-center gap-1 rounded-md bg-teal-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-teal-700"
        >
          {copied ? (
            <>
              <Check className="h-3 w-3" />
              已复制
            </>
          ) : (
            <>
              <Copy className="h-3 w-3" />
              复制
            </>
          )}
        </button>
      </div>
    </div>
  );
}
