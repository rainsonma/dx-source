import { Check, X } from "lucide-react";

export default function PaymentConfirmPage() {
  return (
    <div className="flex min-h-screen w-full items-center justify-center px-4 bg-black/40">
      <div className="flex w-full max-w-[480px] flex-col overflow-hidden rounded-2xl bg-white shadow-[0_8px_32px_rgba(15,23,42,0.1)]">
        {/* Header */}
        <div className="flex flex-col gap-2 px-8 pb-6 pt-7">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-bold text-slate-900">
              确认协议并购买
            </h3>
            <button
              type="button"
              aria-label="关闭"
              className="flex h-7 w-7 items-center justify-center rounded-md bg-slate-100"
            >
              <X className="h-4 w-4 text-slate-500" />
            </button>
          </div>
          <p className="text-[13px] text-slate-500">
            为了更好的保障您的合法权益，请您仔细阅读并同意以下协议
          </p>
        </div>

        <div className="h-px w-full bg-slate-200" />

        {/* Checkbox section */}
        <div className="px-8 py-5">
          <div className="flex gap-2.5">
            <div className="flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded border-[1.5px] border-slate-300 bg-white" />
            <div className="flex flex-col gap-1">
              <span className="text-sm text-slate-700">
                我已阅读并同意以下协议
              </span>
              <span className="text-xs text-teal-600">
                《斗学会员服务协议》《自动续费协议》
              </span>
            </div>
          </div>
        </div>

        <div className="h-px w-full bg-slate-200" />

        {/* Payment method selection */}
        <div className="flex flex-col gap-6 px-8 py-4">
          {/* WeChat - selected */}
          <div className="flex items-center gap-2.5">
            <div className="flex h-[18px] w-[18px] items-center justify-center rounded-full bg-teal-600">
              <Check className="h-2.5 w-2.5 text-white" />
            </div>
            <div className="flex h-[22px] w-[22px] items-center justify-center rounded-[5px] bg-[#07C160]">
              <span className="text-[9px] font-bold text-white">W</span>
            </div>
            <span className="text-sm font-medium text-slate-900">
              微信支付
            </span>
          </div>
          {/* Alipay - unselected */}
          <div className="flex items-center gap-2.5">
            <div className="h-[18px] w-[18px] rounded-full border-[1.5px] border-slate-300 bg-white" />
            <div className="flex h-[22px] w-[22px] items-center justify-center rounded-[5px] bg-[#1677FF]">
              <span className="text-[9px] font-bold text-white">A</span>
            </div>
            <span className="text-sm text-slate-600">支付宝</span>
          </div>
        </div>

        {/* Footer buttons */}
        <div className="flex items-center justify-end gap-3 px-8 pb-5 pt-5">
          <button
            type="button"
            className="rounded-lg border border-slate-300 bg-slate-50 px-6 py-2.5"
          >
            <span className="text-sm font-medium text-slate-600">取消</span>
          </button>
          <button
            type="button"
            className="rounded-lg bg-teal-600 px-6 py-2.5"
          >
            <span className="text-sm font-semibold text-white">立即购买</span>
          </button>
        </div>
      </div>
    </div>
  );
}
