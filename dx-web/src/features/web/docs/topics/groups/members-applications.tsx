import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function MembersApplications() {
  return (
    <>
      <DocSection id="owner-view" title="组主视角">
        <p>
          作为组主，你在小组页会看到两样组员看不到的东西：完整的成员列表（包括每位成员的学习进度），以及待审核的申请列表。这两个入口都在小组详情页顶部的&ldquo;管理&rdquo;区域。
        </p>
      </DocSection>

      <DocSection id="member-list" title="成员管理">
        <p>
          成员列表里每一行是一个成员，显示头像、昵称、等级、打卡进度。组主可以点击某个成员的操作菜单把他踢出小组——被踢出的成员会收到通知，可以随时通过邀请码重新加入（如果组主允许）。
        </p>
      </DocSection>

      <DocSection id="applications" title="申请管理">
        <p>
          当有人用&ldquo;申请加入&rdquo;方式请求加入你的小组时，申请会出现在这里。每条申请展示申请人的基本信息和可选的留言。组主可以接受或拒绝——接受后申请人立即成为成员。
        </p>
      </DocSection>

      <DocSection id="roles" title="只有两级角色">
        <DocCallout variant="info" title="没有中间管理层">
          斗学的小组只有两种角色：组主（owner）和成员（member）。没有&ldquo;副组主&rdquo;或&ldquo;管理员&rdquo;。所有管理权限都集中在组主一人手里——这是一个有意为之的简化设计，避免小组内部的权限纠纷。
        </DocCallout>
      </DocSection>
    </>
  );
}
