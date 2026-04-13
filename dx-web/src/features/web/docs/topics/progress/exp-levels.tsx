import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function ExpLevels() {
  return (
    <>
      <DocSection id="exp-sources" title="经验值从哪里来">
        <p>
          经验值（EXP）是你在斗学的核心成长值，来源有三种：完成闯关（必须正确率 ≥ 60% 才发放）、完成每日挑战、以及游戏中触发的连击奖励。连续打卡本身不直接产生经验，但它能让每次闯关更有动力。
        </p>
      </DocSection>

      <DocSection id="levels" title="等级范围与成长曲线">
        <p>斗学的等级从 Lv. 0 开始，最高 Lv. 100，遵循一条经过调优的指数曲线：</p>
        <DocKeyValue
          items={[
            { key: "起始等级", value: "Lv. 0" },
            { key: "最高等级", value: "Lv. 100" },
            {
              key: "Lv. 0 → Lv. 1",
              value: "100 EXP",
              note: "固定的入门成本",
            },
            {
              key: "Lv. 2 及以后",
              value: "100 × 1.05^(n−2)",
              note: "指数曲线，越高越难升",
            },
            {
              key: "满级所需累计 EXP",
              value: "约 248,531 EXP",
            },
          ]}
        />
      </DocSection>

      <DocSection id="view" title="在哪里查看">
        <p>
          你的当前等级和经验进度显示在学习大厅的数据行、个人中心的头像边上，以及排行榜里。点击等级进度条可以查看到下一级还差多少经验。
        </p>
      </DocSection>

      <DocSection id="threshold" title="为什么是 60% 正确率才发经验">
        <DocCallout variant="info" title="防止随便刷分">
          如果一关正确率低于 60%，说明你还没真正掌握内容——这种情况下发放经验并不反映真实进步。60% 是一个经验阈值，既不至于苛刻（允许犯错），又能保证经验背后的学习质量。
        </DocCallout>
      </DocSection>
    </>
  );
}
