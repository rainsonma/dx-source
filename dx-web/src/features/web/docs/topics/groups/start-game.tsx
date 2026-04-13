import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocSteps } from "@/features/web/docs/primitives/doc-steps";

export default function StartGame() {
  return (
    <>
      <DocSection id="who" title="谁能开局">
        <p>
          小组游戏只有组主可以开局——这是为了防止多人同时触发开局造成混乱。组主开局后，所有在线组员会收到提示并进入同一个游戏房间。不在线的成员不会被拉进来。
        </p>
      </DocSection>

      <DocSection id="flow" title="开局四步走">
        <DocSteps
          steps={[
            {
              title: "选游戏",
              desc: "从小组页面的&ldquo;开局&rdquo;入口选择一个游戏",
            },
            {
              title: "选模式",
              desc: "group_solo（组内个人排名）或 group_team（分组对战）",
            },
            {
              title: "设置起始关卡",
              desc: "所有组员从这一关开始",
            },
            {
              title: "开始游戏",
              desc: "在线组员进入房间后同步开始",
            },
          ]}
        />
      </DocSection>

      <DocSection id="next-level" title="推进到下一关">
        <p>
          当前关卡结束后，组主可以点击&ldquo;下一关&rdquo;让整个房间前进到下一关，所有组员同步切换。这种设计保证了节奏的一致性——组主相当于&ldquo;开火车的司机&rdquo;。
        </p>
      </DocSection>

      <DocSection id="force-end" title="强制结束">
        <p>
          如果因为某种原因需要提前结束整局游戏（例如卡壳、成员掉线太多、组主有事），组主可以随时点&ldquo;强制结束&rdquo;。强制结束后游戏立即收尾，系统按当前进度结算展示。
        </p>
      </DocSection>

      <DocSection id="results" title="结算与排名">
        <DocCallout variant="info" title="两层结算">
          每一关结束时展示本关的个人排名；整局结束时展示累计的成员排名。team 模式还会多一层子分组之间的汇总对比，让每个分队看到自己的整体表现。
        </DocCallout>
      </DocSection>
    </>
  );
}
