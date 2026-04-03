import { Lightbulb } from "lucide-react";
import { BeanPackageGrid } from "@/features/web/purchase/components/bean-package-grid";

export default function BeansPurchasePage() {
  return (
    <div className="flex w-full max-w-[1200px] flex-col items-center gap-6 px-4 py-8 lg:px-8 lg:py-10">
      <div className="flex flex-col items-center gap-2">
        <h1 className="text-2xl font-bold text-slate-900 lg:text-[32px]">
          能量豆充值
        </h1>
        <p className="text-sm text-slate-500">选择适合您的能量豆套餐</p>
      </div>

      <BeanPackageGrid />

      <div className="mt-6 w-full rounded-2xl border border-slate-200 bg-slate-50 p-6">
        <div className="mb-3 flex items-center gap-2">
          <Lightbulb className="h-5 w-5 text-amber-500" />
          <span className="text-base font-bold text-slate-900">能量豆用途</span>
        </div>
        <ul className="flex flex-col gap-1.5 text-sm text-slate-600">
          <li>&bull; 编辑端 AI 功能（课程生成、句子拆分、内容加工等）</li>
          <li>&bull; 高级学习辅助功能</li>
          <li>&bull; 更多即将推出的增值服务</li>
        </ul>
      </div>
    </div>
  );
}
