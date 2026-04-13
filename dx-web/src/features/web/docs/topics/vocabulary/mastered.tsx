import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function Mastered() {
  return (
    <>
      <DocSection id="how" title="如何标为掌握">
        <p>
          两种方式：一是在复习本里连续正确复习几次，系统自动把它转为已掌握；二是在复习本或生词本中手动标记为&ldquo;已掌握&rdquo;——如果你已经非常确定这条词不需要再复习，可以直接跳过后面的间隔。
        </p>
      </DocSection>

      <DocSection id="page" title="已掌握页面">
        <p>侧边栏的&ldquo;已掌握&rdquo;入口展示你的掌握进度：</p>
        <DocKeyValue
          items={[
            {
              key: "累计掌握",
              value: "到目前为止掌握的总词数",
            },
            {
              key: "本周新增",
              value: "本周新进入已掌握的词数",
            },
            {
              key: "本月新增",
              value: "本月新进入已掌握的词数",
            },
          ]}
        />
      </DocSection>

      <DocSection id="removed-from-review" title="不再出现在复习中">
        <p>
          一旦标为已掌握，这条词就会从复习本的待复习列表里移出，不再打扰你。但它不会被物理删除——在已掌握页面还能查到，方便你复盘自己学过的内容。
        </p>
      </DocSection>

      <DocSection id="undo" title="觉得没记牢可以撤回">
        <p>
          如果某条词本来标记为已掌握，但你发现其实记不太清了，可以在已掌握页面撤回这个标记。撤回后这条词会重新进入复习本，按当前进度继续走间隔复习流程。
        </p>
      </DocSection>
    </>
  );
}
