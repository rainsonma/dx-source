import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function HallTour() {
  return (
    <>
      <DocSection id="overview" title="学习大厅长什么样">
        <p>
          登录后你会进入学习大厅，它是整个斗学的中枢。大厅首页从上到下依次是：个性化问候语、系统通知横幅、今日每日挑战卡片、最近游戏进度、今日之星、关键数据行，以及一年的学习热力图。
        </p>
        <p>
          左侧是常驻的侧边栏，包含 12 组功能入口；顶部栏有通知中心、个人菜单。不管你在哪个页面，这两处都是一直可用的。
        </p>
      </DocSection>

      <DocSection id="sidebar" title="侧边栏怎么用">
        <p>
          侧边栏把所有功能按场景分组：首页、游戏、小组、社区、排行、我的、收藏；学习工具（复习、生词、已掌握）；AI 与创作工具；账户入口。鼠标悬停能看到每项的完整名称，点击直接跳转。
        </p>
      </DocSection>

      <DocSection id="stats-row" title="数据行代表什么">
        <p>首页顶部的一行数字是你最关心的四项学习指标：</p>
        <DocKeyValue
          items={[
            {
              key: "经验值 (EXP)",
              value: "闯关和每日挑战累积的成长值，用于升级",
            },
            {
              key: "连续打卡",
              value: "当前已连续学习多少天",
              note: "还会展示你的历史最高纪录",
            },
            {
              key: "掌握单词",
              value: "本周新增 / 累计总量",
            },
            {
              key: "待复习",
              value: "今天需要复习的条目数量",
            },
          ]}
        />
      </DocSection>

      <DocSection id="daily-challenge" title="每日挑战">
        <p>
          首页中部的&ldquo;每日挑战&rdquo;卡片包含两个每天固定的任务：完成任意一次连词成句闯关，可获得双倍经验；在斗学社发一条帖子，算作当天的社区互动。两个任务都不强制完成，但做完能加速成长。
        </p>
        <DocCallout variant="info" title="每天重置一次">
          每日挑战在每天的自然日切换时重置。即使你前一天已经完成了，新的一天还会再次出现可领取状态。
        </DocCallout>
      </DocSection>
    </>
  );
}
