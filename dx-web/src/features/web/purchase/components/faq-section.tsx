import { MessageCircle, ScanLine } from "lucide-react";

export function FaqSection() {
  return (
    <div className="flex w-full flex-col gap-6 py-12">
      {/* Title row */}
      <div className="flex w-full items-center gap-4">
        <div className="h-px flex-1 bg-slate-300" />
        <h2 className="text-xl font-extrabold text-slate-900 lg:text-[28px]">
          我有问题想咨询
        </h2>
        <div className="h-px flex-1 bg-slate-300" />
      </div>

      {/* Contact button + QR popup */}
      <div className="flex w-full flex-col items-center gap-6">
        <button className="flex items-center gap-2.5 rounded-[14px] bg-teal-600 px-8 py-3.5 shadow-[0_8px_24px_rgba(13,148,136,0.19)] hover:bg-teal-700 lg:px-[60px] lg:py-[18px]">
          <MessageCircle className="h-[22px] w-[22px] text-white" />
          <span className="text-base font-bold text-white lg:text-xl">联系斗学小助手</span>
        </button>

        {/* QR popup card */}
        <div className="flex flex-col items-center gap-4 rounded-2xl border border-slate-200 bg-white p-6 shadow-[0_8px_24px_rgba(15,23,42,0.09)]">
          <div className="flex items-center gap-2">
            <ScanLine className="h-4 w-4 text-teal-600" />
            <span className="text-sm font-semibold text-slate-900">
              微信扫码添加
            </span>
          </div>
          <div className="flex h-[180px] w-[180px] items-center justify-center rounded-xl border border-slate-200 bg-slate-50">
            <div className="flex flex-col items-center gap-2">
              <div className="grid grid-cols-3 gap-1">
                {Array.from({ length: 9 }).map((_, i) => (
                  <div
                    key={i}
                    className={`h-4 w-4 rounded-sm ${i % 3 === 0 ? "bg-slate-800" : i % 2 === 0 ? "bg-slate-600" : "bg-slate-400"}`}
                  />
                ))}
              </div>
              <span className="text-[10px] text-slate-400">QR Code</span>
            </div>
          </div>
          <div className="flex items-center gap-1">
            <MessageCircle className="h-3.5 w-3.5 text-slate-400" />
            <span className="text-xs text-slate-400">
              斗学官方客服
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
