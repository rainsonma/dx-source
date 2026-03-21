package constants

// Image role values identify what an uploaded image is used for.
const (
	ImageRoleAdmUserAvatar  = "adm-user-avatar"
	ImageRoleUserAvatar     = "user-avatar"
	ImageRoleCategoryCover  = "category-cover"
	ImageRoleTemplateCover  = "template-cover"
	ImageRoleGameCover      = "game-cover"
	ImageRolePressCover     = "press-cover"
	ImageRoleGameGroupCover = "game-group-cover"
	ImageRolePostImage      = "post-image"
)

// ImageRoleLabels maps each image role to its Chinese label.
var ImageRoleLabels = map[string]string{
	ImageRoleAdmUserAvatar:  "管理员头像",
	ImageRoleUserAvatar:     "用户头像",
	ImageRoleCategoryCover:  "分类封面",
	ImageRoleTemplateCover:  "模板封面",
	ImageRoleGameCover:      "游戏封面",
	ImageRolePressCover:     "出版社封面",
	ImageRoleGameGroupCover: "游戏组封面",
	ImageRolePostImage:      "帖子图片",
}
