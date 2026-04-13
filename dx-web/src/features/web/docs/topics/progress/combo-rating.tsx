import { DocCompareTable } from "@/features/web/docs/primitives/doc-compare-table";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function ComboRating() {
  return (
    <>
      <DocSection id="combo" title="连击奖励">
        <p>
          闯关过程中连续答对题目会触发 combo 奖励——连续的长度越长，奖励越多。具体分段如下：
        </p>
        <DocKeyValue
          items={[
            { key: "3 连击", value: "额外 +3 分" },
            { key: "5 连击", value: "额外 +5 分" },
            { key: "10 连击", value: "额外 +10 分" },
          ]}
        />
        <p>
          连击一旦被答错打断就会从 0 重新计数。PK 模式里机器人也会按难度参数有一定概率故意打断自己的连击，这是故意设计的&ldquo;让一下&rdquo;的机制。
        </p>
      </DocSection>

      <DocSection id="rating" title="四档成绩评分">
        <p>
          关卡结束时根据本局的正确率评出一个成绩等级，用不同颜色区分：
        </p>
        <DocCompareTable
          columns={["阈值", "颜色"]}
          labelHeader="等级"
          rows={[
            { label: "优秀", values: ["正确率 ≥ 90%", "teal 青绿"] },
            { label: "良好", values: ["正确率 ≥ 70%", "blue 蓝"] },
            { label: "及格", values: ["正确率 ≥ 60%", "amber 琥珀"] },
            { label: "继续加油", values: ["正确率 < 60%", "rose 玫红"] },
          ]}
        />
      </DocSection>

      <DocSection id="impact" title="评分对经验的影响">
        <p>
          只有达到&ldquo;及格&rdquo;以上的正确率（即 ≥ 60%）才会发放经验奖励。这一阈值适用于所有模式——单人、PK、小组都一样。低于 60% 不会被罚，只是这一局不计入经验。
        </p>
      </DocSection>
    </>
  );
}
