import { notFound } from "next/navigation";

type PayMethod = "wechat" | "alipay";

interface PayConfig {
  name: string;
  badgeColor: string;
  badgeIcon: string;
  scanText: string;
}

const payData: Record<PayMethod, PayConfig> = {
  wechat: {
    name: "微信支付",
    badgeColor: "bg-[#07C160]",
    badgeIcon: "W",
    scanText: "请使用微信扫一扫完成支付",
  },
  alipay: {
    name: "支付宝",
    badgeColor: "bg-[#1677FF]",
    badgeIcon: "A",
    scanText: "请使用支付宝扫一扫完成支付",
  },
};

const validMethods: PayMethod[] = ["wechat", "alipay"];

export function generateStaticParams() {
  return validMethods.map((method) => ({ method }));
}

export default async function PaymentPage({
  params,
}: {
  params: Promise<{ method: string }>;
}) {
  const { method } = await params;

  if (!validMethods.includes(method as PayMethod)) {
    notFound();
  }

  const config = payData[method as PayMethod];

  return (
    <div className="flex min-h-screen w-full items-center justify-center px-4 bg-gradient-to-b from-teal-100 via-blue-100 via-40% via-purple-100 via-65% to-white to-100%">
      {/* Payment card */}
      <div className="flex w-full max-w-[520px] flex-col overflow-hidden rounded-2xl bg-white shadow-[0_8px_32px_rgba(15,23,42,0.1)]">
        {/* Card top */}
        <div className="flex flex-col gap-3.5 px-7 py-6">
          {/* Method row */}
          <div className="flex items-center gap-2.5">
            <div
              className={`flex h-7 w-7 items-center justify-center rounded-[7px] ${config.badgeColor}`}
            >
              <span className="text-xs font-bold text-white">
                {config.badgeIcon}
              </span>
            </div>
            <span className="text-base font-bold text-slate-900">
              {config.name}
            </span>
          </div>

          {/* Order row */}
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400">订单编号</span>
            <span className="text-xs text-slate-500">
              NO.202403211523090001
            </span>
          </div>

          {/* Amount */}
          <div className="flex items-center gap-0.5">
            <span className="text-4xl font-bold text-teal-600">¥</span>
            <span className="text-4xl font-extrabold text-teal-600">309</span>
            <span className="text-4xl font-bold text-teal-600">.00</span>
          </div>
        </div>

        {/* Divider */}
        <div className="h-px w-full bg-slate-200" />

        {/* QR section */}
        <div className="flex flex-col items-center gap-3 px-7 py-6">
          {/* QR code placeholder */}
          <div className="flex h-[180px] w-[180px] items-center justify-center rounded-lg border border-slate-200 bg-slate-50">
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
          <span className="text-[13px] text-slate-600">{config.scanText}</span>
        </div>
      </div>
    </div>
  );
}
