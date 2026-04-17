package consts

import (
	"math/rand/v2"
	"time"
)

// Greeting is a time-banded greeting for the hall dashboard.
type Greeting struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

type greetingBand struct {
	title     string
	subtitles []string
}

var (
	shanghaiLocation *time.Location
	greetingBands    []greetingBand
)

func init() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic("consts: failed to load Asia/Shanghai timezone: " + err.Error())
	}
	shanghaiLocation = loc

	greetingBands = []greetingBand{
		// 0 — morning: 05–10
		{
			title: "早上好 👋",
			subtitles: []string{
				"继续你的学习之旅，今天也要加油！",
				"新的一天，一起来背几个单词吧！",
				"早起的鸟儿有虫吃，冲呀！",
				"今天也要笑着开始学习哦～",
				"愿你的一天充满阳光和单词",
			},
		},
		// 1 — noon: 11–12
		{
			title: "中午好 🍚",
			subtitles: []string{
				"吃饭前先来几道题热身吧！",
				"午饭后，刷两道 quiz 如何？",
				"中午能量满满，继续冲刺！",
				"午休时间，来场英文小游戏吧",
				"一顿好饭配一页单词，完美！",
			},
		},
		// 2 — afternoon: 13–17
		{
			title: "下午好 ☕",
			subtitles: []string{
				"一杯咖啡配英语，下午更带劲",
				"一起消灭那些顽固的生词吧！",
				"午后微困？来段英语提提神！",
				"坚持一下，今天的目标不远了",
				"让英语给你的下午续点航",
			},
		},
		// 3 — evening: 18–23 and 0–4
		{
			title: "晚上好 🌙",
			subtitles: []string{
				"结束今天前，再多学一点点",
				"夜深人静，正适合练听力",
				"月亮不睡你也别睡，单词等你",
				"睡前温习，记忆更牢哦",
				"今日份英语打卡，完成！",
			},
		},
	}
}

// bandFor returns the band index (0=morning, 1=noon, 2=afternoon, 3=evening)
// for the given hour (0–23). Out-of-range hours fall into evening.
func bandFor(hour int) int {
	switch {
	case hour >= 5 && hour <= 10:
		return 0
	case hour >= 11 && hour <= 12:
		return 1
	case hour >= 13 && hour <= 17:
		return 2
	default:
		return 3
	}
}

// PickGreeting returns a Greeting whose Title matches the hour of t
// (interpreted in Asia/Shanghai) and whose Subtitle is a random entry
// from the band's pool. Uses math/rand/v2 top-level rand.IntN for
// subtitle selection — same pattern as services/api/mock_user_service.go.
func PickGreeting(t time.Time) Greeting {
	hour := t.In(shanghaiLocation).Hour()
	band := greetingBands[bandFor(hour)]
	return Greeting{
		Title:    band.title,
		Subtitle: band.subtitles[rand.IntN(len(band.subtitles))],
	}
}
