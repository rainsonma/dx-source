import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function ProfileView() {
  return (
    <>
      <DocSection id="access" title="进入个人主页">
        <p>
          在任意帖子下方点击作者头像或昵称，就能进入这位用户的个人主页。你自己的个人中心是独立入口（侧边栏&ldquo;我的&rdquo;），但展示的内容类似。
        </p>
      </DocSection>

      <DocSection id="shows" title="个人主页展示什么">
        <DocKeyValue
          items={[
            { key: "昵称", value: "用户的显示名" },
            { key: "简介", value: "用户自己填写的介绍" },
            { key: "等级和经验", value: "当前 Lv. 和 EXP 进度" },
            {
              key: "连续打卡",
              value: "当前 streak 和历史最高",
            },
            {
              key: "关注 / 粉丝数",
              value: "双向计数",
              note: "点击可查看名单",
            },
          ]}
        />
      </DocSection>

      <DocSection id="posts" title="对方发过的帖子">
        <p>
          个人主页下方会展示这位用户发过的所有帖子，按时间倒序排列。你可以点赞、评论、关注这位用户，所有操作和在动态流里是一样的。
        </p>
      </DocSection>
    </>
  );
}
