import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function Notices() {
  return (
    <>
      <DocSection id="list" title="通知中心">
        <p>
          通知中心集中展示斗学发给你的所有系统通知——包括产品更新、活动推送、重要变更。它的入口在侧边栏&ldquo;通知&rdquo;位置，进入后按时间倒序列出所有通知。
        </p>
      </DocSection>

      <DocSection id="badge" title="未读角标">
        <p>
          如果有新通知，侧边栏的通知入口会出现一个红点——它是未读提示，不显示数字，只提示&ldquo;你有未读的通知&rdquo;。进入通知中心后，红点不会立即消失，需要你点开具体通知。
        </p>
      </DocSection>

      <DocSection id="mark-read" title="标记已读">
        <p>
          点开任何一条通知会自动把它标记为已读。所有通知都已读之后，侧边栏的红点会在下次刷新时消失。已读的通知不会自动删除，还会留在通知列表里供你回看。
        </p>
      </DocSection>
    </>
  );
}
