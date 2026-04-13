import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocLink } from "@/features/web/docs/primitives/doc-link";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function GroupsOverview() {
  return (
    <>
      <DocSection id="what" title="学习小组的定位">
        <p>
          学习小组是斗学里的&ldquo;多人空间&rdquo;：一群有共同学习目标的人组合在一起，可以一起开局游戏、互相监督打卡、在小组内讨论。相比 PK 的临时对战，小组是更稳定、更长期的学习关系。
        </p>
        <p>
          小组的用途有很多：朋友间的约定学习、同学之间的刷题小分队、或者家庭成员一起。斗学不限制小组的主题，你可以按任何方式组织。
        </p>
      </DocSection>

      <DocSection id="limits" title="小组的硬限制">
        <DocKeyValue
          items={[
            {
              key: "单个小组成员上限",
              value: "50 人",
              note: "包括组主",
            },
            {
              key: "单个小组子分组上限",
              value: "10 个",
            },
            {
              key: "创建小组",
              value: "需要 VIP 会员",
              note: "加入他人的小组不需要",
            },
          ]}
        />
      </DocSection>

      <DocSection id="two-formats" title="两种玩法">
        <p>
          小组可以以两种方式玩游戏：group_solo（组内所有人各自闯同一游戏，按个人排名结算）或 group_team（成员分成若干子分组，分队对战）。具体差异见{" "}
          <DocLink href="/docs/learning-modes/group-mode">小组共学模式</DocLink>
          。
        </p>
      </DocSection>

      <DocSection id="vip" title="创建需要 VIP">
        <DocCallout variant="warning" title="创建小组需要 VIP">
          只有 VIP 会员可以创建新小组。免费用户可以通过邀请加入别人的小组，作为成员参与学习不受限制。
        </DocCallout>
      </DocSection>
    </>
  );
}
