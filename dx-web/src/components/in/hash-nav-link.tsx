"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useCallback, type MouseEvent, type ReactNode } from "react";

interface HashNavLinkProps {
  /** The hash target, e.g. "features". No leading "#". */
  hash: string;
  /** Page where the anchor lives. Default "/". */
  targetPath?: string;
  className?: string;
  onNavigate?: () => void;
  children: ReactNode;
}

/**
 * Nav link to an in-page anchor that always scrolls on click when already
 * on the target page — even if the URL hash is unchanged. When on a different
 * page, it falls through to Next.js Link navigation.
 */
export function HashNavLink({
  hash,
  targetPath = "/",
  className,
  onNavigate,
  children,
}: HashNavLinkProps) {
  const pathname = usePathname();
  const isOnTargetPage = pathname === targetPath;
  const href = isOnTargetPage ? `#${hash}` : `${targetPath}#${hash}`;

  const handleClick = useCallback(
    (e: MouseEvent<HTMLAnchorElement>) => {
      if (!isOnTargetPage) {
        onNavigate?.();
        return;
      }
      e.preventDefault();
      const el = document.getElementById(hash);
      if (el) {
        el.scrollIntoView({ behavior: "smooth", block: "start" });
        if (window.location.hash !== `#${hash}`) {
          history.replaceState(null, "", `#${hash}`);
        }
      }
      onNavigate?.();
    },
    [isOnTargetPage, hash, onNavigate],
  );

  return (
    <Link href={href} onClick={handleClick} className={className}>
      {children}
    </Link>
  );
}
