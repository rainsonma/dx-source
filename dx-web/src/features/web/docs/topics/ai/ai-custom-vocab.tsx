import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocFlow } from "@/features/web/docs/primitives/doc-flow";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocLink } from "@/features/web/docs/primitives/doc-link";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function AiCustomVocab() {
  return (
    <>
      <DocSection id="what" title="词汇变体：不生成句子，只生成词">
        <p>
          AI 随心学的词汇变体和句子变体是并排的两种用法。句子版让 AI 生成完整的英语文本，词汇版让 AI 直接生成一组相关的单词或短语——适合快速建立一个&ldquo;主题词汇课&rdquo;，而不需要完整的叙事上下文。
        </p>
        <p>
          比如你想做一套&ldquo;厨房器具词汇&rdquo;，给 AI 几个关键词就能得到十几个相关词及其解释，直接用作词汇配对、词汇消消乐等游戏的内容。
        </p>
      </DocSection>

      <DocSection id="flow" title="同样的四步流程">
        <DocFlow
          nodes={[
            { label: "generate", desc: "生成原始词汇" },
            { label: "format", desc: "格式化内容" },
            { label: "break", desc: "拆分单元" },
            { label: "gen items", desc: "生成题目" },
          ]}
        />
        <p>
          词汇版本的流程和句子版完全一致，同样是四步；每一步同样消耗能量豆，失败自动退还。区别只在于生成内容的类型。详见{" "}
          <DocLink href="/wiki/ai/ai-custom-sentence">AI 随心学（句子）</DocLink>
          对流程的完整介绍。
        </p>
      </DocSection>

      <DocSection id="game-limits" title="不同游戏类型的数量上限">
        <p>
          词汇变体的一个重要约束：不同的游戏类型对词对数量有硬性上限。这是为了保证游戏节奏不会被破坏。
        </p>
        <DocKeyValue
          items={[
            {
              key: "vocab-match（配对）",
              value: "每关 5 对",
              note: "节奏舒缓",
            },
            {
              key: "vocab-elimination（消消乐）",
              value: "每关 8 对",
              note: "中等节奏",
            },
            {
              key: "vocab-battle（对轰）",
              value: "每关 20 对",
              note: "高强度刷",
            },
          ]}
        />
      </DocSection>

      <DocSection id="level-limits" title="每课的容量上限">
        <DocKeyValue
          items={[
            {
              key: "每个关卡的词汇单元上限",
              value: "20 个",
            },
            {
              key: "每个课程的关卡上限",
              value: "20 个",
            },
          ]}
        />
        <DocCallout variant="info" title="20 × 20 的容量">
          单个词汇课程的总容量上限约为 20 关 × 20 单元。这个数量对大多数主题而言都绰绰有余；如果你真的需要更多，可以拆成两个相关联的课程。
        </DocCallout>
      </DocSection>
    </>
  );
}
