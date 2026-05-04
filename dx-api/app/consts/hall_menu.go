package consts

// HallMenuItem represents a single sidebar navigation item.
type HallMenuItem struct {
	Icon     string `json:"icon"`
	Label    string `json:"label"`
	Subtitle string `json:"subtitle"`
	Href     string `json:"href"`
}

// HallMenuSection groups related sidebar navigation items.
type HallMenuSection struct {
	Items []HallMenuItem `json:"items"`
}

// HallMenuSections returns the complete sidebar menu structure.
func HallMenuSections() []HallMenuSection {
	return []HallMenuSection{
		{Items: []HallMenuItem{
			{Icon: "LayoutDashboard", Label: "我的主页", Subtitle: "", Href: "/hall"},
			{Icon: "Gamepad2", Label: "学习课程", Subtitle: "选择一个游戏模式，边玩边学英语！", Href: "/hall/games"},
			{Icon: "Gamepad2", Label: "我的课程", Subtitle: "你玩过的所有学习课程", Href: "/hall/games/mine"},
			{Icon: "Users", Label: "学习群组", Subtitle: "浏览并加入学习群组，与小伙伴一起进步", Href: "/hall/groups"},
			{Icon: "Star", Label: "我的收藏", Subtitle: "收藏你喜欢的课程游戏和学习内容", Href: "/hall/favorites"},
			{Icon: "Trophy", Label: "排行榜单", Subtitle: "查看学习排名，与好友一起进步", Href: "/hall/leaderboard"},
		}},
		{Items: []HallMenuItem{
			{Icon: "Library", Label: "AI 词汇库", Subtitle: "管理你的个人词汇库，AI 生成或手动添加", Href: "/hall/ai-vocabs"},
			{Icon: "Sparkles", Label: "AI 随心学", Subtitle: "AI 驱动的个性化英语练习游戏", Href: "/hall/ai-custom"},
		}},
		{Items: []HallMenuItem{
			{Icon: "BookOpen", Label: "生词本", Subtitle: "记录你遇到的新单词和生词", Href: "/hall/unknown"},
			{Icon: "RotateCcw", Label: "复习本", Subtitle: "需要复习巩固的词汇和知识点", Href: "/hall/review"},
			{Icon: "CheckCircle2", Label: "已掌握", Subtitle: "你已经掌握的词汇和知识点", Href: "/hall/mastered"},
		}},
		{Items: []HallMenuItem{
			{Icon: "MessageCircle", Label: "斗学社", Subtitle: "分享学习心得，与学友互动交流", Href: "/hall/community"},
		}},
		{Items: []HallMenuItem{
			{Icon: "Bell", Label: "消息通知", Subtitle: "查看系统通知和公告", Href: "/hall/notices"},
			{Icon: "Medal", Label: "个人中心", Subtitle: "管理你的个人资料和账号信息", Href: "/hall/me"},
		}},
	}
}
