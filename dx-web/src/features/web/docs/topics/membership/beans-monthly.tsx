import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function BeansMonthly() {
  return (
    <>
      <DocSection id="grants" title="每月赠送多少">
        <DocKeyValue
          items={[
            {
              key: "月 / 季 / 年会员",
              value: "每月 10,000 豆",
            },
            {
              key: "终身会员",
              value: "每月 15,000 豆",
              note: "这是终身的独有优势",
            },
            {
              key: "免费用户 / 会员过期",
              value: "没有赠送",
            },
          ]}
        />
      </DocSection>

      <DocSection id="when" title="赠送日是哪一天">
        <p>
          每月的赠送日按你首次付费的月份对应日来算。比如你是 3 月 15 日首次购买会员，那么每个月的 15 日系统都会给你发放赠送豆——4 月 15 日、5 月 15 日，以此类推。
        </p>
      </DocSection>

      <DocSection id="month-end" title="月末日期的特殊处理">
        <DocCallout variant="info" title="29/30/31 号的处理">
          如果你的赠送日原本是 29、30 或 31 号，遇到更短的月份（比如 2 月），赠送会顺延到该月的最后一天。也就是说，你每个月都能收到豆，不会因为自然月长短不一而&ldquo;跳过&rdquo;一个月。
        </DocCallout>
      </DocSection>

      <DocSection id="reset" title="月度清零机制">
        <p>
          新一个月的赠送豆到账前，系统会先把你上个月剩余的&ldquo;赠送豆&rdquo;清零，然后再发放新一批。这是为了让每月的赠送都是&ldquo;用来当月用&rdquo;的——你如果本月没用完，下月会收到新的 10,000 或 15,000，不会累积。
        </p>
        <p>
          重要：清零只针对&ldquo;赠送豆&rdquo;。你自己购买的能量豆不清零，可以随时使用，直到用完为止。
        </p>
      </DocSection>

      <DocSection id="expired" title="会员过期后">
        <p>
          如果你的会员过期，系统不再给你发放每月赠送豆。已经发放过的赠送豆不会被立即收回，但遇到下个月的清零节点时会一并清除。购买的豆始终不受会员状态影响。
        </p>
      </DocSection>
    </>
  );
}
