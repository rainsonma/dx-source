import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function ProfileEdit() {
  return (
    <>
      <DocSection id="editable" title="可以编辑的资料字段">
        <DocKeyValue
          items={[
            {
              key: "昵称",
              value: "全站唯一",
              note: "可以随时修改",
            },
            {
              key: "城市",
              value: "自由填写",
              note: "展示在个人主页",
            },
            {
              key: "简介",
              value: "自由文本",
              note: "一两句话自我介绍",
            },
            {
              key: "头像",
              value: "上传图片",
              note: "通过 imageUrl 绑定到你的账号",
            },
          ]}
        />
      </DocSection>

      <DocSection id="username" title="用户名是只读的">
        <DocCallout variant="info" title="为什么不能改用户名">
          用户名在注册时就确定，之后系统不允许修改。这是因为用户名是你的&ldquo;身份 ID&rdquo;，很多场景都依赖它来识别你。如果你想展示不同的名字，修改昵称即可——昵称才是展示在页面上的。
        </DocCallout>
      </DocSection>

      <DocSection id="where-shown" title="这些信息在哪里会被看到">
        <p>
          你的昵称、头像、简介和等级会出现在：个人中心、排行榜、社区帖子（作为作者信息）、小组成员列表、关注和粉丝列表。城市只在个人主页上展示。简介只在个人主页和你发帖时的作者悬浮卡里出现。
        </p>
      </DocSection>
    </>
  );
}
