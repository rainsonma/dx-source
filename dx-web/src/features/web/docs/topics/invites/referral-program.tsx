import { DocFlow } from "@/features/web/docs/primitives/doc-flow";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function ReferralProgram() {
  return (
    <>
      <DocSection id="page" title="推广页面">
        <p>
          侧边栏的&ldquo;邀请推广&rdquo;入口是你的推广后台。打开后你会看到自己的专属邀请链接、二维码，以及这段时间内的关键统计数据——它是一站式的推广工具。
        </p>
      </DocSection>

      <DocSection id="stats" title="四项核心指标">
        <DocKeyValue
          items={[
            {
              key: "累计佣金",
              value: "到目前为止赚到的全部佣金总额",
            },
            {
              key: "本月新增",
              value: "本月新增邀请到的好友数",
              note: "不论是否已付费",
            },
            {
              key: "待激活",
              value: "已注册但还没付费的好友数",
              note: "他们可能是你的潜在收益",
            },
            {
              key: "成功转化率",
              value: "已付费好友 / 邀请总数",
              note: "反映你的推广效率",
            },
          ]}
        />
      </DocSection>

      <DocSection id="success" title="怎么算推广成功">
        <p>
          推广的判断标准很简单：好友通过你的邀请链接注册斗学账号，并完成首次付费（会员或能量豆任意一种），即视为推广成功。朋友只是注册不付费的话算&ldquo;待激活&rdquo;，不产生佣金。
        </p>
      </DocSection>

      <DocSection id="status-flow" title="推广状态的流转">
        <DocFlow
          nodes={[
            { label: "pending", desc: "待激活" },
            { label: "paid", desc: "已付费" },
            { label: "rewarded", desc: "佣金已发放" },
          ]}
        />
        <p>
          每一笔推广都会经历三个状态：pending 是朋友刚注册但未付费；paid 是朋友完成付费，佣金即将发放；rewarded 是佣金已经打到你的账户。从 pending 到 rewarded 通常需要几小时到一两天。
        </p>
      </DocSection>

      <DocSection id="settle" title="佣金结算">
        <p>
          佣金按时结算，可以随时在推广页申请提现——具体的提现方式和到账时间见推广页顶部的说明。佣金本身没有上限，邀请越多的朋友付费，赚到的佣金就越多。
        </p>
      </DocSection>
    </>
  );
}
