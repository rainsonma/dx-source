// dx-web/src/features/web/home/components/faq-section.tsx
"use client";

import Link from "next/link";
import { motion } from "motion/react";
import { ArrowRight } from "lucide-react";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { revealVariants } from "@/features/web/home/helpers/motion-presets";

interface Faq {
  q: string;
  a: string;
  href: string;
}

const FAQS: Faq[] = [
  {
    q: "斗学适合什么水平的学习者？",
    a: "从零基础到雅思/托福备考都可以：课程按 CEFR 难度分级，多种玩法覆盖听说读写，每天 10 分钟就能坚持。",
    href: "/docs/getting-started/what-is-douxue",
  },
  {
    q: "能量豆是什么？怎么获得和消耗？",
    a: "能量豆用于使用 AI 随心学等高级功能。注册赠送、每日登录、购买、或会员月度赠送都可获得；AI 生成失败会全额退还。",
    href: "/docs/membership/beans-monthly",
  },
  {
    q: "免费用户能玩到哪些内容？",
    a: "免费用户可以玩每个课程的首关、使用基础游戏模式、加入少量学习群。付费后解锁全部关卡与 PK、小组、AI 等能力。",
    href: "/docs/membership/tiers-compare",
  },
  {
    q: "会员自动续费怎么关掉？",
    a: "在购买页或订单中心随时关闭自动续费，当前周期到期后不再扣款，已享权益保留到期。",
    href: "/docs/membership/purchase-flow",
  },
  {
    q: "未成年人可以用吗？需要家长同意吗？",
    a: "8 周岁以上可以使用；涉及付费功能需要监护人阅读并同意《监护人同意书》。我们对未成年人有专门的隐私与内容保护规则。",
    href: "/docs/account/guardian-consent",
  },
  {
    q: "支持微信支付和支付宝吗？",
    a: "支持。订单创建后会跳转到选定的支付方式，30 分钟内未支付订单自动失效。",
    href: "/docs/membership/purchase-flow",
  },
  {
    q: "我的学习数据会怎么保护？",
    a: "数据存储在中国境内，采用传输与存储加密。你可以随时查询、更正、删除或注销账号。完整条款见《隐私政策》。",
    href: "/docs/account/privacy-policy",
  },
  {
    q: "邀请返佣什么时候到账？",
    a: "永久会员邀请的新用户首次付费后，按 30% 比例计入你的返佣余额；完成攻略期与结算后按规则发放。",
    href: "/docs/invites/referral-program",
  },
  {
    q: "如何提交反馈或报告游戏问题？",
    a: "在「我的 → 提交反馈」页提交，支持多种类型；游戏内关卡问题可以直接从结算页上报。",
    href: "/docs/account/feedback",
  },
];

export function FaqSection() {
  return (
    <section className="w-full bg-gradient-to-b from-white to-slate-50 py-[80px] md:py-[100px]">
      <div className="mx-auto flex w-full max-w-[1280px] flex-col items-center gap-10 px-5 md:px-10 md:gap-12 lg:px-[120px]">
        <motion.div
          className="flex flex-col items-center gap-4 text-center"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.35 }}
        >
          <span className="text-sm font-semibold tracking-wide text-violet-500">
            常见问题
          </span>
          <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 md:text-4xl">
            开始之前，你可能想问
          </h2>
        </motion.div>

        <motion.div
          className="w-full max-w-[880px]"
          variants={revealVariants}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.15 }}
        >
          <Accordion type="single" collapsible defaultValue="faq-0">
            {FAQS.map((f, i) => (
              <AccordionItem key={f.q} value={`faq-${i}`}>
                <AccordionTrigger className="text-left text-[15px] font-semibold text-slate-900">
                  {f.q}
                </AccordionTrigger>
                <AccordionContent className="flex flex-col gap-3 text-sm leading-relaxed text-slate-600">
                  <p>{f.a}</p>
                  <Link
                    href={f.href}
                    className="inline-flex items-center gap-1 text-sm font-semibold text-teal-600 hover:text-teal-700"
                  >
                    查看完整说明
                    <ArrowRight className="h-4 w-4" />
                  </Link>
                </AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        </motion.div>
      </div>
    </section>
  );
}
