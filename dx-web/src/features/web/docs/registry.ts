import { BookMarked, BookOpen, Library, Swords, TrendingUp } from "lucide-react";
import type { DocCategory, TopicRef } from "./types";
import WhatIsDouxue from "./topics/getting-started/what-is-douxue";
import SignupSignin from "./topics/getting-started/signup-signin";
import HallTour from "./topics/getting-started/hall-tour";
import FirstSession from "./topics/getting-started/first-session";
import LearningModesOverview from "./topics/learning-modes/overview";
import SingleMode from "./topics/learning-modes/single-mode";
import PkMode from "./topics/learning-modes/pk-mode";
import GroupMode from "./topics/learning-modes/group-mode";
import GameTypes from "./topics/learning-modes/game-types";
import Browsing from "./topics/courses-games/browsing";
import DetailLevels from "./topics/courses-games/detail-levels";
import Favorites from "./topics/courses-games/favorites";
import Unknown from "./topics/vocabulary/unknown";
import Review from "./topics/vocabulary/review";
import Mastered from "./topics/vocabulary/mastered";
import ExpLevels from "./topics/progress/exp-levels";
import ComboRating from "./topics/progress/combo-rating";
import PlayStreak from "./topics/progress/play-streak";
import LeaderboardTopic from "./topics/progress/leaderboard";

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
  {
    slug: "learning-modes",
    title: "三种学习模式",
    description: "单人、PK、小组三种玩法的完整介绍，斗学的核心玩法都在这里。",
    icon: Swords,
    accentClass: "border-rose-200 bg-rose-50 text-rose-600",
    topics: [
      {
        slug: "overview",
        title: "模式总览",
        description: "三种模式的对比表：参与人数、对手、VIP 要求、典型场景。",
        Component: LearningModesOverview,
      },
      {
        slug: "single-mode",
        title: "单人闯关模式",
        description:
          "默认模式，所有人可玩。了解如何开始、关卡解锁规则、会话机制。",
        Component: SingleMode,
      },
      {
        slug: "pk-mode",
        title: "PK 对战模式",
        description:
          "与真人或 AI 机器人实时对战，抢先完成关卡获胜。需要 VIP。",
        Component: PkMode,
      },
      {
        slug: "group-mode",
        title: "小组共学模式",
        description:
          "加入学习小组后一起闯关，支持个人排名和分组对战。",
        Component: GroupMode,
      },
      {
        slug: "game-types",
        title: "游戏类型与技能矩阵",
        description:
          "四种游戏类型 × 三个难度 × 四种学习模式，练习组合丰富。",
        Component: GameTypes,
      },
    ],
  },
  {
    slug: "courses-games",
    title: "课程与游戏",
    description: "如何挑游戏、查看关卡、管理收藏。",
    icon: Library,
    accentClass: "border-blue-200 bg-blue-50 text-blue-600",
    topics: [
      {
        slug: "browsing",
        title: "挑选游戏",
        description: "游戏广场的筛选、搜索和卡片结构一览。",
        Component: Browsing,
      },
      {
        slug: "detail-levels",
        title: "游戏详情与关卡",
        description: "游戏详情页、关卡网格、首关免费规则。",
        Component: DetailLevels,
      },
      {
        slug: "favorites",
        title: "收藏与我的游戏",
        description:
          "收藏游戏、查看收藏列表、区分&ldquo;玩过&rdquo;和&ldquo;我创建&rdquo;。",
        Component: Favorites,
      },
    ],
  },
  {
    slug: "vocabulary",
    title: "词汇管理",
    description: "生词本、复习本、已掌握三本书自动联动，让词汇真正留下来。",
    icon: BookMarked,
    accentClass: "border-emerald-200 bg-emerald-50 text-emerald-600",
    topics: [
      {
        slug: "unknown",
        title: "生词本",
        description: "收集生词、查看统计、转入复习本。",
        Component: Unknown,
      },
      {
        slug: "review",
        title: "复习本与间隔重复",
        description: "按 [1, 3, 7, 14, 30, 90] 天节奏做间隔重复复习。",
        Component: Review,
      },
      {
        slug: "mastered",
        title: "已掌握",
        description: "标记已掌握的词，查看统计，可反向取消。",
        Component: Mastered,
      },
    ],
  },
  {
    slug: "progress",
    title: "成长与激励",
    description: "经验、等级、连击、打卡、排行榜——看得见的成长路径。",
    icon: TrendingUp,
    accentClass: "border-amber-200 bg-amber-50 text-amber-600",
    topics: [
      {
        slug: "exp-levels",
        title: "经验与等级",
        description: "经验来源、Lv. 0 → Lv. 100 成长曲线、60% 正确率阈值。",
        Component: ExpLevels,
      },
      {
        slug: "combo-rating",
        title: "连击与评分",
        description: "连击奖励和四档评分（优秀 / 良好 / 及格 / 继续加油）。",
        Component: ComboRating,
      },
      {
        slug: "play-streak",
        title: "连续打卡",
        description: "streak 每日更新规则和历史最高记录。",
        Component: PlayStreak,
      },
      {
        slug: "leaderboard",
        title: "排行榜",
        description: "EXP / playtime × 日 / 周 / 月 的六种榜单组合。",
        Component: LeaderboardTopic,
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
