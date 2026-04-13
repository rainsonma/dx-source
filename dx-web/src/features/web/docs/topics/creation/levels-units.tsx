import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function LevelsUnits() {
  return (
    <>
      <DocSection id="level" title="关卡：可玩的单位">
        <p>
          关卡是学习者在游戏里直接面对的单位——&ldquo;我要玩哪一关&rdquo;说的就是这个。一个关卡可以包含一个或多个单元，但学习者看到的是一个完整的游戏流程。
        </p>
        <p>
          作为创作者，你在编辑器里可以添加任意数量的关卡，并把它们按顺序排列。学习者在游戏详情页看到的就是这个顺序。
        </p>
      </DocSection>

      <DocSection id="metadata" title="单元：关卡内的内容块">
        <p>
          单元（metadata）是关卡内部的组织单位。通常一个单元围绕一个小主题——比如&ldquo;动物词汇&rdquo;&ldquo;家庭用品&rdquo;&ldquo;问路场景&rdquo;。学习者在关卡内按顺序遇到这些单元。
        </p>
      </DocSection>

      <DocSection id="manage" title="增加、修改、重排、删除">
        <p>
          编辑器里所有这些操作都支持——新增关卡、编辑单元名称、拖拽重新排列顺序、删除不要的部分。改动会立即保存，不需要手动&ldquo;提交&rdquo;。
        </p>
      </DocSection>

      <DocSection id="content-types" title="内容类型限制">
        <DocCallout variant="info" title="内容类型与难度绑定">
          可选的内容类型有 4 种：单词 (word) / 组合 (block) / 短语 (phrase) / 句子 (sentence)。但具体能选哪些取决于课程难度——初级允许全部 4 种，中级只允许后 3 种，高级只允许句子。这是为了让难度和内容粒度匹配。
        </DocCallout>
      </DocSection>
    </>
  );
}
