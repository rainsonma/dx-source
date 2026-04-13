import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocLink } from "@/features/web/docs/primitives/doc-link";
import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocSteps } from "@/features/web/docs/primitives/doc-steps";

export default function FirstSession() {
  return (
    <>
      <DocSection id="why" title="为什么从这里开始">
        <p>
          如果这是你第一次用斗学，不用纠结&ldquo;怎么开始&rdquo;——跟着下面这条 10 分钟的路径走一遍，你会跑通从挑游戏、选关卡、到拿经验奖励的完整流程，对产品形成完整的第一印象。
        </p>
      </DocSection>

      <DocSection id="steps" title="10 分钟跑通学习流程">
        <DocSteps
          steps={[
            {
              title: "打开游戏页",
              desc: "在侧边栏点击&ldquo;游戏&rdquo;，按分类或出版社挑一个你感兴趣的课程",
            },
            {
              title: "进入游戏详情",
              desc: "点击任意游戏卡片，查看游戏简介和它的全部关卡",
            },
            {
              title: "选第一关",
              desc: "免费用户只能玩每个游戏的第一关，所以直接点第一关进入即可",
            },
            {
              title: "选择难度和学习模式",
              desc: "推荐初级 + 写，是门槛最低的组合，熟悉玩法后再换其它搭配",
            },
            {
              title: "开始闯关",
              desc: "按提示答题，正确率达到 60% 才能获得经验值，尽力就好",
            },
          ]}
        />
      </DocSection>

      <DocSection id="after" title="结算页能看到什么">
        <p>
          关卡结束后会弹出结算页，展示你的总得分、成绩评级（优秀 / 良好 / 及格 / 继续加油）、本局获得的经验值，以及错题回顾——你可以在这里回看哪些题答错了。如果这一关还有下一关，页面上会有&ldquo;下一关&rdquo;按钮直接继续。
        </p>
      </DocSection>

      <DocSection id="what-next" title="接下来可以做什么">
        <p>
          打完第一关只是开始。你可以先看看{" "}
          <DocLink href="/docs/learning-modes/overview">三种学习模式</DocLink>
          {" "}
          了解单人、PK、小组之间的差异；也可以直接尝试一下 PK
          模式，挑战机器人对手会比单人更紧张刺激（PK 需要 VIP 会员）。
        </p>
        <DocCallout variant="tip" title="慢慢来，不急">
          好玩的游戏自己会把你带回来，不需要一次刷完。比起一天硬刷两小时，每天 10-15 分钟&ldquo;想再来一局&rdquo;的那种冲动，才是真正能让英语长进的方式。
        </DocCallout>
      </DocSection>
    </>
  );
}
