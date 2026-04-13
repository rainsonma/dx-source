import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function PostsComments() {
  return (
    <>
      <DocSection id="post" title="发一条帖子">
        <p>
          斗学社（社区入口&ldquo;community&rdquo;）是一个轻量的社区，主要用来分享学习心得、提问、交流。进入斗学社后，右上角有&ldquo;发帖&rdquo;按钮，点击打开发帖窗口，填写正文（可选图片和标签）即可发送。
        </p>
      </DocSection>

      <DocSection id="post-rules" title="帖子的长度和格式规则">
        <DocKeyValue
          items={[
            {
              key: "正文长度",
              value: "最多 2000 字",
              note: "足够写一段完整的心得",
            },
            {
              key: "图片",
              value: "可选，单张",
              note: "上传后与帖子绑定",
            },
            {
              key: "标签",
              value: "最多 5 个",
              note: "每个标签最多 20 字",
            },
          ]}
        />
      </DocSection>

      <DocSection id="comments" title="评论">
        <p>
          每篇帖子下方都可以评论。评论的正文长度限制是 500 字，比帖子短。一条帖子下的评论按时间顺序展示。
        </p>
        <DocCallout variant="info" title="不支持嵌套回复">
          斗学社的评论是&ldquo;扁平&rdquo;的——你可以回复帖子，但不能回复某条评论本身。这个设计是为了保持讨论线性、好读，避免评论楼层过深。
        </DocCallout>
      </DocSection>

      <DocSection id="edit-delete" title="编辑和删除">
        <p>
          自己发的帖子和评论可以随时编辑或删除。删除是软删除——其他用户不会再看到这条内容，但系统会保留数据记录，这是为了可能的审核和数据恢复需要。
        </p>
      </DocSection>
    </>
  );
}
