import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocLink } from "@/features/web/docs/primitives/doc-link";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function DetailLevels() {
  return (
    <>
      <DocSection id="detail" title="游戏详情页">
        <p>
          从游戏广场点任意卡片进入的是游戏详情页。详情页顶部是封面和基本信息（名称、简介、作者、游戏模式），下方是这个游戏的全部关卡列表。这是你开始一局的入口。
        </p>
      </DocSection>

      <DocSection id="level-grid" title="关卡网格">
        <p>
          关卡按顺序排列成网格或列表。每一关都是一个独立的学习单位，通常围绕一个主题（一组相关词汇、一个语法点、一段故事）。你可以从任意关卡进入，不必按顺序玩——只要解锁状态允许。
        </p>
      </DocSection>

      <DocSection id="first-level-free" title="首关免费，其余需 VIP">
        <p>
          斗学对所有用户开放每个游戏的第一关。这一关不是&ldquo;试玩版&rdquo;，它和 VIP 版的第一关完全一样——有完整的题目、完整的奖励、完整的结算。这样设计是为了让你先试后买：真正玩过再决定要不要开通会员。
        </p>
        <DocCallout variant="warning" title="非首关需要 VIP">
          每个游戏的第 2 关及以后都需要开通 VIP 会员才能进入。点击会触发会员升级提示。
        </DocCallout>
      </DocSection>

      <DocSection id="three-modes-from-here" title="从这里进入多种模式">
        <p>
          在详情页选好关卡后，你可以选择以单人、PK 或小组模式开始本关。具体差异见{" "}
          <DocLink href="/wiki/learning-modes/overview">多种学习模式总览</DocLink>
          。
        </p>
      </DocSection>
    </>
  );
}
