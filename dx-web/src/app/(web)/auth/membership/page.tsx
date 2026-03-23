import { CircleCheck } from "lucide-react";
import { PricingGrid } from "@/features/web/auth/components/pricing-grid";
import { TestimonialsGrid } from "@/features/web/auth/components/testimonials-grid";
import { FaqSection } from "@/features/web/auth/components/faq-section";

export default function MembershipPage() {
  return (
    <div className="flex w-full max-w-[1200px] flex-col items-center gap-5 px-4 py-8 lg:gap-6 lg:px-8 lg:py-10">
      {/* Title */}
      <div className="flex flex-col items-center gap-2">
        <h1 className="text-2xl font-bold text-slate-900 lg:text-[32px]">会员订阅套餐</h1>
        <p className="text-sm text-slate-500">
          选择适合您的会员方案，解锁更多学习功能
        </p>
      </div>

      {/* Current plan badge */}
      <div className="flex items-center gap-1.5 rounded-full border border-teal-600 bg-teal-50 px-3 py-1.5">
        <CircleCheck className="h-3.5 w-3.5 text-teal-600" />
        <span className="text-xs font-medium text-teal-600">当前方案: 免费版</span>
      </div>

      {/* Pricing grid */}
      <PricingGrid />

      {/* Testimonials section */}
      <div className="flex w-full flex-col items-center gap-8 pt-12">
        <div className="flex w-full items-center gap-4">
          <div className="h-px flex-1 bg-slate-300" />
          <h2 className="text-xl font-extrabold text-slate-900 lg:text-[28px]">会员真实体验</h2>
          <div className="h-px flex-1 bg-slate-300" />
        </div>
        <TestimonialsGrid />
      </div>

      {/* FAQ section */}
      <FaqSection />
    </div>
  );
}
