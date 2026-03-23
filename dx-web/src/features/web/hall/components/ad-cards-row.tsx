import Link from "next/link";
import { Crown, ArrowRight, Gift } from "lucide-react";

export function AdCardsRow() {
  return (
    <div className="flex w-full flex-col gap-4 lg:flex-row">
      {/* Upgrade Pro card */}
      <div className="flex w-full items-center justify-between overflow-hidden rounded-[14px] bg-gradient-to-b from-teal-600 to-teal-700 px-7 py-5">
        <div className="flex h-full flex-col justify-center gap-2">
          <span className="w-fit rounded-full bg-white/20 px-3 py-1 text-[11px] font-semibold text-white">
            限时优惠
          </span>
          <h3 className="text-[22px] font-extrabold text-white">
            升级 Pro 会员
          </h3>
          <p className="text-[13px] text-white/80">
            解锁无限关卡，无限内容、AI 定制和专属练习
          </p>
        </div>
        <Link
          href="/auth/membership"
          className="flex shrink-0 items-center gap-1.5 rounded-lg bg-white px-5 py-2 text-[13px] font-semibold text-teal-700 hover:bg-white/90"
        >
          <Crown className="h-3.5 w-3.5" />
          立即升级
        </Link>
      </div>

      {/* Invite friends card */}
      <div className="flex w-full items-center justify-between overflow-hidden rounded-[14px] bg-gradient-to-b from-violet-600 to-violet-700 px-7 py-5">
        <div className="flex h-full flex-col justify-center gap-2">
          <span className="w-fit rounded-full bg-white/20 px-3 py-1 text-[11px] font-semibold text-white">
            限时活动
          </span>
          <h3 className="text-[22px] font-extrabold text-white">
            邀请好友一起学
          </h3>
          <p className="text-[13px] text-white/80">
            每邀请一位好友，都有收益，惊喜连连
          </p>
        </div>
        <Link
          href="/hall/invite"
          className="flex shrink-0 items-center gap-1.5 rounded-lg bg-white px-5 py-2 text-[13px] font-semibold text-violet-700 hover:bg-white/90"
        >
          <Gift className="h-3.5 w-3.5" />
          邀请好友
          <ArrowRight className="h-3.5 w-3.5" />
        </Link>
      </div>
    </div>
  );
}
