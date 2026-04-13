import { Clock, Zap } from "lucide-react";
import { DocFeatureGrid } from "@/features/web/docs/primitives/doc-feature-grid";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function Leaderboard() {
  return (
    <>
      <DocSection id="what" title="排行榜维度">
        <p>
          斗学的排行榜从两个维度切分：2 个指标 × 3 个时段 = 6 张榜。你可以根据自己关心的是&ldquo;经验增长&rdquo;还是&ldquo;投入时长&rdquo;，以及看的是短期表现还是长期积累，在榜单之间切换。
        </p>
      </DocSection>

      <DocSection id="metrics" title="两个指标">
        <DocFeatureGrid
          columns={2}
          items={[
            {
              icon: Zap,
              iconColor: "text-teal-600",
              title: "经验值 EXP",
              desc: "按一段时间内获得的经验总和排序，反映学习成效。",
            },
            {
              icon: Clock,
              iconColor: "text-blue-600",
              title: "游戏时长 playtime",
              desc: "按一段时间内的学习秒数排序，反映时间投入。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="periods" title="三个时段">
        <p>
          三个时段分别是日榜、周榜、月榜。边界按自然日、自然周、自然月切分——周一到周日是一周，月初到月末是一月。每天凌晨系统会更新前一日的数据。
        </p>
      </DocSection>

      <DocSection id="display" title="榜单如何展示">
        <p>
          每张榜单顶部是前 3 名的&ldquo;上榜台&rdquo;视觉展示，金银铜三色区分；第 4 名开始是普通列表，最多展示前 100 名。如果你不在前 100，榜单底部会悬浮显示你当前的排名，方便知道自己离上榜还差多少。
        </p>
      </DocSection>
    </>
  );
}
