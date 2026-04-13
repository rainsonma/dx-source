import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocFlow } from "@/features/web/docs/primitives/doc-flow";
import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocSteps } from "@/features/web/docs/primitives/doc-steps";

export default function PurchaseFlow() {
  return (
    <>
      <DocSection id="steps" title="五步完成购买">
        <DocSteps
          steps={[
            {
              title: "选会员档位",
              desc: "在购买页选择月度、季度、年度或终身其中之一",
            },
            {
              title: "生成订单",
              desc: "确认档位后系统创建一笔待支付订单",
            },
            {
              title: "选支付方式",
              desc: "支持微信支付或支付宝，任选其一",
            },
            {
              title: "扫码完成支付",
              desc: "按屏幕提示扫描二维码完成实际付款",
            },
            {
              title: "权益到账",
              desc: "支付成功后后端自动把会员权益写入你的账户，通常 1 分钟内生效",
            },
          ]}
        />
      </DocSection>

      <DocSection id="lifecycle" title="订单的状态流转">
        <DocFlow
          nodes={[
            { label: "pending", desc: "待支付" },
            { label: "paid", desc: "已支付" },
            { label: "fulfilled", desc: "权益已发放" },
          ]}
        />
        <p>
          每笔订单会经历这三个状态。&ldquo;pending&rdquo;是你刚生成订单还没支付的时候；&ldquo;paid&rdquo;是支付平台已收到钱但后端还没来得及发放权益的短暂状态；&ldquo;fulfilled&rdquo;是会员已经到你的账户里了。
        </p>
      </DocSection>

      <DocSection id="expire" title="订单 30 分钟过期">
        <DocCallout variant="warning" title="30 分钟未支付会失效">
          订单创建后如果 30 分钟内没有完成支付，会自动进入过期状态，无法继续支付。过期后你需要回到购买页重新生成一笔新订单。这样设计是为了避免大量长期悬挂的未支付订单。
        </DocCallout>
      </DocSection>
    </>
  );
}
