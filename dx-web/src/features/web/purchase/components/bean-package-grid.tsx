"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { BeanPackageCard } from "@/features/web/purchase/components/bean-package-card";
import { BEAN_PACKAGES } from "@/consts/bean-package";
import { PAYMENT_METHODS } from "@/consts/payment-method";
import { orderApi } from "@/lib/api-client";

export function BeanPackageGrid() {
  const router = useRouter();
  const [loading, setLoading] = useState<string | null>(null);

  async function handlePurchase(slug: string) {
    if (loading) return;
    setLoading(slug);
    try {
      const res = await orderApi.createBeansOrder({
        package: slug,
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
    <div className="flex w-full gap-4 overflow-x-auto pb-2 lg:overflow-visible">
      {BEAN_PACKAGES.map((pkg) => (
        <BeanPackageCard
          key={pkg.slug}
          beans={pkg.beans}
          bonus={pkg.bonus}
          price={pkg.price}
          tag={pkg.tag}
          highlight={pkg.slug === "beans-10"}
          disabled={loading !== null}
          onPurchase={() => handlePurchase(pkg.slug)}
        />
      ))}
    </div>
  );
}
