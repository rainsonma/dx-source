import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocLink } from "@/features/web/docs/primitives/doc-link";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function Unknown() {
  return (
    <>
      <DocSection id="add" title="如何加入生词">
        <p>
          在闯关过程中遇到不认识的单词、短语或句子，点击&ldquo;加入生词本&rdquo;按钮即可把它收入生词本。这个动作本身不影响游戏进度，只是把内容记下来供之后复习。
        </p>
      </DocSection>

      <DocSection id="stats" title="生词本页面能看到什么">
        <p>侧边栏的&ldquo;生词&rdquo;入口展示了你的生词总览：</p>
        <DocKeyValue
          items={[
            {
              key: "生词总量",
              value: "到目前为止收藏的全部生词数量",
            },
            {
              key: "今日新增",
              value: "今天新加入的生词数量",
            },
            {
              key: "近 3 天",
              value: "最近三天累积的新生词数量",
            },
          ]}
        />
      </DocSection>

      <DocSection id="manage" title="单个和批量删除">
        <p>
          列表中的每一条都可以单独删除；也可以勾选多条进行批量删除。删除后这条内容不再出现在生词本里，下次再遇到时可以重新加入。
        </p>
      </DocSection>

      <DocSection id="next" title="接下来呢">
        <p>
          加入生词本只是第一步。要让词真正留下来，需要把它转入{" "}
          <DocLink href="/wiki/vocabulary/review">复习本</DocLink>
          {" "}
          ——那里按间隔重复的节奏自动安排复习时机，比你自己&ldquo;想起来再翻&rdquo;要有效得多。
        </p>
      </DocSection>
    </>
  );
}
