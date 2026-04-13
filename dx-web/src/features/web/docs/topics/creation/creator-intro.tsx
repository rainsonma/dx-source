import { BookOpen, Layers, List, Puzzle } from "lucide-react";
import { DocFeatureGrid } from "@/features/web/docs/primitives/doc-feature-grid";
import { DocFlow } from "@/features/web/docs/primitives/doc-flow";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function CreatorIntro() {
  return (
    <>
      <DocSection id="why" title="为什么要自己创作课程">
        <p>
          斗学的游戏库已经相当丰富，但总有一些内容只有你自己才知道想练什么——比如你的专业词汇、孩子的课本内容、或者你和朋友要一起刷的某个特定主题。这时候自己做一个课程就很合适。
        </p>
        <p>
          创作出来的课程可以自己玩、分享给朋友、或者公开发布到游戏广场让所有斗学用户都能看到。
        </p>
      </DocSection>

      <DocSection id="structure" title="课程的层级结构">
        <p>斗学课程是一个四层嵌套的结构，从大到小：</p>
        <DocFeatureGrid
          columns={2}
          items={[
            {
              icon: BookOpen,
              iconColor: "text-teal-600",
              title: "课程 (course)",
              desc: "一个完整的游戏课程包，是最外层容器",
            },
            {
              icon: Layers,
              iconColor: "text-blue-600",
              title: "关卡 (level)",
              desc: "课程里一个可以独立玩的单位",
            },
            {
              icon: List,
              iconColor: "text-amber-600",
              title: "单元 (metadata)",
              desc: "关卡内部的一个&ldquo;内容块&rdquo;，通常围绕一个主题",
            },
            {
              icon: Puzzle,
              iconColor: "text-purple-600",
              title: "内容条目 (content item)",
              desc: "最小的学习单位，就是一道具体的题目",
            },
          ]}
        />
      </DocSection>

      <DocSection id="paths" title="两种创建路径">
        <p>
          你可以从零开始手动写每一条内容，也可以让 AI 帮你生成。两种方式在同一个编辑器里都支持，可以混用——AI 生成基础内容，然后你手动修改细节，是最常见的工作流。
        </p>
      </DocSection>

      <DocSection id="lifecycle" title="生命周期">
        <DocFlow
          nodes={[
            { label: "draft", desc: "草稿，只有你能看到" },
            { label: "published", desc: "已发布，公开可玩" },
            { label: "withdraw", desc: "已撤回，下架但保留" },
          ]}
        />
        <p>
          新建的课程初始状态是 draft（草稿）。编辑完成后点发布进入 published 状态，其他用户就能在游戏广场找到并玩。如果想下架，可以撤回到 withdraw 状态——不会删除，之后还能重新发布。
        </p>
      </DocSection>
    </>
  );
}
