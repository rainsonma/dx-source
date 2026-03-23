interface WechatQrCardProps {
  label: string;
}

export function WechatQrCard({ label }: WechatQrCardProps) {
  return (
    <div className="flex w-[180px] flex-col items-center gap-3">
      <span className="text-center text-[13px] font-semibold text-slate-900">
        {label}
      </span>
      <div className="flex h-40 w-40 items-center justify-center rounded-xl border border-slate-200 bg-white">
        <span className="text-xs text-slate-400">微信二维码</span>
      </div>
      <span className="text-center text-xs text-slate-400">
        打开微信扫一扫
      </span>
    </div>
  );
}
