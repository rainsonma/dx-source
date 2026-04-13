import { BookOpen, Brain, Gamepad2, Users } from "lucide-react";
import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocFeatureGrid } from "@/features/web/docs/primitives/doc-feature-grid";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function WhatIsDouxue() {
  return (
    <>
      <DocSection id="what" title="斗学是什么？">
        <p>
          斗学是一款融合游戏化机制与 AI 辅助的英语学习平台。它把背单词、做题、练听说这些传统学习动作，重新设计成有节奏、有反馈、可以和朋友一起玩的游戏流程，让学习这件事本身变得足够有趣，愿意每天坚持。
        </p>
        <p>
          和大多数刷题型 App 不同，斗学强调&ldquo;边玩边学&rdquo;：你不是在完成任务，而是在闯关、对战、协作，学习效果是副产品。
        </p>
      </DocSection>

      <DocSection id="core-values" title="我们相信什么">
        <p>
          我们相信&ldquo;以句子为单位学英语&rdquo;比孤立背单词更接近真实语言的使用方式。我们相信好的学习工具应当是&ldquo;陪你一起玩&rdquo;的，不是&ldquo;监督你打卡&rdquo;的。我们相信重复与反馈是掌握英语的真正核心，而这两件事都可以被游戏化做得更愉快。
        </p>
        <DocCallout variant="tip" title="小贴士">
          斗学支持网页端，在 Chrome、Safari、Edge 等主流浏览器中都能获得最佳体验。
        </DocCallout>
      </DocSection>

      <DocSection id="four-pillars" title="四大核心能力">
        <p>下面这四件事，构成了斗学日常使用的主要场景：</p>
        <DocFeatureGrid
          columns={2}
          items={[
            {
              icon: Gamepad2,
              iconColor: "text-teal-600",
              title: "三种学习模式",
              desc: "单人闯关、PK 对战、小组共学，同一套知识可以用三种不同节奏来学。",
            },
            {
              icon: BookOpen,
              iconColor: "text-blue-600",
              title: "词汇追踪系统",
              desc: "生词本、复习本、已掌握三本书自动联动，间隔重复帮你把词真正记牢。",
            },
            {
              icon: Users,
              iconColor: "text-amber-600",
              title: "社区与小组",
              desc: "在斗学社发帖交流，加入学习小组和朋友一起组局学习。",
            },
            {
              icon: Brain,
              iconColor: "text-purple-600",
              title: "AI 智能创作",
              desc: "给定关键词，让 AI 帮你生成属于你自己的课程内容和词汇关卡。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="who" title="适合谁用">
        <p>
          斗学适合所有想把英语学扎实、但又受不了&ldquo;坐下来背书&rdquo;的人：想提升词汇量的学生、想捡起英语的上班族、希望孩子有趣地学英语的家长、或者单纯觉得刷题太无聊的你。
        </p>
      </DocSection>
    </>
  );
}
