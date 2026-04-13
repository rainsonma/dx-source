import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocSteps } from "@/features/web/docs/primitives/doc-steps";

export default function RedeemCodes() {
  return (
    <>
      <DocSection id="what" title="兑换码是什么">
        <p>
          兑换码是一串可以直接换成权益的字符串——通常是一段字母和数字的组合。来源可能是斗学的活动奖品、官方发放的福利、或者客服补偿。兑换成功后对应的权益会直接进入你的账户。
        </p>
      </DocSection>

      <DocSection id="what-can-redeem" title="可以兑换什么">
        <p>
          兑换码目前支持两类权益：一段时间的会员（例如&ldquo;一个月会员&rdquo;），或者一定数量的能量豆（例如 10,000 豆）。具体能换什么由发放兑换码的一方决定，兑换时系统会告诉你。
        </p>
      </DocSection>

      <DocSection id="steps" title="使用兑换码">
        <DocSteps
          steps={[
            {
              title: "进入兑换码页面",
              desc: "在侧边栏找到&ldquo;兑换码&rdquo;入口",
            },
            {
              title: "粘贴或输入兑换码",
              desc: "注意区分大小写——兑换码通常是大小写敏感的",
            },
            {
              title: "点击兑换",
              desc: "系统会实时校验并发放权益，成功后立即生效",
            },
          ]}
        />
      </DocSection>

      <DocSection id="errors" title="兑换失败的常见原因">
        <DocCallout variant="warning" title="三种常见错误">
          &ldquo;已被使用&rdquo;——这串兑换码以前被人（包括你自己）用过了，每串码通常只能用一次。&ldquo;兑换码不存在&rdquo;——可能是输错了，检查一下大小写和字符；也可能是假的。&ldquo;已过期&rdquo;——兑换码有有效期，过了就不能用了。
        </DocCallout>
      </DocSection>
    </>
  );
}
