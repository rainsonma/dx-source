import { BookOpen } from "lucide-react";
import type { DocCategory, TopicRef } from "./types";
import WhatIsDouxue from "./topics/getting-started/what-is-douxue";
import SignupSignin from "./topics/getting-started/signup-signin";
import HallTour from "./topics/getting-started/hall-tour";
import FirstSession from "./topics/getting-started/first-session";

export const DOC_CATEGORIES: DocCategory[] = [
  {
    slug: "getting-started",
    title: "开始使用",
    description: "第一次使用斗学？从这里开始。",
    icon: BookOpen,
    accentClass: "border-teal-200 bg-teal-50 text-teal-600",
    topics: [
      {
        slug: "what-is-douxue",
        title: "认识斗学",
        description:
          "斗学是一款融合游戏化机制与 AI 辅助的英语学习平台。来了解它的核心理念和四大能力。",
        Component: WhatIsDouxue,
      },
      {
        slug: "signup-signin",
        title: "注册与登录",
        description:
          "邮箱验证码、账号密码、微信扫码三种方式都可以，加上完整的账号规则和忘记密码流程。",
        Component: SignupSignin,
      },
      {
        slug: "hall-tour",
        title: "学习首页导览",
        description:
          "登录进入大厅后，你会看到仪表盘、侧边栏、数据行、每日挑战。这里是每一块区域的含义。",
        Component: HallTour,
      },
      {
        slug: "first-session",
        title: "新手第一课",
        description:
          "10 分钟跑通第一次学习：挑游戏、选关卡、选难度、完成答题、看结算。",
        Component: FirstSession,
      },
    ],
  },
];

export function findCategory(slug: string): DocCategory | undefined {
  return DOC_CATEGORIES.find((c) => c.slug === slug);
}

export function flatTopics(): TopicRef[] {
  return DOC_CATEGORIES.flatMap((category) =>
    category.topics.map((topic) => ({ category, topic })),
  );
}

export function findTopic(
  catSlug: string,
  topicSlug: string,
):
  | { ref: TopicRef; prev: TopicRef | null; next: TopicRef | null }
  | undefined {
  const flat = flatTopics();
  const index = flat.findIndex(
    ({ category, topic }) =>
      category.slug === catSlug && topic.slug === topicSlug,
  );
  if (index === -1) return undefined;
  return {
    ref: flat[index],
    prev: index > 0 ? flat[index - 1] : null,
    next: index < flat.length - 1 ? flat[index + 1] : null,
  };
}
