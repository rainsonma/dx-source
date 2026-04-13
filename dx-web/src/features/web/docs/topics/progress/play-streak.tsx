import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function PlayStreak() {
  return (
    <>
      <DocSection id="what" title="什么是连续打卡">
        <p>
          连续打卡（play streak）是你连续每天都有学习记录的天数。每天只要完成至少一次任意学习动作（闯关、PK、小组游戏都算），当天的打卡就算完成了，第二天继续就能累积。
        </p>
        <p>
          连续打卡不发放直接的经验奖励，它的作用是&ldquo;看得见的坚持&rdquo;——每次登录看到自己已经连续 N 天，是真正能推动你今天也打开来学一会儿的动力。
        </p>
      </DocSection>

      <DocSection id="update-rules" title="每日更新规则">
        <p>
          每天凌晨 2 点，斗学会跑一次全站的 streak 更新任务，按以下规则调整：
        </p>
        <DocKeyValue
          items={[
            {
              key: "昨天有学习记录",
              value: "streak + 1",
              note: "连续成功",
            },
            {
              key: "前天或更早才有记录",
              value: "streak 重置为 1",
              note: "连续中断",
            },
            {
              key: "今天已经学过了",
              value: "当天不再变动",
              note: "重复登录不会加码",
            },
          ]}
        />
      </DocSection>

      <DocSection id="max" title="历史最高纪录">
        <p>
          除了当前 streak，系统还记录你的最高 streak（max_play_streak）——这是一个&ldquo;只增不减&rdquo;的数字，即使你中断过，它会永远显示你曾经达到过的最长连续天数。
        </p>
      </DocSection>

      <DocSection id="tip" title="维持连续的窍门">
        <DocCallout variant="tip" title="一关也算">
          保持连续不需要每天玩很久，完成一次任何学习动作即可——包括一次 PK、一次闯关、甚至帮组员完成一次小组游戏。忙的时候 5 分钟也能撑起一天的打卡。
        </DocCallout>
      </DocSection>
    </>
  );
}
