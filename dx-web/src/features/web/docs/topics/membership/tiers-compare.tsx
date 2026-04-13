import { DocCompareTable } from "@/features/web/docs/primitives/doc-compare-table";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function TiersCompare() {
  return (
    <>
      <DocSection id="five-tiers" title="五档会员一览">
        <p>
          斗学提供五档会员：免费、月度、季度、年度、终身。免费用户可以体验每个游戏的第一关和所有社区功能；付费会员解锁全部关卡、PK、小组创建、AI 随心学，并每月获得赠送的能量豆。
        </p>
      </DocSection>

      <DocSection id="compare" title="完整权益对比">
        <DocCompareTable
          columns={[
            "免费",
            "月度 ¥39",
            "季度 ¥99",
            "年度 ¥309",
            "终身 ¥1999",
          ]}
          labelHeader="权益"
          rows={[
            {
              label: "有效期",
              values: ["永久", "1 个月", "3 个月", "12 个月", "永久"],
            },
            {
              label: "全部关卡",
              values: [false, true, true, true, true],
            },
            {
              label: "PK 对战",
              values: [false, true, true, true, true],
            },
            {
              label: "创建学习小组",
              values: [false, true, true, true, true],
            },
            {
              label: "AI 随心学",
              values: [false, true, true, true, true],
            },
            {
              label: "每月能量豆赠送",
              values: ["——", "10,000", "10,000", "10,000", "15,000"],
            },
          ]}
        />
      </DocSection>
    </>
  );
}
