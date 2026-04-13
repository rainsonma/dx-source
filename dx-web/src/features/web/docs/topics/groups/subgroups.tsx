import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function Subgroups() {
  return (
    <>
      <DocSection id="what" title="子分组是什么">
        <p>
          子分组是小组内部的&ldquo;队伍&rdquo;，专门为 group_team（分组对战）模式设计的。把小组成员分到不同的子分组后，再以 team 模式开局游戏，分组之间就可以互相对抗，看哪一队总分最高。
        </p>
      </DocSection>

      <DocSection id="rules" title="子分组规则">
        <DocKeyValue
          items={[
            {
              key: "每个小组的子分组上限",
              value: "10 个",
            },
            {
              key: "成员归属",
              value: "一个成员最多属于一个子分组",
              note: "不能跨队",
            },
            {
              key: "可以有未分组的成员",
              value: "是",
              note: "但他们不能参与 team 模式对战",
            },
          ]}
        />
      </DocSection>

      <DocSection id="assign" title="如何分配成员到子分组">
        <p>
          组主在小组管理页可以创建新的子分组，起名字，然后把成员逐个分配进去。分配过程是组主手动完成的——斗学不会自动帮你分队，因为通常组主对成员更熟悉，分队结果更合理。
        </p>
      </DocSection>

      <DocSection id="when" title="什么时候用得上">
        <DocCallout variant="info" title="仅 team 模式需要">
          如果你的小组只玩 group_solo（组内个人排名）模式，子分组功能完全可以不用。子分组只在 group_team 模式下才会影响结算。
        </DocCallout>
      </DocSection>
    </>
  );
}
