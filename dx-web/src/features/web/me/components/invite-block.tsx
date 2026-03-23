"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { ChevronRight, Copy, Check } from "lucide-react";

/** Invite code block with copy button, invite URL, and referral link */
export function InviteBlock({ inviteCode }: { inviteCode: string }) {
  const [codeCopied, setCodeCopied] = useState(false);
  const [urlCopied, setUrlCopied] = useState(false);
  const [inviteUrl, setInviteUrl] = useState(`/invite/${inviteCode}`);

  useEffect(() => {
    setInviteUrl(`${window.location.origin}/invite/${inviteCode}`);
  }, [inviteCode]);

  async function handleCopyCode() {
    await navigator.clipboard.writeText(inviteCode);
    setCodeCopied(true);
    setTimeout(() => setCodeCopied(false), 2000);
  }

  async function handleCopyUrl() {
    await navigator.clipboard.writeText(inviteUrl);
    setUrlCopied(true);
    setTimeout(() => setUrlCopied(false), 2000);
  }

  return (
    <div className="rounded-2xl border border-border bg-card p-6">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-base font-bold text-foreground">邀请推广</h3>
        <Link
          href="/hall/invite"
          className="flex items-center gap-1 text-sm font-medium text-teal-600 hover:text-teal-700"
        >
          去推广
          <ChevronRight className="h-4 w-4" />
        </Link>
      </div>

      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-1">
          <span className="text-xs text-muted-foreground">邀请码</span>
          <div className="flex items-center gap-3">
            <span className="rounded-lg bg-muted px-4 py-2 font-mono text-lg font-bold tracking-widest text-foreground">
              {inviteCode}
            </span>
            <button
              onClick={handleCopyCode}
              className="flex items-center gap-1 text-sm text-teal-600 hover:text-teal-700"
            >
              {codeCopied ? (
                <>
                  <Check className="h-4 w-4" />
                  已复制
                </>
              ) : (
                <>
                  <Copy className="h-4 w-4" />
                  复制
                </>
              )}
            </button>
          </div>
        </div>

        <div className="flex flex-col gap-1">
          <span className="text-xs text-muted-foreground">邀请链接</span>
          <div className="flex items-center gap-3">
            <span className="truncate rounded-lg bg-muted px-4 py-2 text-sm text-muted-foreground">
              {inviteUrl}
            </span>
            <button
              onClick={handleCopyUrl}
              className="flex shrink-0 items-center gap-1 text-sm text-teal-600 hover:text-teal-700"
            >
              {urlCopied ? (
                <>
                  <Check className="h-4 w-4" />
                  已复制
                </>
              ) : (
                <>
                  <Copy className="h-4 w-4" />
                  复制
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
