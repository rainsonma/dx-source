import type { ComponentType } from "react";

export type LegalAgreementSlug =
  | "user-agreement"
  | "privacy-policy"
  | "guardian-consent"
  | "product-service";

export type LegalAgreement = {
  slug: LegalAgreementSlug;
  title: string;
  shortTitle: string;
  description: string;
  effectiveDate: string;
  lastUpdated: string;
  Component: ComponentType;
};
