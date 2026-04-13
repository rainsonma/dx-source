import Link from "next/link";
import { ExternalLink } from "lucide-react";
import type { ReactNode } from "react";

type Props = {
  href: string;
  external?: boolean;
  children: ReactNode;
};

export function DocLink({ href, external, children }: Props) {
  const className =
    "inline-flex items-center gap-1 text-teal-600 underline underline-offset-2 hover:text-teal-700";
  if (external) {
    return (
      <a
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        className={className}
      >
        {children}
        <ExternalLink className="h-3.5 w-3.5" aria-hidden="true" />
      </a>
    );
  }
  return (
    <Link href={href} className={className}>
      {children}
    </Link>
  );
}
