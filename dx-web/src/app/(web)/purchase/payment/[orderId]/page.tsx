import { OrderPayment } from "@/features/web/purchase/components/order-payment";

export default async function PaymentPage({
  params,
}: {
  params: Promise<{ orderId: string }>;
}) {
  const { orderId } = await params;

  return (
    <div className="flex w-full flex-1 items-center justify-center px-4 py-10">
      <OrderPayment orderId={orderId} />
    </div>
  );
}
