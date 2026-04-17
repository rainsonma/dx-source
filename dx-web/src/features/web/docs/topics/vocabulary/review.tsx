import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocLink } from "@/features/web/docs/primitives/doc-link";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function Review() {
  return (
    <>
      <DocSection id="spaced-repetition" title="什么是间隔重复">
        <p>
          间隔重复是一种基于记忆科学的学习方法：每复习一次同一条内容，下次再复习的时间间隔就拉长一点。刚学会时需要 1 天后复习，记住一段时间后 3 天、7 天、14 天……到最后隔 3 个月才需要再看一眼。
        </p>
        <p>
          这样做的好处是：你把学习精力花在&ldquo;马上要忘&rdquo;的东西上，而不是反复复习已经熟的东西。长期来看，学到的每一条都能用最少的时间成本留下来。
        </p>
      </DocSection>

      <DocSection id="intervals" title="斗学的间隔表">
        <p>斗学采用的是一个经过调优的六级间隔：</p>
        <DocKeyValue
          items={[
            { key: "第 0 次复习后", value: "1 天后再复习" },
            { key: "第 1 次复习后", value: "3 天后" },
            { key: "第 2 次复习后", value: "7 天后" },
            { key: "第 3 次复习后", value: "14 天后" },
            { key: "第 4 次复习后", value: "30 天后" },
            { key: "第 5 次及以后", value: "90 天后", note: "最长间隔，不再拉长" },
          ]}
        />
      </DocSection>

      <DocSection id="three-states" title="复习本的三种状态">
        <p>
          打开复习本页面，你会看到三组内容：&ldquo;待复习&rdquo;是今天应该复习但还没做的；&ldquo;逾期&rdquo;是应该复习但超过了预定日期的；&ldquo;今日已复习&rdquo;是今天已经完成的。每天登录先看这三组，逾期的优先处理。
        </p>
      </DocSection>

      <DocSection id="flow" title="复习后发生什么">
        <p>
          完成一次复习后，系统按当前间隔把这条词推到下一档。如果你连续正确复习几次，这条词最终会被视为&ldquo;已掌握&rdquo;，不再出现在复习列表里——详情见{" "}
          <DocLink href="/wiki/vocabulary/mastered">已掌握</DocLink>
          。
        </p>
      </DocSection>
    </>
  );
}
