import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function NewCourse() {
  return (
    <>
      <DocSection id="entry" title="在哪里创建">
        <p>
          进入&ldquo;我的游戏&rdquo;页面，右上角有&ldquo;新建&rdquo;按钮。点击后弹出新建课程窗口，填完基本信息即可进入编辑器。
        </p>
      </DocSection>

      <DocSection id="fields" title="基本信息字段">
        <DocKeyValue
          items={[
            { key: "名称", value: "必填，会显示在游戏卡片上" },
            { key: "描述", value: "可选，一两句话说明课程主题" },
            { key: "封面图", value: "可选，用作卡片和详情页的主视觉" },
            {
              key: "分类",
              value: "从分类树选择一个或多个",
              note: "影响广场的筛选",
            },
            {
              key: "出版社",
              value: "可选",
              note: "如果是针对特定教材",
            },
            {
              key: "模式",
              value: "四选一",
              note: "一个课程只能对应一种游戏玩法",
            },
          ]}
        />
      </DocSection>

      <DocSection id="cover" title="封面上传规则">
        <DocCallout variant="warning" title="封面限制">
          封面图最大 2MB，格式仅支持 JPEG 和 PNG。超出限制会被系统拒绝，需要你先压缩或转格式再上传。
        </DocCallout>
      </DocSection>

      <DocSection id="next" title="创建完成后">
        <p>
          填完基本信息后课程立即创建成功，进入草稿状态。页面会跳转到课程编辑器，你可以开始添加关卡、单元和具体内容。
        </p>
      </DocSection>
    </>
  );
}
