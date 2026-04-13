import { FileEdit, Scissors, Sparkles } from "lucide-react";
import { DocFeatureGrid } from "@/features/web/docs/primitives/doc-feature-grid";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function ContentItems() {
  return (
    <>
      <DocSection id="what" title="内容条目是最小学习单位">
        <p>
          内容条目（content item）是斗学里最小的学习单位——一条就是一道具体的题目：一个要背的单词、一个要翻译的句子、一个要听写的短语。学习者闯关时遇到的每一题都是一个内容条目。
        </p>
      </DocSection>

      <DocSection id="add-methods" title="三种添加方式">
        <DocFeatureGrid
          columns={3}
          items={[
            {
              icon: FileEdit,
              iconColor: "text-teal-600",
              title: "手动添加",
              desc: "逐条输入或从外部粘贴，适合内容量不大的精细课程。",
            },
            {
              icon: Sparkles,
              iconColor: "text-purple-600",
              title: "AI 生成",
              desc: "让 AI 按当前单元的主题批量生成，适合快速填充内容。",
            },
            {
              icon: Scissors,
              iconColor: "text-amber-600",
              title: "从单元分解",
              desc: "把一段长文本（比如一篇文章）自动拆成多条内容条目。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="reorder" title="拖拽批量重排">
        <p>
          添加进来的内容条目可以通过拖拽手柄调整顺序。学习者会按这个顺序遇到题目，所以合理的排序能让关卡的难度曲线更平滑。
        </p>
      </DocSection>

      <DocSection id="delete" title="删除">
        <p>
          支持单个删除和整关清空。整关清空是&ldquo;这一关的所有内容条目都删掉，但保留关卡本身&rdquo;——适合你想重做某一关但保留结构的场景。
        </p>
      </DocSection>
    </>
  );
}
