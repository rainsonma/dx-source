import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocCompareTable } from "@/features/web/docs/primitives/doc-compare-table";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function BeansPackages() {
  return (
    <>
      <DocSection id="what-for" title="能量豆是什么">
        <p>
          能量豆是斗学的&ldquo;软通货&rdquo;——和人民币不同，它只能在斗学内部使用，主要用于消耗 AI 随心学的各种生成动作（生成故事、格式化、拆分、生成题目）。每次 AI 生成都会扣一定数量的豆；失败时自动退还，不会白扣。
        </p>
      </DocSection>

      <DocSection id="packages" title="五档充值包">
        <p>
          除了每月会员赠送的豆，你也可以直接花人民币买豆。斗学提供五个不同档位的充值包，档位越大赠送比例越高：
        </p>
        <DocCompareTable
          columns={["价格", "基础豆", "赠送豆", "合计"]}
          labelHeader="档位"
          rows={[
            {
              label: "1 元包",
              values: ["¥1", "1,000", "——", "1,000"],
            },
            {
              label: "5 元包",
              values: ["¥5", "5,000", "——", "5,000"],
            },
            {
              label: "10 元包",
              values: ["¥10", "10,000", "+1,000", "11,000"],
            },
            {
              label: "50 元包",
              values: ["¥50", "50,000", "+7,500", "57,500"],
            },
            {
              label: "100 元包",
              values: ["¥100", "100,000", "+20,000", "120,000"],
            },
          ]}
        />
      </DocSection>

      <DocSection id="tags" title="两个醒目标签">
        <DocCallout variant="tip" title="推荐档位">
          10 元包挂&ldquo;超值推荐&rdquo;标签——第一次充值的用户最容易选，赠送比例约 10%。100 元包挂&ldquo;最划算&rdquo;标签——赠送比例达 20%，适合重度使用 AI 功能的用户。
        </DocCallout>
      </DocSection>
    </>
  );
}
