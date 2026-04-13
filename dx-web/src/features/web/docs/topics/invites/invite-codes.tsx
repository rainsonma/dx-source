import { DocCompareTable } from "@/features/web/docs/primitives/doc-compare-table";
import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocSlug } from "@/features/web/docs/primitives/doc-slug";

export default function InviteCodes() {
  return (
    <>
      <DocSection id="two-codes" title="斗学里的两种邀请码">
        <p>
          斗学有两种邀请码，它们看起来像但作用完全不同：一种是&ldquo;用户邀请码&rdquo;，每个账号一个，用于推广——朋友通过它注册视为你的推广；另一种是&ldquo;小组邀请码&rdquo;，每个学习小组一个，用于邀请别人加入你的小组。
        </p>
      </DocSection>

      <DocSection id="compare" title="两种邀请码的区别">
        <DocCompareTable
          columns={["用户邀请码", "小组邀请码"]}
          labelHeader="维度"
          rows={[
            {
              label: "生成者",
              values: ["每个用户注册时自动生成一个", "创建小组时自动生成"],
            },
            {
              label: "作用",
              values: [
                "新用户注册时建立推广关系",
                "被邀请者加入指定的学习小组",
              ],
            },
            {
              label: "链接格式",
              values: ["通过 ref cookie 传递", "/g/[code]"],
            },
            {
              label: "触发点",
              values: ["新用户第一次注册", "任意用户（不论新老）点击后加入"],
            },
          ]}
        />
      </DocSection>

      <DocSection id="usage" title="如何使用它们">
        <p>
          你分享个人推广链接给想邀请的朋友——这是用户邀请码在起作用，对方注册后会成为你的推广关系。如果你想让好友加入你创建的某个具体小组，用小组的 <DocSlug>/g/[code]</DocSlug>{" "}
          链接——对方点开就能直接进入那个小组。两种邀请码互不冲突，可以同时使用。
        </p>
      </DocSection>
    </>
  );
}
