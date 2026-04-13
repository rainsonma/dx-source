import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocFlow } from "@/features/web/docs/primitives/doc-flow";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function PublishWithdraw() {
  return (
    <>
      <DocSection id="checklist" title="发布前请检查">
        <p>
          发布一个课程前，确保基本结构是完整的：至少有一个关卡、每个关卡至少有一个单元、每个单元至少有一条内容条目。否则学习者点进来会看到空关卡，体验不好。
        </p>
      </DocSection>

      <DocSection id="publish" title="发布到游戏广场">
        <p>
          在编辑器顶部点&ldquo;发布&rdquo;按钮，课程从 draft 变为 published。几秒钟内它会出现在游戏广场，按分类、出版社、游戏模式被其他用户搜到和筛出。你的课程从此向所有斗学用户开放。
        </p>
      </DocSection>

      <DocSection id="withdraw" title="撤回：下架不是删除">
        <p>
          如果想把一个已发布的课程下架，点&ldquo;撤回&rdquo;即可。撤回后课程状态变为 withdraw，游戏广场不再展示，但课程内容不会丢失，你的&ldquo;我的游戏&rdquo;里还能看到。什么时候想重新发布，再点一次&ldquo;发布&rdquo;就能回到 published。
        </p>
        <DocCallout variant="info" title="撤回的用途">
          常见的用途是发现内容有错想修改，或者觉得还没打磨完不想让别人看到。撤回 → 修改 → 重新发布是一个完整的迭代流程。
        </DocCallout>
      </DocSection>

      <DocSection id="lifecycle" title="生命周期总结">
        <DocFlow
          nodes={[
            { label: "draft", desc: "新建或撤回后的编辑状态" },
            { label: "published", desc: "公开发布，所有人可玩" },
            { label: "withdraw", desc: "下架但保留数据" },
          ]}
        />
      </DocSection>
    </>
  );
}
