import { BookOpen, Building2, Puzzle } from "lucide-react";
import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocFeatureGrid } from "@/features/web/docs/primitives/doc-feature-grid";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function Browsing() {
  return (
    <>
      <DocSection id="layout" title="游戏广场长什么样">
        <p>
          游戏广场是斗学最核心的&ldquo;内容入口&rdquo;。它由顶部的筛选栏和下方的卡片网格组成：左边可以按不同维度筛选，右边是一排排可滚动的游戏卡片，滚到底部自动加载更多，不需要翻页。
        </p>
      </DocSection>

      <DocSection id="filters" title="三个筛选维度">
        <p>筛选栏支持三种维度，互相之间不冲突，可以同时启用多个：</p>
        <DocFeatureGrid
          columns={3}
          items={[
            {
              icon: BookOpen,
              iconColor: "text-teal-600",
              title: "按分类",
              desc: "树形分类浏览（如 K12 / 雅思 / 日常会话），支持多层级。",
            },
            {
              icon: Building2,
              iconColor: "text-blue-600",
              title: "按出版社",
              desc: "按教材或内容方筛选，找自己熟悉的课程来源。",
            },
            {
              icon: Puzzle,
              iconColor: "text-amber-600",
              title: "按游戏模式",
              desc: "连词成句 / 词汇对轰 / 词汇配对 / 词汇消消乐，四选一。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="search" title="按名称搜索">
        <p>
          筛选栏旁边有一个搜索框，支持按游戏名称的关键字搜索——输入几个字即可从几千个游戏中找到目标。搜索和筛选可以同时使用。
        </p>
      </DocSection>

      <DocSection id="cards" title="游戏卡片上的信息">
        <p>
          每张游戏卡片展示：封面图、游戏名称、作者、所属分类、游戏模式。点击卡片进入游戏详情页查看关卡和更多信息。
        </p>
        <DocCallout variant="tip" title="喜欢的游戏可以收藏">
          在卡片或详情页点击收藏按钮，这个游戏就会出现在侧边栏的&ldquo;收藏&rdquo;列表里，方便下次直接找到。
        </DocCallout>
      </DocSection>
    </>
  );
}
