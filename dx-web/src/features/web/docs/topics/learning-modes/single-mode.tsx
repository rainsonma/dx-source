import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocFlow } from "@/features/web/docs/primitives/doc-flow";
import { DocLink } from "@/features/web/docs/primitives/doc-link";
import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocSteps } from "@/features/web/docs/primitives/doc-steps";

export default function SingleMode() {
  return (
    <>
      <DocSection id="what" title="什么是单人闯关">
        <p>
          单人闯关是斗学的默认模式，所有用户都可以玩——不需要 VIP，不需要在线对手。你面对的是系统，目标是完成每一关的题目、积累经验、解锁下一关。
        </p>
        <p>
          相比 PK 和小组的实时对抗，单人更安静、更自由：没有倒计时压力，可以按自己的节奏来，随时中途退出下次继续。
        </p>
      </DocSection>

      <DocSection id="start" title="如何开始一局">
        <DocSteps
          steps={[
            {
              title: "挑游戏",
              desc: "从游戏广场或你的收藏中选一个课程包",
            },
            {
              title: "进入游戏详情",
              desc: "在详情页选择要挑战的关卡",
            },
            {
              title: "选择 degree / pattern / difficulty",
              desc: "初/中/高级 × 听/说/读/写 × 简单/普通/困难，按需组合",
            },
            {
              title: "开始答题",
              desc: "按游戏类型答题，连续答对可以触发 combo 奖励",
            },
            {
              title: "查看结算",
              desc: "关卡结束后系统展示得分、评级、经验奖励和错题回顾",
            },
          ]}
        />
      </DocSection>

      <DocSection id="levels" title="关卡解锁规则">
        <p>
          免费用户可以玩每个游戏的第一关；其余关卡需要开通 VIP 会员才能解锁。这个设计是为了让你先试后买，第一关就能完整体验游戏的玩法和节奏。
        </p>
        <DocCallout variant="warning" title="非 VIP 只能玩第一关">
          如果你在非第一关点击开始，系统会提示需要升级会员才能继续。第一关本身没有任何功能限制，玩法和 VIP 用户完全一样。
        </DocCallout>
      </DocSection>

      <DocSection id="session-lifecycle" title="会话机制">
        <p>
          每一次开始闯关，斗学都会在后端记录一个&ldquo;会话&rdquo;——它包含你的进度、得分、连击、剩余内容等状态。整个生命周期如下：
        </p>
        <DocFlow
          nodes={[
            { label: "start", desc: "开始会话" },
            { label: "answer / skip", desc: "答题或跳过" },
            { label: "complete", desc: "关卡完成" },
          ]}
        />
        <p>
          如果你中途退出（比如关了浏览器），下一次重新进入同一关的同一组参数（同样的难度、学习模式），系统会自动恢复你的会话，从上次的位置继续，不会丢失进度。
        </p>
      </DocSection>

      <DocSection id="combo" title="连击和评分">
        <p>
          连续答对题目会触发 combo 奖励，3 / 5 / 10 连击分别给额外分数。关卡结束后根据正确率评出四档成绩（优秀 / 良好 / 及格 / 继续加油），只有达到 60% 正确率才会发放经验奖励。完整规则见{" "}
          <DocLink href="/wiki/progress/combo-rating">连击与评分</DocLink>
          。
        </p>
      </DocSection>
    </>
  );
}
