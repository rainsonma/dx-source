export type PostAuthor = {
  id: string;
  nickname: string;
  avatar_url: string | null;
};

export type Post = {
  id: string;
  content: string;
  image_url: string | null;
  tags: string[];
  like_count: number;
  comment_count: number;
  is_liked: boolean;
  is_bookmarked: boolean;
  author: PostAuthor;
  created_at: string;
};

export type Comment = {
  id: string;
  content: string;
  author: PostAuthor;
  parent_id: string | null;
  created_at: string;
};

export type CommentWithReplies = {
  comment: Comment;
  replies: Comment[];
};

export type FeedTab = "latest" | "hot" | "following" | "bookmarks";
