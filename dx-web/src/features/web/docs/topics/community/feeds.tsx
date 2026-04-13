import { Bookmark, Flame, Heart, Newspaper } from "lucide-react";
import { DocFeatureGrid } from "@/features/web/docs/primitives/doc-feature-grid";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function Feeds() {
  return (
    <>
      <DocSection id="tabs" title="四种动态流视角">
        <p>
          斗学社首页顶部有四个 tab，分别对应四种不同的内容呈现方式：
        </p>
        <DocFeatureGrid
          columns={2}
          items={[
            {
              icon: Newspaper,
              iconColor: "text-teal-600",
              title: "最新 (latest)",
              desc: "按发布时间倒序，最新的帖子排在最前面。",
            },
            {
              icon: Flame,
              iconColor: "text-rose-600",
              title: "热门 (hot)",
              desc: "按互动热度（点赞、评论、收藏）排序的精选内容。",
            },
            {
              icon: Heart,
              iconColor: "text-pink-600",
              title: "关注中 (following)",
              desc: "只显示你关注的用户发的帖子。",
            },
            {
              icon: Bookmark,
              iconColor: "text-amber-600",
              title: "我收藏的 (bookmarked)",
              desc: "你收藏过的帖子合集，方便回看。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="pagination" title="无限滚动与游标分页">
        <p>
          四个 tab 都使用无限滚动——滚到底部时自动加载更多帖子，不需要手动翻页。底层用的是游标分页（cursor pagination），这意味着每次加载只请求下一段数据，新发布的帖子不会插入到你正在浏览的位置。
        </p>
      </DocSection>
    </>
  );
}
