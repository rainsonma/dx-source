import { UserAgreementDoc } from "./documents/user-agreement";
import { PrivacyPolicyDoc } from "./documents/privacy-policy";
import { GuardianConsentDoc } from "./documents/guardian-consent";
import { ProductServiceDoc } from "./documents/product-service";
import { EFFECTIVE_DATE, LAST_UPDATED } from "./constants";
import type { LegalAgreement, LegalAgreementSlug } from "./types";

export {
  UserAgreementDoc,
  PrivacyPolicyDoc,
  GuardianConsentDoc,
  ProductServiceDoc,
};

export const LEGAL_AGREEMENTS: LegalAgreement[] = [
  {
    slug: "user-agreement",
    title: "用户协议",
    shortTitle: "《用户协议》",
    description:
      "斗学账号注册、账号管理、用户权责、知识产权、免责与注销等完整条款。",
    effectiveDate: EFFECTIVE_DATE,
    lastUpdated: LAST_UPDATED,
    Component: UserAgreementDoc,
  },
  {
    slug: "privacy-policy",
    title: "隐私政策",
    shortTitle: "《隐私政策》",
    description:
      "我们如何收集、使用、存储、共享、保护您的个人信息，以及您的权利行使方式。",
    effectiveDate: EFFECTIVE_DATE,
    lastUpdated: LAST_UPDATED,
    Component: PrivacyPolicyDoc,
  },
  {
    slug: "guardian-consent",
    title: "监护人同意书",
    shortTitle: "《监护人同意书》",
    description:
      "未成年人使用斗学前，监护人需知情并同意的相关条款与权责说明。",
    effectiveDate: EFFECTIVE_DATE,
    lastUpdated: LAST_UPDATED,
    Component: GuardianConsentDoc,
  },
  {
    slug: "product-service",
    title: "产品服务协议",
    shortTitle: "《产品服务协议》",
    description:
      "会员订阅与支付、服务暂停与终止、退款规则、知识产权及争议解决等条款。",
    effectiveDate: EFFECTIVE_DATE,
    lastUpdated: LAST_UPDATED,
    Component: ProductServiceDoc,
  },
];

export function getAgreementBySlug(slug: LegalAgreementSlug): LegalAgreement {
  const found = LEGAL_AGREEMENTS.find((a) => a.slug === slug);
  if (!found) {
    throw new Error(`[legal] unknown agreement slug: ${slug}`);
  }
  return found;
}
