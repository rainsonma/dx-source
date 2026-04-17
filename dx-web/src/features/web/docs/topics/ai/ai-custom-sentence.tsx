import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocFlow } from "@/features/web/docs/primitives/doc-flow";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocLink } from "@/features/web/docs/primitives/doc-link";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function AiCustomSentence() {
  return (
    <>
      <DocSection id="what" title="AI 随心学（句子）是什么">
        <p>
          AI 随心学是斗学提供的&ldquo;让 AI 帮你生成课程&rdquo;能力。你只需要给出几个关键词和目标难度，AI 就会生成一段完整的英语内容（一个故事、一段对话、一组相关句子），然后自动拆成可以闯关的题目。
        </p>
        <p>
          适合的场景：你想练某个特定主题（比如&ldquo;机场登机&rdquo;&ldquo;商务邮件&rdquo;），但现有的游戏里没有刚好合适的内容；或者你想给自己的学生/孩子定制一套课程。
        </p>
      </DocSection>

      <DocSection id="inputs" title="需要你提供什么">
        <DocKeyValue
          items={[
            {
              key: "关键词",
              value: "3–10 个英文或中文关键词",
              note: "决定内容主题",
            },
            {
              key: "难度 (CEFR)",
              value: "a1-a2 / b1-b2 / c1-c2",
              note: "初级 / 中级 / 高级",
            },
          ]}
        />
      </DocSection>

      <DocSection id="flow" title="四步生成流程">
        <DocFlow
          nodes={[
            { label: "generate", desc: "生成原始故事" },
            { label: "format", desc: "格式化句子" },
            { label: "break", desc: "拆分单元" },
            { label: "gen items", desc: "生成题目" },
          ]}
        />
        <p>
          整个过程是分四步进行的：先生成一段完整内容，然后逐句格式化，再按句子拆成学习单元，最后为每个单元生成可闯关的题目。每一步都可以独立查看结果，不满意可以重新生成那一步。
        </p>
      </DocSection>

      <DocSection id="cost" title="每一步都消耗能量豆">
        <p>
          AI 随心学的每一步都要消耗能量豆。能量豆是斗学的软通货，每月会员自动赠送一部分，也可以单独购买。如果某一步 AI 生成失败（网络、模型异常），消耗的豆会自动退还，不会白白扣掉。详情见{" "}
          <DocLink href="/wiki/membership/beans-monthly">月度赠送与清零</DocLink>
          。
        </p>
        <DocCallout variant="warning" title="仅 VIP 可用">
          AI 随心学功能仅对 VIP 会员开放。免费用户在尝试进入时会看到升级提示。
        </DocCallout>
      </DocSection>

      <DocSection id="after" title="生成完成后">
        <p>
          生成完的课程会自动保存到你的&ldquo;我创建的游戏&rdquo;列表里，状态是草稿。你可以继续在编辑器里调整内容、添加更多单元，或者直接发布出去——发布后这个游戏会进入游戏广场，其他用户也能玩。
        </p>
      </DocSection>
    </>
  );
}
