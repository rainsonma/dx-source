import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function LikesFollows() {
  return (
    <>
      <DocSection id="like" title="点赞">
        <p>
          每篇帖子右下角都有点赞按钮。点击一次点赞，再点一次取消——是一个 toggle 动作，没有&ldquo;点踩&rdquo;这种反向操作。点赞数实时更新。
        </p>
      </DocSection>

      <DocSection id="bookmark" title="收藏帖子">
        <p>
          除了点赞，你还可以收藏帖子——收藏是&ldquo;私人&rdquo;的，帖子作者不会看到。收藏后这条帖子会出现在社区动态流的&ldquo;我收藏的&rdquo;tab 里，方便之后回看。
        </p>
      </DocSection>

      <DocSection id="follow" title="关注其他用户">
        <p>
          在帖子作者或任意用户的个人主页可以关注该用户。关注是单向的——你关注对方并不自动让对方关注你，和微博的关注模型一致。关注之后可以在&ldquo;关注中&rdquo;动态流里看到对方发过的帖子。
        </p>
        <DocCallout variant="info" title="不能关注自己">
          系统不允许关注自己——这是一个基本约束。除此之外没有其它限制，关注人数没有上限。
        </DocCallout>
      </DocSection>
    </>
  );
}
