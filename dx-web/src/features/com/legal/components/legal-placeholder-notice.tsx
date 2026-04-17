import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { PLACEHOLDERS } from "@/features/com/legal/constants";

type Props = { fields: (keyof typeof PLACEHOLDERS)[] };

export function LegalPlaceholderNotice({ fields }: Props) {
  return (
    <DocCallout variant="tip" title="法律信息占位">
      <p>
        本页以下 <code>{"{{...}}"}</code>{" "}
        标记的字段需在法律团队审阅后替换为正式内容，当前版本用于产品功能验证，<strong>非最终法律文本</strong>。
      </p>
      <ul className="mt-2 flex flex-wrap gap-2">
        {fields.map((k) => (
          <li
            key={k}
            className="rounded bg-amber-50 px-2 py-0.5 font-mono text-[12px] text-amber-700"
          >
            {PLACEHOLDERS[k]}
          </li>
        ))}
      </ul>
    </DocCallout>
  );
}
