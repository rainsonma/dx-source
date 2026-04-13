import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocCompareTable } from "@/features/web/docs/primitives/doc-compare-table";
import { DocLink } from "@/features/web/docs/primitives/doc-link";
import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocSteps } from "@/features/web/docs/primitives/doc-steps";

export default function GroupMode() {
  return (
    <>
      <DocSection id="what" title="什么是小组共学">
        <p>
          小组共学是让一群人一起进入同一场学习对局的模式——它适合朋友、同学、学习搭子一起&ldquo;开黑&rdquo;。一局小组游戏由组主开局，之后所有在线组员会进入同一个房间同步开始。
        </p>
        <p>
          要玩小组模式，你必须先加入一个学习小组。如果你还没有小组，先去{" "}
          <DocLink href="/docs/groups/create-join">创建或加入一个小组</DocLink>
          。
        </p>
      </DocSection>

      <DocSection id="two-formats" title="两种开局方式">
        <DocCompareTable
          columns={["group_solo（个人排名）", "group_team（分组对战）"]}
          labelHeader="维度"
          rows={[
            {
              label: "参与方式",
              values: [
                "组内所有人各自闯同一游戏",
                "成员按子分组分队，分队之间互相对抗",
              ],
            },
            {
              label: "结算依据",
              values: ["个人分数排名", "子分组汇总分数 + 个人分数"],
            },
            {
              label: "适合人数",
              values: ["任意", "人多时更有意思"],
            },
          ]}
        />
      </DocSection>

      <DocSection id="how-to-start" title="组主如何开局">
        <DocSteps
          steps={[
            {
              title: "选游戏",
              desc: "组主在小组页选择这次要一起玩的游戏",
            },
            {
              title: "选模式",
              desc: "选 group_solo 个人排名，或 group_team 分组对战",
            },
            {
              title: "设置起始关卡",
              desc: "所有组员将从这一关开始",
            },
            {
              title: "开始游戏",
              desc: "组员进入房间后同步开始闯关",
            },
          ]}
        />
      </DocSection>

      <DocSection id="progress" title="关卡推进和结束">
        <p>
          进入游戏后，组主掌握节奏：当前关卡结束时，可以点击&ldquo;下一关&rdquo;把整个房间推进到下一关，也可以随时&ldquo;强制结束&rdquo;整局游戏。组员不能自行推进，这是为了保证所有人节奏一致。
        </p>
        <DocCallout variant="info" title="结算页展示">
          每一关结束后展示本关排名和个人得分；整局结束后还会展示累计的小组成绩。team 模式下会多一层子分组的汇总分数对比。
        </DocCallout>
      </DocSection>
    </>
  );
}
