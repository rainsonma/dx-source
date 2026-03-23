import { TestimonialCard } from "@/features/web/auth/components/testimonial-card";

const testimonials = [
  {
    name: "Emily",
    tier: "月度会员",
    quote:
      "每天用斗学练口语，感觉像在玩游戏一样轻松，不知不觉词汇量翻了一倍，四级一次就过了！",
  },
  {
    name: "Kevin",
    tier: "年度会员",
    quote:
      "通勤路上用斗学刷单词，碎片时间也能高效学习，半年下来英语会议再也不怕了。",
  },
  {
    name: "小林妈妈",
    tier: "季度会员",
    quote:
      "孩子自从用了斗学，每天主动要求学英语，成绩从班级中游直接冲到前五，太惊喜了！",
  },
  {
    name: "Alex",
    tier: "终身会员",
    quote:
      "把考研真题导入斗学自定义刷题，游戏化复习效率超高，每天坚持2小时，英语顺利上岸！",
  },
  {
    name: "Sarah",
    tier: "月度会员",
    quote:
      "斗学的AI智能陪练太强了，像有一个随时在线的外教，口语进步特别明显，推荐给所有朋友了。",
  },
  {
    name: "David",
    tier: "终身会员",
    quote:
      "买了终身会员完全不后悔，课程持续更新，游戏越来越丰富，性价比真的太高了！",
  },
  {
    name: "Jessica",
    tier: "季度会员",
    quote:
      "斗学的听力闯关模式太有意思了，每天通勤练30分钟，现在看美剧基本不用字幕了！",
  },
  {
    name: "小陈",
    tier: "年度会员",
    quote:
      "词汇对战太上头了，和朋友一起PK背单词，不知不觉就记住了三千多个生词，根本停不下来。",
  },
  {
    name: "王老师",
    tier: "终身会员",
    quote:
      "作为英语老师，我推荐全班同学用斗学课后练习，班级平均分提了15分，家长都来感谢我。",
  },
  {
    name: "Lucy",
    tier: "月度会员",
    quote:
      "听说读写模式让我打字速度和拼写准确率都提升了，现在写英文邮件再也不用查单词了。",
  },
  {
    name: "张伟",
    tier: "年度会员",
    quote:
      "学习小组功能太赞了，和组员互相督促打卡，坚持了三个月，雅思从5.5提到了7分！",
  },
  {
    name: "Mike",
    tier: "季度会员",
    quote:
      "高级音效发音功能帮我纠正了好多发音问题，现在同事都说我口语听起来像在国外待过。",
  },
  {
    name: "赵哥",
    tier: "终身会员",
    quote:
      "推广佣金真的能赚到钱，分享给身边的朋友，每个月多了一笔零花钱，学英语还能赚钱太棒了。",
  },
  {
    name: "Amy",
    tier: "季度会员",
    quote:
      "单词消消乐是我家孩子最爱的模式，边玩边学，每天主动要求玩半小时，比任何辅导班都管用。",
  },
  {
    name: "李明",
    tier: "年度会员",
    quote:
      "生词本和错题本功能特别实用，薄弱点一目了然，针对性复习效率高了好几倍，强烈推荐！",
  },
];

export function TestimonialsGrid() {
  return (
    <div className="grid w-full grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3">
      {testimonials.map((t) => (
        <TestimonialCard
          key={t.name}
          name={t.name}
          tier={t.tier}
          quote={t.quote}
        />
      ))}
    </div>
  );
}
