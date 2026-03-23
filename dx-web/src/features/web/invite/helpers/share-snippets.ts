/** Platform-specific share snippet templates for the invite system */

export type ShareTab = {
  key: string;
  label: string;
  snippets: string[];
};

/** Default share tabs with 5 snippets each. Use {link} as placeholder for invite URL. */
export const SHARE_TABS: ShareTab[] = [
  {
    key: "wechat-moments",
    label: "微信朋友圈",
    snippets: [
      "发现一个超棒的英语学习平台——斗学！游戏化学习方式，让背单词变得超有趣\uD83D\uDD25 拼写竞速、词汇对战、听力闯关...边玩边学，效率翻倍！用我的邀请链接注册，一起来斗学吧\uD83D\uDC47\n{link}",
      "还在为英语发愁？试试「斗学」吧！我已经用了一段时间，词汇量提升明显\uD83D\uDCAA 通过游戏闯关的方式学英语，完全不枯燥。现在通过我的链接注册，我们都能获得佣金奖励哦~\n{link}",
      "朋友圈安利一个宝藏英语学习App——斗学 douxue.cc \uD83C\uDFAF 把背单词做成了游戏对战模式，有拼写竞速、词汇PK、听力挑战等。每天花15分钟，进步看得见！戳链接注册\uD83D\uDC47\n{link}",
      "自从用了斗学，背单词居然上瘾了\uD83D\uDE02 游戏化的学习模式太对味了，拼写竞速打到停不下来！强烈推荐给想提升英语的朋友们\u2728 链接放这儿，快来一起斗\uD83D\uDC47\n{link}",
      "英语学习也可以很好玩\uD83C\uDFAE 斗学把背单词变成了闯关游戏，越玩越想学！已经坚持了好几周，词汇量蹭蹭涨\uD83D\uDCC8 分享给同样想学英语的你\uD83D\uDC47\n{link}",
    ],
  },
  {
    key: "message",
    label: "消息评论",
    snippets: [
      "推荐你一个学英语的App「斗学」，游戏化背单词超有趣，我一直在用！\n{link}",
      "这个英语学习平台你试过吗？斗学，边玩游戏边背单词，效果特别好！\n{link}",
      "最近在用斗学学英语，拼写竞速模式太上头了哈哈，分享给你试试~\n{link}",
      "想提升英语的话推荐斗学，游戏闯关的方式学单词，比传统背单词有意思多了\n{link}",
      "给你安利一个宝藏App——斗学！学英语跟玩游戏一样，每天15分钟就有进步\n{link}",
    ],
  },
  {
    key: "weibo",
    label: "微博",
    snippets: [
      "#英语学习# #斗学# 发现一个超有趣的英语学习平台！斗学把背单词变成了游戏对战，拼写竞速、词汇PK、听力挑战，边玩边学效率翻倍\uD83D\uDD25 推荐给所有想学英语的小伙伴，戳链接注册\uD83D\uDC47\n{link}",
      "#学英语# #每日打卡# 今日份学习记录\u270D\uFE0F 在斗学上完成了拼写竞速挑战！游戏化学习模式真的让背单词不再枯燥\uD83D\uDCAF 想一起来的朋友看这里\uD83D\uDC47\n{link}",
      "#英语干货分享# 安利一个宝藏英语学习App「斗学」\uD83C\uDF1F 不同于传统背单词软件，它把学习做成了游戏闯关模式，有排行榜、有对战、有连击奖励！学英语终于不痛苦了\uD83D\uDE0D\n{link}",
      "#背单词# #斗学# 用斗学学英语第30天\uD83D\uDCC5 词汇量从3000涨到了4500+，游戏闯关模式让我每天都想打卡\uD83C\uDFAE 强烈推荐给想提升英语的宝子们\uD83D\uDC47\n{link}",
      "#英语提升# 谁说学英语一定要苦哈哈的？斗学用游戏的方式教你背单词\uD83C\uDFAF 拼写竞速超刺激，听力闯关也很有意思！分享给想学英语又怕无聊的你\u2728\n{link}",
    ],
  },
  {
    key: "xiaohongshu",
    label: "小红书",
    snippets: [
      "\uD83D\uDCD6 英语学习App测评｜斗学\n\u2B50\u2B50\u2B50\u2B50\u2B50\n\u2705 游戏化背单词，拼写竞速超上头\n\u2705 词汇对战模式，和好友PK更有动力\n\u2705 听力闯关，沉浸式提升\n每天15分钟，词汇量蹭蹭涨\uD83D\uDCC8\n链接放评论区，快来一起学\uD83D\uDC47\n{link}",
      "\uD83C\uDFAE 学英语也能上瘾？！\n用了斗学一个月的真实感受：\n\uD83D\uDD25 背单词终于不枯燥了\n\uD83D\uDCAA 词汇量提升超明显\n\uD83C\uDFC6 排行榜机制很有激励感\n真心推荐给想学英语的姐妹们！\n{link}",
      "\u2728 宝藏英语学习App分享\n还在用传统方法背单词吗？\n试试「斗学」吧！\n\uD83C\uDFAF 游戏闯关式学习\n\uD83D\uDCCA 实时记录学习进度\n\uD83E\uDD1D 邀请好友一起学有奖励\n学英语也可以很快乐\uD83D\uDE0D\n{link}",
      "\uD83D\uDCA1 自学英语一定要知道的App\n「斗学」我已经用了好几周了！\n\uD83D\uDCDD 拼写竞速：练拼写超有效\n\uD83D\uDD0A 听力挑战：边听边学\n\uD83C\uDFC5 连击奖励：越玩越有成就感\n适合所有水平的英语学习者\uD83D\uDC4D\n{link}",
      "\uD83D\uDE80 英语提升秘诀｜游戏化学习\n分享我最近在用的英语学习平台「斗学」\n\u2764\uFE0F 最喜欢的功能：\n1\uFE0F\u20E3 拼写竞速模式\n2\uFE0F\u20E3 词汇量测试\n3\uFE0F\u20E3 学习数据统计\n零碎时间就能学，强烈推荐\uD83D\uDD25\n{link}",
    ],
  },
];
