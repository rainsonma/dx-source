import {
  Coins,
  HeadphonesIcon,
  Sparkles,
  Swords,
  Unlock,
  Users2,
} from "lucide-react";
import { DocFeatureGrid } from "@/features/web/docs/primitives/doc-feature-grid";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function Benefits() {
  return (
    <>
      <DocSection id="benefits" title="会员解锁的全部能力">
        <p>付费会员在斗学里能解锁下面这几件事：</p>
        <DocFeatureGrid
          columns={2}
          items={[
            {
              icon: Unlock,
              iconColor: "text-teal-600",
              title: "全部关卡",
              desc: "免费用户只能玩每个游戏的第一关，会员解锁之后的所有关卡。",
            },
            {
              icon: Swords,
              iconColor: "text-rose-600",
              title: "PK 对战",
              desc: "所有 PK 模式（和机器人、和真人）都需要 VIP。",
            },
            {
              icon: Users2,
              iconColor: "text-violet-600",
              title: "创建学习小组",
              desc: "加入别人的小组免费，但创建自己的小组需要会员。",
            },
            {
              icon: Sparkles,
              iconColor: "text-purple-600",
              title: "AI 随心学",
              desc: "让 AI 根据关键词生成课程的能力，句子版和词汇版都需要会员。",
            },
            {
              icon: Coins,
              iconColor: "text-amber-600",
              title: "每月能量豆赠送",
              desc: "月/季/年会员每月 10,000 豆，终身会员每月 15,000 豆。",
            },
            {
              icon: HeadphonesIcon,
              iconColor: "text-blue-600",
              title: "学习服务支持",
              desc: "中高等级会员在遇到问题时可以获得优先响应。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="tiers-equal" title="各档会员的权益差异">
        <p>
          除了&ldquo;终身会员每月 15,000 豆而非 10,000 豆&rdquo;这一条，其余权益在所有付费档位之间是完全一致的。也就是说，月度、季度、年度三档的区别只在于时长和总价，解锁的功能完全相同。
        </p>
      </DocSection>
    </>
  );
}
