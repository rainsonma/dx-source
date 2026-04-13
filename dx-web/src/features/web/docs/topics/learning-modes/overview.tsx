import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocCompareTable } from "@/features/web/docs/primitives/doc-compare-table";
import { DocLink } from "@/features/web/docs/primitives/doc-link";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function LearningModesOverview() {
  return (
    <>
      <DocSection id="three-modes" title="多种模式，一套知识">
        <p>
          斗学把同一套游戏内容设计成了多种玩法：单人闯关、PK 对战、小组共学。它们不是彼此替代的不同&ldquo;游戏&rdquo;，而是同一个游戏的三种节奏——你可以在心情不同的时候自由切换。
        </p>
        <p>
          单人是默认体验，谁都能玩；PK 把两个人放到同一关卡上抢先；小组把一群人组织起来一起完成或对抗。选择哪一种取决于你现在想要的感觉，而不是学习内容。
        </p>
      </DocSection>

      <DocSection id="compare" title="一表看懂">
        <DocCompareTable
          columns={["单人闯关", "PK 对战", "小组共学"]}
          labelHeader="维度"
          rows={[
            {
              label: "参与人数",
              values: ["1 人", "2 人", "多人"],
            },
            {
              label: "对手",
              values: ["无（自己）", "随机 或 指定", "小组成员"],
            },
            {
              label: "VIP 要求",
              values: [false, true, true],
            },
            {
              label: "是否实时",
              values: [false, true, true],
            },
            {
              label: "典型场景",
              values: ["自由练习", "紧张刺激", "组队学习"],
            },
          ]}
        />
      </DocSection>

      <DocSection id="shared-basics" title="多种模式的共同基础">
        <p>
          不管你选哪一种模式，底层都共享同一套体系：四种游戏类型、三个难度（初/中/高）、四种学习模式（听说读写）。也就是说你的游戏选择一旦确定，切换学习模式不会让你重新学一套新机制——详见{" "}
          <DocLink href="/docs/learning-modes/game-types">游戏类型与技能矩阵</DocLink>
          。
        </p>
      </DocSection>

      <DocSection id="how-to-choose" title="我该选哪个">
        <DocCallout variant="tip" title="三个简单判断">
          想自己安静练习？选单人。想来点紧张刺激的对抗？选 PK——可以随机匹配立刻开始，也可以指定一个朋友约战。
          想和朋友一起开局、看到彼此的进度？选小组。
        </DocCallout>
      </DocSection>
    </>
  );
}
