"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { PricingCard } from "@/features/web/purchase/components/pricing-card";
import { USER_GRADE_PRICES } from "@/consts/user-grade";
import { PAYMENT_METHODS } from "@/consts/payment-method";
import { orderApi } from "@/lib/api-client";

const plans = [
  {
    grade: "free",
    name: "免费会员",
    price: `¥${USER_GRADE_PRICES.free}`,
    period: "",
    features: [
      "免费课程内容",
      "免费游戏试玩",
      "少量学习小组",
      "分享推广佣金",
    ],
    bgColor: "bg-slate-500",
    borderColor: "#475569",
    ctaText: "当前方案",
  },
  {
    grade: "month",
    name: "月度会员",
    price: `¥${USER_GRADE_PRICES.month}`,
    period: "/月",
    features: [
      "全部课程内容",
      "全部游戏畅玩",
      "AI - 智能助力",
      "高级音效发音",
      "更多学习小组",
      "分享推广佣金",
    ],
    bgColor: "bg-blue-500",
    borderColor: "#2563EB",
    ctaText: "立即订阅",
  },
  {
    grade: "season",
    name: "季度会员",
    price: `¥${USER_GRADE_PRICES.season}`,
    period: "/季",
    features: [
      "全部课程内容",
      "全部游戏畅玩",
      "AI - 智能助力",
      "高级音效发音",
      "更多学习小组",
      "分享推广佣金",
      "学习服务支持",
    ],
    bgColor: "bg-violet-500",
    borderColor: "#6D28D9",
    ctaText: "立即订阅",
  },
  {
    grade: "year",
    name: "年度会员",
    price: `¥${USER_GRADE_PRICES.year}`,
    period: "/年",
    features: [
      "全部课程内容",
      "全部游戏畅玩",
      "AI - 智能助力",
      "高级音效发音",
      "更多学习小组",
      "更多辅助功能",
      "分享推广佣金",
      "高级服务支持",
    ],
    bgColor: "bg-gradient-to-b from-[#0D7369] to-[#0A5C53]",
    ctaText: "立即订阅",
    highlight: true,
  },
  {
    grade: "lifetime",
    name: "终身会员",
    price: `¥${USER_GRADE_PRICES.lifetime}`,
    period: "",
    features: [
      "解锁所有权益",
      "功能永久有效",
      "永久没有续费",
      "全部课程内容",
      "全部游戏畅玩",
      "AI - 智能助力",
      "高级音效发音",
      "更多学习小组",
      "更多辅助功能",
      "更多推广佣金",
      "专属服务支持",
    ],
    bgColor: "bg-[#ca9302]",
    borderColor: "#D97706",
    ctaText: "立即订阅",
  },
];

export function PricingGrid() {
  const router = useRouter();
  const [loading, setLoading] = useState<string | null>(null);

  async function handleSubscribe(grade: string) {
    if (loading) return;
    setLoading(grade);
    try {
      const res = await orderApi.createMembershipOrder({
        grade,
        paymentMethod: PAYMENT_METHODS.WECHAT,
      });
      if (res.code === 0 && res.data?.id) {
        router.push(`/purchase/payment/${res.data.id}`);
      }
    } finally {
      setLoading(null);
    }
  }

  return (
    <div className="grid w-full grid-cols-1 gap-4 md:grid-cols-3 lg:grid-cols-5">
      {plans.map((p) => (
        <PricingCard
          key={p.name}
          name={p.name}
          price={p.price}
          period={p.period}
          features={p.features}
          bgColor={p.bgColor}
          borderColor={p.borderColor}
          ctaText={loading === p.grade ? "创建订单..." : p.ctaText}
          highlight={p.highlight}
          disabled={p.grade === "free" || loading !== null}
          onSubscribe={() => handleSubscribe(p.grade)}
        />
      ))}
    </div>
  );
}
