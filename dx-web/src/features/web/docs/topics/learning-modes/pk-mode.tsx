import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocCompareTable } from "@/features/web/docs/primitives/doc-compare-table";
import { DocFlow } from "@/features/web/docs/primitives/doc-flow";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocSteps } from "@/features/web/docs/primitives/doc-steps";

export default function PkMode() {
  return (
    <>
      <DocSection id="what" title="什么是 PK 模式">
        <p>
          PK 是斗学的实时 1v1 对战模式——你和对手同时进入同一关卡，同一批题目，谁先完成并取得更高分数谁赢。整个过程有即时的对手进度显示，节奏紧张。
        </p>
        <p>
          PK 模式适合你已经熟悉基础玩法、想来点刺激的时候；或者和朋友约战，互相比拼。
        </p>
        <DocCallout variant="warning" title="仅 VIP 可用">
          PK 的所有形式都需要 VIP 会员。如果你还没开通，会在进入时被提示升级。
        </DocCallout>
      </DocSection>

      <DocSection id="two-types" title="两种对战方式">
        <DocCompareTable
          columns={["random（随机匹配）", "specified（指定对手）"]}
          labelHeader="维度"
          rows={[
            {
              label: "匹配方式",
              values: ["系统自动为你配对", "自己选择一个人"],
            },
            {
              label: "等待时间",
              values: ["即开即战", "需要对方同意"],
            },
            {
              label: "对方资格",
              values: ["——", "对方必须是 VIP 且当前在线"],
            },
            {
              label: "典型场景",
              values: ["想立刻开始", "和朋友约战"],
            },
          ]}
        />
      </DocSection>

      <DocSection id="difficulty" title="随机匹配的难度选择">
        <p>
          选择&ldquo;随机匹配&rdquo;时，你可以挑一个想要的难度档位——从轻松到紧张三档：
        </p>
        <DocKeyValue
          items={[
            {
              key: "easy（简单）",
              value: "轻松热身",
              note: "适合放松节奏、刚开始练",
            },
            {
              key: "normal（普通）",
              value: "正常强度",
              note: "大多数人的首选",
            },
            {
              key: "hard（困难）",
              value: "高强度对抗",
              note: "适合想挑战极限的你",
            },
          ]}
        />
      </DocSection>

      <DocSection id="invite-flow" title="指定对手的邀请流程">
        <p>
          选择&ldquo;指定对手&rdquo;时，你需要主动发起邀请：
        </p>
        <DocSteps
          steps={[
            {
              title: "选游戏和参数",
              desc: "先把游戏、关卡、难度、学习模式定好，和单人模式流程一致",
            },
            {
              title: "选择对手",
              desc: "从在线的好友或列表中挑一个人，该人必须也是 VIP",
            },
            {
              title: "发起邀请",
              desc: "系统会实时推送通知给对方",
            },
            {
              title: "等待回应",
              desc: "对方选择接受或拒绝——接受后双方自动进入同一个对战房间",
            },
          ]}
        />
      </DocSection>

      <DocSection id="invite-lifecycle" title="邀请的状态流转">
        <DocFlow
          nodes={[
            { label: "pending", desc: "等待回应" },
            { label: "accepted", desc: "对战中" },
            { label: "declined", desc: "已拒绝" },
          ]}
        />
        <p>
          如果对方长时间没有任何回应，邀请会进入过期状态，你可以重新发起。
        </p>
      </DocSection>

      <DocSection id="results" title="PK 结束后">
        <p>
          率先完成关卡的一方获胜，双方都会看到最终的分数对比、排名和本局获得的经验值。无论输赢，完成的一方都能获得对应的经验奖励——PK 模式鼓励参与，而不是只奖励胜利。
        </p>
      </DocSection>
    </>
  );
}
