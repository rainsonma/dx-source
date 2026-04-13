import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function Favorites() {
  return (
    <>
      <DocSection id="favorite" title="如何收藏游戏">
        <p>
          在游戏卡片的角落或详情页顶部都有&ldquo;收藏&rdquo;按钮，点一下即加入收藏，再点一下取消。这是一个纯 toggle 动作，没有任何额外确认。
        </p>
      </DocSection>

      <DocSection id="favorites-page" title="收藏列表">
        <p>
          侧边栏的&ldquo;收藏&rdquo;入口会展示你所有已收藏的游戏。列表按收藏时间倒序排列，最近加入的在最上面，方便随时返回。
        </p>
      </DocSection>

      <DocSection id="my-games" title="我的游戏">
        <p>
          &ldquo;我的游戏&rdquo;页面分成两部分：一部分是你玩过的游戏（按最后一次学习时间排序），另一部分是你自己创建的课程。玩过的游戏是自动记录的，不需要手动收藏。
        </p>
      </DocSection>

      <DocSection id="no-limit" title="没有数量上限">
        <p>
          收藏没有硬性上限，你可以把任何觉得可能会再玩的游戏加进去。和&ldquo;我的游戏&rdquo;不同，收藏是你主动选择保留的列表，不会被任何自动操作清理。
        </p>
      </DocSection>
    </>
  );
}
