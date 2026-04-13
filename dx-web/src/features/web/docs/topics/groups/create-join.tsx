import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocSlug } from "@/features/web/docs/primitives/doc-slug";
import { DocSteps } from "@/features/web/docs/primitives/doc-steps";

export default function CreateJoin() {
  return (
    <>
      <DocSection id="create" title="创建小组">
        <DocSteps
          steps={[
            {
              title: "确认你是 VIP",
              desc: "只有 VIP 会员可以创建新小组，免费用户会被拦下",
            },
            {
              title: "填写基本信息",
              desc: "给小组起一个名字，可选填一段简短描述",
            },
            {
              title: "生成邀请码和二维码",
              desc: "系统自动为你的小组生成独一无二的邀请码和对应的二维码图片，用于邀请朋友加入",
            },
          ]}
        />
      </DocSection>

      <DocSection id="join" title="加入小组">
        <p>
          斗学提供三种加入小组的方式：输入别人给你的邀请码；扫描小组的二维码；在小组详情页点击&ldquo;申请加入&rdquo;。前两种是直接加入，第三种需要组主审核。
        </p>
      </DocSection>

      <DocSection id="invite-link" title="邀请链接格式">
        <p>
          小组邀请链接的格式是 <DocSlug>/g/[code]</DocSlug>
          ——这是可以直接分享给朋友的完整 URL。对方点开链接会看到小组的基本信息，如果他已经登录斗学就可以直接加入，未登录的话会引导他先登录。
        </p>
      </DocSection>

      <DocSection id="leave" title="退出小组">
        <p>
          普通成员可以随时退出小组——不需要组主同意，也没有冷却时间。但组主不能退出自己创建的小组，必须先解散小组（目前不支持转让组主权限）。
        </p>
        <DocCallout variant="warning" title="组主退出 = 解散小组">
          如果你是组主且不想继续运营这个小组，只能选择解散。解散后所有成员自动退出，小组数据不再可访问。
        </DocCallout>
      </DocSection>
    </>
  );
}
