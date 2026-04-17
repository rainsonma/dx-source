import { Fragment } from "react";
import { AgreementLink } from "./agreement-link";
import type { LegalAgreementSlug } from "@/features/com/legal/types";

type Props = {
  prefix: string;
  slugs: LegalAgreementSlug[];
  className?: string;
};

export function AgreementInlineList({ prefix, slugs, className }: Props) {
  return (
    <p className={className ?? "text-xs text-slate-400"}>
      {prefix}
      {slugs.map((slug, i) => (
        <Fragment key={slug}>
          <AgreementLink
            slug={slug}
            className="text-slate-500 hover:text-slate-700"
          />
          {i < slugs.length - 1 && "、"}
        </Fragment>
      ))}
    </p>
  );
}
