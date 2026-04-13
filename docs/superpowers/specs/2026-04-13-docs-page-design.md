# /docs Help Center — Design Spec

**Date:** 2026-04-13
**Scope:** `dx-web` frontend only (no backend changes)
**Target route:** `/docs` (replaces existing placeholder)
**Audience:** End-user consumers (learners + creators). Admin-only features are excluded.

## Purpose

Replace the current placeholder `/docs` page with a full, detailed, professional help center covering every consumer-facing feature of Douxue. Every topic is grounded in the current codebase (routes, services, constants, schemas) — no speculation, no editorial prose, no copy lifted from unrelated products.

The help center ships as a multi-page static site under `/docs/[category]/[topic]`, built from a single typed registry so adding, reordering, or renaming topics remains a one-edit operation.

## Current State

The existing `/docs` page is a placeholder mockup with fabricated content:

- `dx-web/src/app/(web)/(home)/docs/page.tsx` — thin wrapper that renders `<DocsPageContent />`
- `dx-web/src/features/web/home/components/docs/docs-content.tsx` — hardcoded single-page TSX referencing a fictional `DouxueSDK`, `Webhook`, and `API 参考` sidebar items

Neither file contains anything salvageable. Both are removed/replaced in this work.

## Goals

1. Cover every consumer feature verified in the codebase, including creator features (any VIP user can use them).
2. Emphasize the three learning modes (single / PK / group) as a prominent first-class category.
3. Professional, consistent with the existing landing page visual language (teal-600 accent, slate base, Lucide icons, shadcn primitives, `rounded-[10px]` cards).
4. Single-source-of-truth IA — sidebar, breadcrumbs, prev/next, metadata, and dynamic routing all read from one typed registry.
5. Every topic statically prerendered at build time.
6. Zero new runtime dependencies (shadcn `accordion` added via `npx shadcn add accordion` if not already present, but that's a source copy, not an npm install).

## Non-Goals / Out of Scope

- Client-side search
- Internationalization (docs ship Chinese-only)
- Real screenshots of the running app (we use `DocPseudoUI` stylized placeholders instead)
- Admin-only features (notice publishing, redeem-code generation)
- Editorial/learning-method essays ("如何陪孩子学英语", "为什么用句子学英语")
- Per-topic feedback (thumbs up/down)
- Page-view analytics
- OG images per topic

## Information Architecture

12 categories, 48 topics total. Every item is verified against the codebase. Categories are ordered to roughly follow a new user's journey: orientation → core gameplay → progression → social → advanced → monetization → account.

### Full IA

**1. 开始使用 (Getting Started)** — 4 topics

- **1.1 认识斗学** — 斗学是什么（英语学习 + 游戏化 + AI）· 核心理念（以句子为单位、边玩边学）· 四大卖点（学习模式 / 词汇追踪 / 社区小组 / AI 辅助）· 适合谁
- **1.2 注册与登录** — 三种方式（邮箱验证码 / 账号密码 / 微信扫码）· 用户名规则（≤30 字符，字母数字 `-_`）· 密码规则（≥8，大小写+数字）· 找回密码流程 · 被邀请注册（ref cookie）
- **1.3 学习首页导览** — Dashboard 区块（问候、通知横幅、每日挑战、游戏进度、今日之星、数据行、学习热力图）· 侧边栏 12 块区分 · 数据行含义（EXP / streak / 掌握 / 待复习）· 每日挑战两个固定任务
- **1.4 新手第一课** — 10 分钟推荐路径：挑游戏 → 选关卡 → 选 beginner + write → 完成答题 → 看结算

**2. 三种学习模式 (Learning Modes) ★ emphasized** — 5 topics

- **2.1 模式总览** — 三种模式对比大表（参与人数 / 对手 / VIP 要求 / 典型场景）· 共享的基础（游戏类型、难度、学习模式）· 如何选模式
- **2.2 单人闯关模式** — 默认模式，所有人可玩 · 启动流程：游戏详情→选关→选 degree/pattern/difficulty · 免费只能玩第一关，VIP 解锁全部 · 会话机制（start→answer/skip→complete→restore）· combo 奖励
- **2.3 PK 对战模式** — 仅 VIP · 两种对战（random 机器人 / specified 真人）· 机器人三档难度表（easy/normal/hard 的正确率 / 延迟 / 连击打断率）· 邀请真人流程（对方需 VIP + 在线）· 对战结算与排名
- **2.4 小组共学模式** — 需要先加入学习小组 · 组主开局 → 全员进入房间 · 两种开局方式（group_solo 个人排名 / group_team 分组对战）· 关卡推进与强制结束
- **2.5 游戏类型与技能矩阵** — 四种游戏（word-sentence / vocab-battle / vocab-match / vocab-elimination）各自玩法 · 三个难度（beginner 全内容 / intermediate block-phrase-sentence / advanced 仅 sentence）· 四种学习模式（听说读写）· 难度×模式矩阵图

**3. 课程与游戏 (Courses & Games)** — 3 topics

- **3.1 挑选游戏** — 游戏广场布局 · 三维筛选（分类树 / 出版社 / 模式）· 名称搜索 · 无限滚动
- **3.2 游戏详情与关卡** — 详情页组成 · 关卡网格 · 首关免费规则 · 从详情进入三种模式
- **3.3 收藏与我的游戏** — 收藏 toggle · 收藏列表 · "我的游戏"（玩过的 + 创建的）

**4. 词汇管理 (Vocabulary)** — 3 topics

- **4.1 生词本** — 如何加入生词 · 页面统计（总数 / 今日新增 / 近 3 天）· 单个/批量删除 · 转入复习本
- **4.2 复习本与间隔重复** — 间隔重复原理 · 间隔表 `[1, 3, 7, 14, 30, 90]` 天 · 三种状态（待复习 / 逾期 / 今日已复习）· 复习后自动推进
- **4.3 已掌握** — 如何标为掌握 · 统计（总量 / 本周 / 本月）· 不再出现在复习中 · 反向取消

**5. 成长与激励 (Progress & Rewards)** — 4 topics

- **5.1 经验与等级** — EXP 来源 · 等级范围 Lv.0 → Lv.100 · 曲线公式（Lv.0→1 = 100 EXP，Lv.2+ 为 `100 × 1.05^(n-2)`）· 最高累计 ≈ 248,531 · 60% 正确率阈值才发 EXP
- **5.2 连击与评分** — Combo 奖励机制（3 / 5 / 10 连击分别 +3 / +5 / +10 分）· 四档评分（≥90% 优秀 teal / ≥70% 良好 blue / ≥60% 及格 amber / <60% 继续加油 rose）
- **5.3 连续打卡** — streak 概念 · 每日凌晨 2 点定时任务的更新规则（昨天有 → +1，前天或更早 → 重置为 1，今天已有 → 不变）· max_play_streak 历史最高
- **5.4 排行榜** — 两指标（EXP / playtime）× 三时段（日 / 周 / 月）· Top 3 上榜台 · 4+ 列表 · 我的排名 · 前 100 名

**6. 斗学社与好友 (Community & Social)** — 4 topics

- **6.1 发帖与评论** — 帖子内容（文本 ≤2000 字 + 可选图片 + 标签 ≤5 个 × ≤20 字）· 评论 ≤500 字 · 不支持嵌套回复 · 编辑 / 软删除
- **6.2 点赞、收藏与关注** — 点赞 / 收藏 / 关注都是 toggle · 单向关注 · 不能关注自己
- **6.3 社区动态流** — 四个 tab（latest / hot / following / bookmarked）· 无限滚动 · 游标分页
- **6.4 个人主页与粉丝** — 从帖子点头像进入 · 展示：昵称、简介、等级、streak、EXP · 对方的帖子流

**7. 学习小组 (Study Groups)** — 5 topics

- **7.1 什么是学习小组** — 小组的作用 · 硬限制（最多 50 人 / 最多 10 子分组 / 创建需 VIP）· 两种玩法（solo / team）
- **7.2 创建与加入** — 创建流程（VIP → 名称 → 自动生成邀请码+QR）· 加入方式（邀请码 / 二维码 / 申请）· 邀请链接格式 `/g/[code]` · 组主不能退出（必须解散）
- **7.3 成员与申请管理** — 组主视角（成员列表、踢人、申请审核）· 成员角色仅 owner / member 两级
- **7.4 子分组 (Subgroups)** — 用于 team 模式 · 最多 10 个 · 分配成员 · 成员最多属一个子分组
- **7.5 开局与推进游戏** — 组主选游戏→选模式→起始关 · 开始游戏 · next-level 推进 · force-end 强制结束 · 结算（小组排名 + 个人分数）

**8. AI 智能学习 (AI Features)** — 2 topics

- **8.1 AI 随心学（句子）** — 从关键词 + 难度（CEFR a1-a2 / b1-b2 / c1-c2）生成故事 · 四步流程（generate → format → break → generate items）· 每步消耗能量豆（5 豆/次，失败退还）· 生成后作为自建课程继续编辑
- **8.2 AI 随心学（词汇）** — 生成词汇列表（vocab 变体）· 同样四步 · 词汇模式限制（match 5 对 / elimination 8 对 / battle 20 对）· 每课最多 20 元数据 × 20 关

**9. 创作课程 (Content Creation)** — 5 topics

- **9.1 创作者入门** — 为什么创作 · 课程结构（关卡 → 单元 → 内容条目）· 两种创建路径（从零 / AI 生成）· 生命周期（draft → published → withdraw）
- **9.2 新建课程** — 基本信息（名称、描述、封面、分类、出版社、模式）· 封面上传规则（≤2MB，JPEG/PNG）
- **9.3 关卡与单元** — 关卡是可玩单位 · 单元是关卡内内容块 · 增/改/重排/删 · 内容类型限制（word/block/phrase/sentence）
- **9.4 内容条目（题目）** — 内容条目是最小单位 · 三种添加方式（手动 / AI 生成 / 从单元分解）· 批量重排 · 单个删除 / 整关清空
- **9.5 发布与撤回** — 发布前检查 · 发布后进入游戏广场 · 撤回（下架不删）· 重新发布

**10. 会员与能量豆 (Membership & Beans)** — 5 topics

- **10.1 会员等级对比** — 五档（free / month ¥39 / season ¥99 / year ¥309 / lifetime ¥1999）· 期限对应 · 完整权益对比大表
- **10.2 会员权益** — 解锁全部关卡 · PK 对战 · 小组创建 · AI 随心学 · 每月能量豆赠送 · 学习服务支持（高等级）
- **10.3 购买流程** — 五步（选等级 → 生成订单 → 选支付 → 扫码 → 权益到账）· 订单状态流（pending → paid → fulfilled）· 30 分钟过期规则 · 支持微信/支付宝
- **10.4 能量豆购买** — 五档包装表（1k / 5k / 10k+1k / 50k+7.5k / 100k+20k）· "超值推荐" / "最划算" 标签 · 赠送比例计算
- **10.5 月度赠送与清零** — 会员赠送节奏（月/季/年会员 10k 豆，终身 15k 豆）· 赠送日按购买月份对应日 · 月末日期特殊处理（29/30/31 → 短月末日）· 清零机制 · 过期会员不再赠送

**11. 邀请与兑换 (Invites & Redeem)** — 3 topics

- **11.1 邀请好友赚佣金** — 推广页面（统计 + 链接 + 二维码）· 四项统计（累计 / 本月新增 / 待激活 / 转化率）· 推广成功条件（好友注册后付费）· 佣金结算
- **11.2 邀请码与群组码** — 用户邀请码 vs 群组邀请码（`/g/[code]`）· 何时显示、何时使用
- **11.3 兑换码** — 兑换入口 · 可兑换内容（会员期限 / 能量豆）· 失败原因（已使用 / 不存在 / 过期）

**12. 账户与帮助 (Account & Support)** — 5 topics

- **12.1 个人资料** — 可编辑项（昵称唯一、城市、简介）· 头像上传（通过 imageId 绑定）· 用户名只读 · 展示位置
- **12.2 账号安全** — 改邮箱（需新邮箱验证码）· 改密码（需当前密码）· 登出 · 多设备 session-replaced 机制（新设备登录踢旧）
- **12.3 通知中心** — 通知列表 · 未读数角标 · 标记已读
- **12.4 提交反馈** — 反馈类型五种（feature / content / ux / bug / other）· 内容字段 · 联系方式可选
- **12.5 常见问题 (FAQ)** — 五个分区用 `<DocFaqAccordion>` 展示：账号与登录 · 会员与支付 · 游戏与学习 · AI 与能量豆 · 技术与兼容性

## Routing & File Structure

### Routes

```
src/app/(web)/(home)/docs/
  layout.tsx                        # Shared shell: LandingHeader + 3-col docs frame + Footer
  page.tsx                          # /docs                    — landing: hero + category grid
  [category]/
    page.tsx                        # /docs/[category]          — category header + topic list
    [topic]/
      page.tsx                      # /docs/[category]/[topic]  — topic page with content, breadcrumb, prev/next
```

Only **3 route files** drive all 48 topic pages. Both dynamic routes export `generateStaticParams` so every URL is statically prerendered at build time.

- Unknown category slug → `notFound()`
- Unknown topic slug → `notFound()`
- Each page exports `generateMetadata` pulling title + description from the registry

**Slug convention:** English kebab-case (`learning-modes/pk-mode`). URLs remain clean when shared. Chinese display labels live in the registry.

### Feature folder layout

```
src/features/web/docs/
  types.ts                          # DocCategory, DocTopic types
  registry.ts                       # Single source of truth (12 cats × 48 topics)

  components/                       # Layout + shell components
    docs-layout.tsx                 # 3-col layout (sidebar · content · TOC)
    docs-sidebar.tsx                # Category + topic nav
    docs-sidebar-drawer.tsx         # Mobile drawer wrapper (shadcn Sheet)
    docs-breadcrumb.tsx             # 文档 / 分类 / 主题 breadcrumb
    docs-prev-next.tsx              # 上一页 / 下一页 buttons
    docs-home-hero.tsx              # /docs landing hero
    docs-category-grid.tsx          # 12-card category grid
    docs-category-index.tsx         # Topic list for /docs/[category]
    docs-toc.tsx                    # Right-rail page TOC (client: scroll-spy)

  primitives/                       # Reusable content blocks
    doc-section.tsx                 # H2 wrapper with anchor id
    doc-callout.tsx                 # info / tip / warning / success variants
    doc-steps.tsx                   # Numbered step list
    doc-feature-grid.tsx            # Icon + title + desc card grid
    doc-compare-table.tsx           # Side-by-side comparison table
    doc-key-value.tsx               # Field / value reference table
    doc-flow.tsx                    # Horizontal chevron flow
    doc-slug.tsx                    # Inline code pill
    doc-link.tsx                    # Styled internal link
    doc-faq-accordion.tsx           # Collapsible Q&A (client)
    doc-pseudo-ui.tsx               # "What you see on screen" mock
    doc-code-block.tsx              # Dark code block

  topics/                           # One file per topic (48 files)
    getting-started/
      what-is-douxue.tsx
      signup-signin.tsx
      hall-tour.tsx
      first-session.tsx
    learning-modes/
      overview.tsx
      single-mode.tsx
      pk-mode.tsx
      group-mode.tsx
      game-types.tsx
    courses-games/
      browsing.tsx
      detail-levels.tsx
      favorites.tsx
    vocabulary/
      unknown.tsx
      review.tsx
      mastered.tsx
    progress/
      exp-levels.tsx
      combo-rating.tsx
      play-streak.tsx
      leaderboard.tsx
    community/
      posts-comments.tsx
      likes-follows.tsx
      feeds.tsx
      profile-view.tsx
    groups/
      overview.tsx
      create-join.tsx
      members-applications.tsx
      subgroups.tsx
      start-game.tsx
    ai/
      ai-custom-sentence.tsx
      ai-custom-vocab.tsx
    creation/
      creator-intro.tsx
      new-course.tsx
      levels-units.tsx
      content-items.tsx
      publish-withdraw.tsx
    membership/
      tiers-compare.tsx
      benefits.tsx
      purchase-flow.tsx
      beans-packages.tsx
      beans-monthly.tsx
    invites/
      referral-program.tsx
      invite-codes.tsx
      redeem-codes.tsx
    account/
      profile-edit.tsx
      security.tsx
      notices.tsx
      feedback.tsx
      faq.tsx
```

### Registry sketch

```ts
// src/features/web/docs/types.ts
import type { LucideIcon } from "lucide-react";

export type DocTopic = {
  slug: string;
  title: string;
  description: string;
  Component: React.ComponentType;
};

export type DocCategory = {
  slug: string;
  title: string;
  description: string;
  icon: LucideIcon;
  accentClass: string; // e.g. "text-teal-600 bg-teal-50 border-teal-200"
  topics: DocTopic[];
};
```

```ts
// src/features/web/docs/registry.ts
import { BookOpen, Swords /* ... */ } from "lucide-react";
import WhatIsDouxue from "./topics/getting-started/what-is-douxue";
import SignupSignin from "./topics/getting-started/signup-signin";
// ... all 48 topic imports

import type { DocCategory } from "./types";

export const DOC_CATEGORIES: DocCategory[] = [
  {
    slug: "getting-started",
    title: "开始使用",
    description: "第一次使用斗学？从这里开始。",
    icon: BookOpen,
    accentClass: "text-teal-600 bg-teal-50 border-teal-200",
    topics: [
      { slug: "what-is-douxue", title: "认识斗学",     description: "...", Component: WhatIsDouxue },
      { slug: "signup-signin",  title: "注册与登录",   description: "...", Component: SignupSignin },
      // ...
    ],
  },
  // ... 11 more categories
];

// Helpers used by routes + components
export type TopicRef = { category: DocCategory; topic: DocTopic };

export function findCategory(slug: string): DocCategory | undefined;
export function findTopic(
  catSlug: string,
  topicSlug: string
): { ref: TopicRef; prev: TopicRef | null; next: TopicRef | null } | undefined;
export function flatTopics(): TopicRef[];
```

### Why one registry

- Sidebar reads `DOC_CATEGORIES` directly
- Breadcrumbs use `findTopic(cat, topic)` for labels
- Prev/Next uses `flatTopics()` for reading-order neighbors (spans the entire help center)
- Landing grid maps over `DOC_CATEGORIES`
- Dynamic route uses `findTopic(...).topic.Component` to render
- `generateStaticParams` enumerates from `DOC_CATEGORIES`
- `generateMetadata` pulls title + description from the registry entry

Adding a topic = 1 component file + 1 registry line. Reordering = reorder registry. Renaming = 1 edit.

### Static generation

Every route is a server component. Every dynamic route exports `generateStaticParams` listing all slug combinations from the registry. Build output: **61 static HTML pages** under `/docs/*` (1 landing + 12 category indexes + 48 topic pages), zero runtime API calls, works with the backend offline.

## Page Layouts

### The shell (`docs/layout.tsx`)

Persistent 3-column frame, ≤1280px centered:

- **Left sidebar (260px)** — Category + topic nav (all categories expanded). Hidden below `lg` (<1024px); opens as a `Sheet` drawer via a sticky "目录" button.
- **Main content (flex-1)** — Varies by page type.
- **Right TOC (220px)** — Scroll-spy page outline. Visible only `≥xl` and only on topic pages. Tiny client component.

Shell also wraps `LandingHeader` (reads `dx_token` cookie to render isLoggedIn state) and `Footer`, replacing the per-page inline copies in the current placeholder.

### Page type 1 — `/docs` landing

- Hero: H1 "斗学帮助中心" + subtitle + two CTA buttons (热门主题 / 常见问题)
- Hot topics row: 3 curated cards (认识斗学 · 模式总览 · 会员等级对比)
- Category grid: 12 cards (3×4 on desktop, 2×6 on tablet, 1×12 on mobile). Each card shows colored icon, title, description, "N 个主题 →"
- Footer strip: "找不到答案？在账户 → 提交反馈告诉我们 →"

### Page type 2 — `/docs/[category]`

- Breadcrumb: `文档 / {category.title}`
- Category header: colored icon circle + H1 + subtitle
- Topic list: numbered rows (number circle + title + one-line description + right chevron), each linking to the topic page
- Bottom nav: `← 返回文档首页` and `下一分类: {next category title} →`

### Page type 3 — `/docs/[category]/[topic]`

- Breadcrumb: `文档 / {category.title} / {topic.title}`
- H1 = topic title
- Lead paragraph (longer than the registry description; written per topic)
- Divider
- Content body: topic component renders a sequence of `<DocSection>` blocks (H2 + anchor) plus primitives
- Prev/Next footer: spans the entire help center reading order (not just within-category). Disabled when at first or last topic.

## Visual Primitives

All server components except `DocFaqAccordion` (needs open/close state) and the scroll-spy TOC wrapper. No new runtime dependencies — shadcn `accordion` is added via `npx shadcn add accordion` if not already present.

| # | Primitive | Purpose |
|---|---|---|
| 1 | `DocSection` | H2 wrapper with explicit anchor id, drives right-rail TOC |
| 2 | `DocCallout` | Boxed callout, 4 variants (info / tip / warning / success) |
| 3 | `DocSteps` | Numbered walkthrough rows (circle + title + desc) |
| 4 | `DocFeatureGrid` | Icon + title + desc card grid |
| 5 | `DocCompareTable` | Side-by-side comparison table (true/false → icons; strings → text) |
| 6 | `DocKeyValue` | Two-column field → value reference table |
| 7 | `DocFlow` | Horizontal flexbox flow (nodes + chevron arrows), no SVG |
| 8 | `DocSlug` | Inline monospace pill for backend slug references |
| 9 | `DocLink` | Styled internal Next.js Link, external variant available |
| 10 | `DocFaqAccordion` | Collapsible Q&A (client component, shadcn Accordion) |
| 11 | `DocPseudoUI` | Stylized "what you see on screen" mock (substitute for screenshots) |
| 12 | `DocCodeBlock` | Dark monospace block (for rare structured-text examples) |

**Section headings:** H1 = topic title (per page). H2 = section heading, tracked by TOC, each authored with an explicit `id` prop. H3+ = subsections, not tracked by TOC.

**Colors:** Stay inside the existing landing palette (teal-600 primary, slate base, amber/rose/blue/emerald for callout variants). No new brand colors.

**Icons:** Exclusively `lucide-react`.

## Implementation Notes

### Entry points

- **Landing footer link** — `src/components/in/footer.tsx` gets a `/docs` link under the resources column (verify footer has a resources column; if not, add one)
- **Hall sidebar link** — Add a "帮助中心" item in the hall sidebar under the account section, linking to `/docs` (opens in same tab)
- **Per-topic feedback prompt** — Each topic page footer includes a small "这篇文档有帮助吗？告诉我们 →" link pointing to the existing feedback page

### Metadata / SEO

- `generateMetadata` in `[category]/page.tsx` and `[category]/[topic]/page.tsx` reads the registry
- Topic page: `<title>` = `"{topic.title} — 斗学帮助中心"`, `<meta description>` = `topic.description`
- Category page: `<title>` = `"{category.title} — 斗学帮助中心"`, description = `category.description`
- Landing page: fixed `"斗学帮助中心"` + hero subtitle
- No OG images in v1

### Files created

- `src/app/(web)/(home)/docs/layout.tsx`
- `src/app/(web)/(home)/docs/page.tsx` (rewritten)
- `src/app/(web)/(home)/docs/[category]/page.tsx`
- `src/app/(web)/(home)/docs/[category]/[topic]/page.tsx`
- `src/features/web/docs/types.ts`
- `src/features/web/docs/registry.ts`
- `src/features/web/docs/components/*.tsx` (9 files)
- `src/features/web/docs/primitives/*.tsx` (12 files)
- `src/features/web/docs/topics/**/*.tsx` (48 files)

### Files modified

- `src/components/in/footer.tsx` — add `/docs` link (if footer has a links section)
- Hall sidebar component (TBD location — plan will pin down) — add 帮助中心 item
- `components/ui/accordion.tsx` — added via `npx shadcn add accordion` if not present

### Files deleted

- `src/features/web/home/components/docs/docs-content.tsx` (placeholder, fully replaced)
- `src/features/web/home/components/docs/` directory if it becomes empty after the delete

## Acceptance Criteria

1. `/docs` loads and shows the hero + 12-card category grid
2. `/docs/getting-started` loads with category header and 4-topic list
3. `/docs/getting-started/what-is-douxue` loads with full content, breadcrumb, prev/next
4. `/docs/learning-modes/pk-mode` renders the PK mode topic with comparison table, callouts, robot difficulty key-value block, and flow diagram
5. `/docs/account/faq` renders the FAQ with 5 collapsible accordion sections
6. Every one of the 48 topic URLs loads without error
7. Unknown category slug (e.g. `/docs/nonsense`) returns 404
8. Unknown topic slug (e.g. `/docs/getting-started/nonsense`) returns 404
9. Active topic highlight in the sidebar follows the current URL
10. Prev/Next footer navigates in reading order across categories
11. Right-rail TOC on a topic page tracks scroll position and highlights active H2
12. On viewport <1024px, sidebar hides and opens as a drawer via the top "目录" button
13. `npm run build` completes without TypeScript errors
14. `npm run lint` completes clean
15. All 48 topics match their content outlines in this spec
16. No new runtime dependencies (shadcn accordion copy is acceptable)
17. Placeholder file `docs-content.tsx` is deleted

## Open Questions / Risks

- **Shadcn Accordion presence** — if not already in `components/ui/accordion.tsx`, it gets added via `npx shadcn add accordion`. This is the only external source copy; zero npm dependencies added.
- **Hall sidebar location** — the implementation plan will identify the exact file hosting the hall sidebar (likely `dx-web/src/features/web/hall/components/hall-sidebar.tsx` or similar) and add the 帮助中心 link there.
- **Footer links section** — `src/components/in/footer.tsx` is checked during planning for an existing links area; if missing, the footer link is added inline.
- **Content length** — "as detailed as possible" means each topic averages ~500–1200 characters of Chinese prose plus primitives. Rough total: 30k–60k characters of content to author. The implementation plan will sequence this across multiple work units so reviewers can approve incrementally.
- **Reading-order Prev/Next** — spans the entire 48-topic flat list (not within-category). If user navigates via Prev/Next linearly, they read the whole help center top-to-bottom.

## Verification Plan

Since docs are static content, the verification relies on build + manual route checks rather than unit tests:

1. `npm run build` — catches TS errors in topic components, registry imports, `generateStaticParams` output
2. `npm run lint` — clean
3. `npm run dev` then manually load:
   - `/docs` (landing)
   - `/docs/getting-started` (category index)
   - `/docs/getting-started/what-is-douxue` (first topic)
   - `/docs/learning-modes/pk-mode` (primitive-heavy topic)
   - `/docs/account/faq` (accordion topic)
   - `/docs/nonsense` (404)
   - `/docs/getting-started/nonsense` (404)
4. Resize browser to <1024px and verify sidebar drawer works
5. Click through prev/next to verify reading order
6. Click sidebar links to verify active highlight
7. Scroll a long topic to verify right TOC scroll-spy tracking

No unit tests are added. The content is static prose; Next.js handles routing; primitives are tiny presentational components. Tests would verify implementation details, not behavior users care about.
