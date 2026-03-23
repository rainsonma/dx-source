"use client";

import { RedeemInputCard } from "@/features/web/redeem/components/redeem-input-card";
import { RedeemHistoryTable } from "@/features/web/redeem/components/redeem-history-table";
import { RedeemAdminSection } from "@/features/web/redeem/components/redeem-admin-section";

type RedeemContentProps = {
  username: string | null;
};

/** Main redeem page content orchestrating all sections */
export function RedeemContent({ username }: RedeemContentProps) {
  const isAdmin = username === "rainson";

  return (
    <>
      {isAdmin && <RedeemAdminSection />}
      <RedeemInputCard />
      <RedeemHistoryTable />
    </>
  );
}
