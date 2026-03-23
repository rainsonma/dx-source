export const IMAGE_ROLES = {
  ADM_USER_AVATAR: "adm-user-avatar",
  USER_AVATAR: "user-avatar",
  CATEGORY_COVER: "category-cover",
  TEMPLATE_COVER: "template-cover",
  GAME_COVER: "game-cover",
  PRESS_COVER: "press-cover",
  GAME_GROUP_COVER: "game-group-cover",
  POST_IMAGE: "post-image",
} as const;

export type ImageRole = (typeof IMAGE_ROLES)[keyof typeof IMAGE_ROLES];

export const IMAGE_ROLE_LABELS: Record<ImageRole, string> = {
  "adm-user-avatar": "管理员头像",
  "user-avatar": "用户头像",
  "category-cover": "分类封面",
  "template-cover": "模板封面",
  "game-cover": "游戏封面",
  "press-cover": "出版社封面",
  "game-group-cover": "游戏组封面",
  "post-image": "帖子图片",
};
