import { AlertCircle, Bug, MessageSquare, MoreHorizontal, Palette } from "lucide-react";
import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocFeatureGrid } from "@/features/web/docs/primitives/doc-feature-grid";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function Feedback() {
  return (
    <>
      <DocSection id="entry" title="在哪里反馈">
        <p>
          反馈入口通常在个人中心或账户设置里。我们期待你随时告诉我们产品哪里好、哪里不好、希望看到什么——这类输入对斗学下一步做什么有直接影响。
        </p>
      </DocSection>

      <DocSection id="types" title="五种反馈类型">
        <p>提交反馈时请先选一个类型，方便我们分流处理：</p>
        <DocFeatureGrid
          columns={3}
          items={[
            {
              icon: MessageSquare,
              iconColor: "text-teal-600",
              title: "功能建议",
              desc: "你希望斗学能做什么现在还没做的事。",
            },
            {
              icon: AlertCircle,
              iconColor: "text-amber-600",
              title: "内容纠错",
              desc: "发现题目、翻译、词条有误。",
            },
            {
              icon: Palette,
              iconColor: "text-blue-600",
              title: "界面体验",
              desc: "用得不顺手、布局有问题、交互别扭。",
            },
            {
              icon: Bug,
              iconColor: "text-rose-600",
              title: "Bug 报告",
              desc: "系统出错、崩溃、页面空白、数据丢失。",
            },
            {
              icon: MoreHorizontal,
              iconColor: "text-slate-600",
              title: "其它",
              desc: "不属于上面任何一类但你想说的话。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="contact" title="联系方式（可选）">
        <p>
          反馈表单里有一个&ldquo;联系方式&rdquo;字段，你可以留邮箱或其它联系方式。留了我们才能在需要追问时找到你；不留也可以提交，只是你不会收到针对你的回复。
        </p>
      </DocSection>

      <DocSection id="what-happens" title="我们会处理">
        <DocCallout variant="info" title="每一条都会认真看">
          我们会认真看每一条反馈——无论是夸还是骂。但受限于人力，不一定每一条都能单独回复。如果你的问题很紧急且留了联系方式，我们会优先处理。
        </DocCallout>
      </DocSection>
    </>
  );
}
