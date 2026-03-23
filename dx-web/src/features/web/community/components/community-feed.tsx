"use client";

import {
  Heart,
  MessageCircle,
  Share2,
  Bookmark,
  UserPlus,
  Send,
  Ban,
  ShieldAlert,
  Plus,
} from "lucide-react";
import { TabPill } from "@/components/in/tab-pill";

const feedTabs = [
  { label: "最新", active: true },
  { label: "热门", active: false },
  { label: "推荐", active: false },
  { label: "关注", active: false },
  { label: "收藏", active: false },
];

interface Comment {
  avatar: string;
  avatarBg: string;
  avatarColor: string;
  name: string;
  time: string;
  content: string;
  likes: number;
}

interface Post {
  avatar: string;
  avatarBg: string;
  avatarColor: string;
  name: string;
  badge: string;
  time: string;
  content: string;
  tags: string[];
  likes: number;
  comments: number;
  shares: number;
  liked: boolean;
  bookmarked: boolean;
  following: boolean;
  hasImage?: boolean;
  commentList: Comment[];
}

const posts: Post[] = [
  {
    avatar: "陈",
    avatarBg: "bg-teal-100",
    avatarColor: "text-teal-700",
    name: "陈学霸",
    badge: "Pro",
    time: "3小时前 · 英语学习",
    content:
      "今天终于把《新概念英语》第三册的 Lesson 25 啃完了！分享一下我的笔记和心得，希望对大家有帮助。重点是虚拟语气的用法，真的很容易搞混 大家有什么好的记忆方法吗？",
    tags: ["#英语学习", "#新概念英语", "#虚拟语气"],
    likes: 128,
    comments: 32,
    shares: 15,
    liked: false,
    bookmarked: false,
    following: false,
    commentList: [
      {
        avatar: "王",
        avatarBg: "bg-blue-100",
        avatarColor: "text-blue-700",
        name: "王小明",
        time: "2小时前",
        content:
          "虚拟语气确实很难！我的方法是把 if 从句分三种类型背例句，然后每天造一个句子练习，坚持一周就有感觉了",
        likes: 24,
      },
      {
        avatar: "李",
        avatarBg: "bg-purple-100",
        avatarColor: "text-purple-700",
        name: "李雨薇",
        time: "1小时前",
        content:
          "同在学第三册！Lesson 25 的被动语态也很绕，我整理了一份对比表格，要的话可以私信我~",
        likes: 18,
      },
    ],
  },
  {
    avatar: "张",
    avatarBg: "bg-amber-100",
    avatarColor: "text-amber-700",
    name: "张老师",
    badge: "",
    time: "5小时前 · 英语学习",
    content:
      "给大家分享一个高效背单词的方法：间隔重复法 + 场景联想。把每个新单词放到一个你熟悉的场景中，配合 Anki 的间隔重复提醒，记忆效果翻倍！下面是我整理的思维导图",
    tags: ["#背单词", "#学习方法"],
    likes: 256,
    comments: 48,
    shares: 34,
    liked: true,
    bookmarked: true,
    following: true,
    hasImage: true,
    commentList: [
      {
        avatar: "刘",
        avatarBg: "bg-green-100",
        avatarColor: "text-green-700",
        name: "刘大壮",
        time: "4小时前",
        content:
          "张老师的方法太实用了！我用间隔重复法一个月背了 800 个单词，比之前死记硬背效率高太多了",
        likes: 42,
      },
    ],
  },
];

function PostCard({ post }: { post: Post }) {
  return (
    <div className="flex flex-col gap-4 rounded-xl border border-border bg-card p-5">
      {/* Post header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div
            className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-full ${post.avatarBg}`}
          >
            <span className={`text-sm font-semibold ${post.avatarColor}`}>
              {post.avatar}
            </span>
          </div>
          <div className="flex flex-col gap-0.5">
            <div className="flex items-center gap-2">
              <span className="text-sm font-semibold text-foreground">
                {post.name}
              </span>
              {post.badge && (
                <span className="rounded bg-amber-100 px-1.5 py-0.5 text-[10px] font-semibold text-amber-700">
                  {post.badge}
                </span>
              )}
            </div>
            <span className="text-xs text-muted-foreground">{post.time}</span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {post.following ? (
            <span className="rounded-full bg-muted px-4 py-1.5 text-[13px] font-medium text-muted-foreground">
              已关注
            </span>
          ) : (
            <button
              type="button"
              className="flex items-center gap-1.5 rounded-full border border-teal-600 px-4 py-1.5 text-[13px] font-semibold text-teal-600"
            >
              <UserPlus className="h-3.5 w-3.5" />
              关注
            </button>
          )}
          <button
            type="button"
            className="hidden h-8 w-8 items-center justify-center rounded-full border border-border sm:flex"
          >
            <Ban className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
          <button
            type="button"
            className="hidden h-8 w-8 items-center justify-center rounded-full border border-border sm:flex"
          >
            <ShieldAlert className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
        </div>
      </div>

      {/* Post content */}
      <p className="text-sm leading-relaxed text-foreground">{post.content}</p>

      {post.hasImage && (
        <div className="h-[220px] w-full rounded-[10px] bg-muted" />
      )}

      {/* Tags */}
      <div className="flex flex-wrap gap-2">
        {post.tags.map((tag) => (
          <span
            key={tag}
            className="rounded-md bg-teal-50 px-2.5 py-1 text-xs font-medium text-teal-600"
          >
            {tag}
          </span>
        ))}
      </div>

      <div className="h-px w-full bg-border" />

      {/* Actions */}
      <div className="flex w-full justify-around">
        <button type="button" className="flex items-center gap-1.5 rounded-lg px-4 py-2">
          <Heart className={`h-[18px] w-[18px] ${post.liked ? "text-red-500" : "text-muted-foreground"}`} />
          <span className={`text-[13px] font-medium ${post.liked ? "text-red-500" : "text-muted-foreground"}`}>
            {post.likes}
          </span>
        </button>
        <button type="button" className="flex items-center gap-1.5 rounded-lg px-4 py-2">
          <MessageCircle className="h-[18px] w-[18px] text-muted-foreground" />
          <span className="text-[13px] font-medium text-muted-foreground">{post.comments}</span>
        </button>
        <button type="button" className="flex items-center gap-1.5 rounded-lg px-4 py-2">
          <Share2 className="h-[18px] w-[18px] text-muted-foreground" />
          <span className="text-[13px] font-medium text-muted-foreground">{post.shares}</span>
        </button>
        <button type="button" className="rounded-lg px-4 py-2">
          <Bookmark className={`h-[18px] w-[18px] ${post.bookmarked ? "text-teal-600" : "text-muted-foreground"}`} />
        </button>
      </div>

      {/* Comments */}
      <div className="flex flex-col gap-3.5 rounded-[10px] bg-muted p-4">
        {post.commentList.map((comment, ci) => (
          <div key={ci}>
            {ci > 0 && <div className="mb-3.5 h-px w-full bg-border" />}
            <div className="flex gap-2.5">
              <div
                className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full ${comment.avatarBg}`}
              >
                <span className={`text-xs font-semibold ${comment.avatarColor}`}>
                  {comment.avatar}
                </span>
              </div>
              <div className="flex flex-col gap-1.5">
                <div className="flex items-center gap-2">
                  <span className="text-[13px] font-semibold text-foreground">{comment.name}</span>
                  <span className="text-xs text-muted-foreground">{comment.time}</span>
                </div>
                <p className="text-[13px] leading-relaxed text-muted-foreground">{comment.content}</p>
                <div className="flex items-center gap-3">
                  <button type="button" className="flex items-center gap-1 text-xs text-muted-foreground">
                    <Heart className="h-3 w-3" />
                    {comment.likes}
                  </button>
                  <button type="button" className="text-xs text-muted-foreground">回复</button>
                </div>
              </div>
            </div>
          </div>
        ))}
        <span className="text-xs font-medium text-teal-600">
          查看全部 {post.comments} 条评论
        </span>
        <div className="flex items-center gap-2.5">
          <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-border">
            <span className="text-[10px] font-semibold text-muted-foreground">我</span>
          </div>
          <div className="flex h-9 flex-1 items-center rounded-full border border-border bg-card px-3.5">
            <span className="text-[13px] text-muted-foreground">写下你的评论...</span>
          </div>
          <button
            type="button"
            className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-teal-600"
          >
            <Send className="h-4 w-4 text-white" />
          </button>
        </div>
      </div>
    </div>
  );
}

export function CommunityFeed() {
  return (
    <>
      {/* Tab row */}
      <div className="flex w-full flex-col items-start justify-between gap-3 sm:flex-row sm:items-center">
        <div className="flex flex-wrap items-center gap-2">
          {feedTabs.map((tab) => (
            <TabPill key={tab.label} label={tab.label} active={tab.active} />
          ))}
        </div>
        <button
          type="button"
          className="flex items-center gap-2 rounded-[10px] bg-teal-600 px-5 py-2.5 text-sm font-semibold text-white hover:bg-teal-700"
        >
          <Plus className="h-4 w-4" />
          发帖
        </button>
      </div>

      {/* Main columns */}
      <div className="flex flex-col gap-6 lg:flex-row">
        {/* Feed column */}
        <div className="flex flex-1 flex-col gap-4">
          {posts.map((post, i) => (
            <PostCard key={i} post={post} />
          ))}
        </div>

        {/* Right column */}
        <div className="flex w-full shrink-0 flex-col gap-4 lg:w-[300px]">
          <TrendingTopicsPanel />
          <SuggestedUsersPanel />
        </div>
      </div>
    </>
  );
}

const trendingTopics = [
  { rank: 1, name: "#四六级备考", count: "1.2k 讨论" },
  { rank: 2, name: "#雅思写作技巧", count: "856 讨论" },
  { rank: 3, name: "#口语打卡挑战", count: "643 讨论" },
  { rank: 4, name: "#每日一句", count: "521 讨论" },
];

function TrendingTopicsPanel() {
  return (
    <div className="flex flex-col gap-4 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-2">
        <span className="text-base font-bold text-foreground">热门话题</span>
      </div>
      {trendingTopics.map((topic) => (
        <div key={topic.rank} className="flex items-center gap-3">
          <span
            className={`text-sm font-bold ${topic.rank <= 3 ? "text-teal-600" : "text-muted-foreground"}`}
          >
            {topic.rank}
          </span>
          <div className="flex flex-col gap-0.5">
            <span className="text-sm font-semibold text-foreground">{topic.name}</span>
            <span className="text-xs text-muted-foreground">{topic.count}</span>
          </div>
        </div>
      ))}
    </div>
  );
}

const suggestedUsers = [
  { name: "李老师", desc: "雅思教学 10 年经验", avatar: "李", avatarBg: "bg-teal-100", avatarColor: "text-teal-700" },
  { name: "单词小王子", desc: "连续打卡 365 天", avatar: "单", avatarBg: "bg-amber-100", avatarColor: "text-amber-700" },
  { name: "英语达人Amy", desc: "口语教练 · 2.3k 粉丝", avatar: "A", avatarBg: "bg-purple-100", avatarColor: "text-purple-700" },
];

function SuggestedUsersPanel() {
  return (
    <div className="flex flex-col gap-4 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-2">
        <UserPlus className="h-[18px] w-[18px] text-teal-600" />
        <span className="text-base font-bold text-foreground">推荐关注</span>
      </div>
      {suggestedUsers.map((user) => (
        <div key={user.name} className="flex items-center gap-3">
          <div
            className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full ${user.avatarBg}`}
          >
            <span className={`text-sm font-semibold ${user.avatarColor}`}>{user.avatar}</span>
          </div>
          <div className="flex flex-1 flex-col gap-0.5">
            <span className="text-sm font-semibold text-foreground">{user.name}</span>
            <span className="text-xs text-muted-foreground">{user.desc}</span>
          </div>
          <button
            type="button"
            className="rounded-lg bg-teal-600 px-3 py-1.5 text-xs font-semibold text-white"
          >
            关注
          </button>
        </div>
      ))}
    </div>
  );
}
