import { Eye, Headphones, Layers, Mic, PenLine, Puzzle, Sparkles, Swords } from "lucide-react";
import { DocCompareTable } from "@/features/web/docs/primitives/doc-compare-table";
import { DocFeatureGrid } from "@/features/web/docs/primitives/doc-feature-grid";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function GameTypes() {
  return (
    <>
      <DocSection id="four-games" title="四种游戏类型">
        <p>
          斗学的所有游戏都属于以下四种玩法之一，分别训练不同的语言能力：
        </p>
        <DocFeatureGrid
          columns={2}
          items={[
            {
              icon: Layers,
              iconColor: "text-teal-600",
              title: "连词成句",
              desc: "把打乱顺序的单词重新拼接成一个正确的句子，训练语序和句法。",
            },
            {
              icon: Swords,
              iconColor: "text-rose-600",
              title: "词汇对轰",
              desc: "快速识别单词的中英文对应关系，考反应速度和熟练度。",
            },
            {
              icon: Puzzle,
              iconColor: "text-blue-600",
              title: "词汇配对",
              desc: "成对连线匹配单词与释义，节奏舒缓但需要准确。",
            },
            {
              icon: Sparkles,
              iconColor: "text-amber-600",
              title: "词汇消消乐",
              desc: "消除匹配的单词对，越玩越快，越记越牢。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="difficulty" title="三个难度等级">
        <p>
          每个游戏都可以在三个难度中选择一个——难度决定内容类型和复杂度：
        </p>
        <DocKeyValue
          items={[
            {
              key: "beginner（初级）",
              value: "所有内容类型都可选（单词 / 组合 / 短语 / 句子）",
              note: "适合新手入门",
            },
            {
              key: "intermediate（中级）",
              value: "仅组合、短语、句子",
              note: "去掉了基础单词",
            },
            {
              key: "advanced（高级）",
              value: "仅完整句子",
              note: "最具挑战",
            },
          ]}
        />
      </DocSection>

      <DocSection id="patterns" title="四种学习模式（听说读写）">
        <p>
          同一道题在不同&ldquo;学习模式&rdquo;下考察的能力完全不同——听说读写对应的是语言的四种基本技能：
        </p>
        <DocFeatureGrid
          columns={4}
          items={[
            {
              icon: Headphones,
              iconColor: "text-teal-600",
              title: "听 (listen)",
              desc: "播放音频，考察听力理解",
            },
            {
              icon: Mic,
              iconColor: "text-rose-600",
              title: "说 (speak)",
              desc: "跟读发音，考察口语输出",
            },
            {
              icon: Eye,
              iconColor: "text-blue-600",
              title: "读 (read)",
              desc: "阅读文字，考察理解速度",
            },
            {
              icon: PenLine,
              iconColor: "text-amber-600",
              title: "写 (write)",
              desc: "默认模式，看题作答，考察综合能力",
            },
          ]}
        />
      </DocSection>

      <DocSection id="matrix" title="难度 × 学习模式矩阵">
        <p>
          三个难度和四种学习模式可以任意搭配，形成 12 种组合。例如&ldquo;初级 + 写&rdquo;是最平易近人的入门组合，&ldquo;高级 + 听&rdquo;则相当有挑战。下表展示了所有组合都是可用的：
        </p>
        <DocCompareTable
          columns={["听", "说", "读", "写"]}
          labelHeader="难度"
          rows={[
            {
              label: "beginner",
              values: [true, true, true, true],
            },
            {
              label: "intermediate",
              values: [true, true, true, true],
            },
            {
              label: "advanced",
              values: [true, true, true, true],
            },
          ]}
        />
      </DocSection>
    </>
  );
}
