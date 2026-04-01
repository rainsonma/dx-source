package seeders

import (
	"fmt"
	"log"
	"math/rand"

	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
)

type UserSeeder struct{}

func (s *UserSeeder) Signature() string {
	return "UserSeeder"
}

func (s *UserSeeder) Run() error {
	namedPw, err := helpers.HashPassword("Password123")
	if err != nil {
		return fmt.Errorf("failed to hash named password: %w", err)
	}
	mockPw, err := helpers.HashPassword("Mock!@#Pass")
	if err != nil {
		return fmt.Errorf("failed to hash mock password: %w", err)
	}

	query := facades.Orm().Query()

	// ── Named users (rainson & june) ──────────────────────────────
	namedUsers := []struct {
		Username string
		Nickname string
		Grade    string
		Email    string
	}{
		{"rainson", "Rainson", "lifetime", "rainsonma@gmail.com"},
		{"june", "June", "lifetime", ""},
	}

	for _, u := range namedUsers {
		nickname := u.Nickname
		var email *string
		if u.Email != "" {
			email = &u.Email
		}

		var existing models.User
		if err := query.Where("username", u.Username).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.User{
				ID:         uuid.Must(uuid.NewV7()).String(),
				Username:   u.Username,
				Nickname:   &nickname,
				Grade:      u.Grade,
				Email:      email,
				Password:   namedPw,
				InviteCode: helpers.GenerateInviteCode(8),
				IsActive:   true,
			}); err != nil {
				return fmt.Errorf("failed to create user %s: %w", u.Username, err)
			}
		} else {
			if _, err := query.Where("username", u.Username).Update(&models.User{
				Nickname: &nickname,
				Grade:    u.Grade,
				Email:    email,
				Password: namedPw,
				IsActive: true,
			}); err != nil {
				return fmt.Errorf("failed to update user %s: %w", u.Username, err)
			}
		}
	}

	// ── 1200 mock users ──────────────────────────────────────────
	grades := []string{"month", "season", "year", "lifetime"}
	mockUsers := buildMockUsers()

	for _, m := range mockUsers {
		grade := grades[rand.Intn(len(grades))]
		nickname := m.Nickname

		var existing models.User
		if err := query.Where("username", m.Username).First(&existing); err != nil || existing.ID == "" {
			if err := query.Create(&models.User{
				ID:         uuid.Must(uuid.NewV7()).String(),
				Username:   m.Username,
				Nickname:   &nickname,
				Grade:      grade,
				Password:   mockPw,
				InviteCode: helpers.GenerateInviteCode(8),
				IsActive:   true,
				IsMock:     true,
			}); err != nil {
				return fmt.Errorf("failed to create mock user %s: %w", m.Username, err)
			}
		} else {
			if _, err := query.Where("username", m.Username).Update(&models.User{
				Nickname: &nickname,
				Grade:    grade,
				Password: mockPw,
				IsActive: true,
				IsMock:   true,
			}); err != nil {
				return fmt.Errorf("failed to update mock user %s: %w", m.Username, err)
			}
		}
	}

	log.Println("Seeded 1202 users (2 named + 1200 mock)")
	return nil
}

type mockUser struct {
	Username string
	Nickname string
}

// buildMockUsers returns 1200 users with realistic usernames and nicknames.
func buildMockUsers() []mockUser {
	firstNames := []string{
		"emma", "liam", "olivia", "noah", "ava", "ethan", "sophia", "mason",
		"isabella", "james", "mia", "logan", "charlotte", "benjamin", "amelia",
		"lucas", "harper", "henry", "evelyn", "alexander", "abigail", "jackson",
		"emily", "sebastian", "luna", "aiden", "chloe", "owen", "grace", "samuel",
		"ella", "jacob", "scarlett", "michael", "aria", "daniel", "lily", "matthew",
		"zoey", "carter", "riley", "jayden", "nora", "wyatt", "hannah", "jack",
		"hazel", "julian", "violet", "luke", "aurora", "gabriel", "savannah",
		"caleb", "audrey", "ryan", "brooklyn", "nathan", "penelope", "elijah",
		"stella", "isaac", "claire", "lincoln", "skylar", "joshua", "paisley",
		"andrew", "ellie", "connor", "anna", "hudson", "caroline", "adam",
		"genesis", "thomas", "madelyn", "leo", "aaliyah", "miles", "kennedy",
		"william", "kinsley", "david", "allison", "joseph", "maya", "john",
		"sarah", "dylan", "madeline", "landon", "naomi", "eli", "alice",
		"adrian", "hailey", "jaxon", "eva", "asher", "autumn", "christopher",
		"quinn", "charles", "nevaeh", "ezra", "piper", "maverick", "ruby",
		"josiah", "serenity", "colton", "willow", "cooper", "taylor", "ian",
		"emilia", "carson", "mackenzie", "axel", "isla", "declan", "brianna",
		"easton", "jordan", "damian", "alexa", "hunter", "peyton", "kayden",
		"layla", "robert", "kylie", "angel", "nicole", "dominic", "melanie",
		"tristan", "gianna", "max", "julia", "nolan", "olive", "tucker",
		"delilah", "kai", "ivy", "blake", "brielle", "jace", "valentina",
		"bentley", "diana", "beckett", "kara", "levi", "morgan", "tyler",
		"sydney", "gavin", "ariel", "brandon", "alexis", "chase", "maria",
		"travis", "vera", "marcus", "lila", "cole", "daisy", "grant",
		"summer", "derek", "paige", "finn", "keira", "roman", "elise",
		"jesse", "melody", "vincent", "sienna", "elliot", "camille", "brody",
		"reagan", "dustin", "vivian", "griffin", "fiona", "reid", "diana",
		"callum", "eden", "rowan", "sage", "felix", "margot", "jasper",
		"esme", "simon", "wren", "theo", "iris", "oscar", "nadia",
		"hugo", "lydia", "spencer", "clara", "malcolm", "josie",
		"bodhi", "june", "soren", "mila", "milo", "amber", "harris",
		"athena", "porter", "serena", "wade", "ivy", "brock", "elena",
	}

	nickFirsts := []string{
		"Emma", "Liam", "Olivia", "Noah", "Ava", "Ethan", "Sophia", "Mason",
		"Isabella", "James", "Mia", "Logan", "Charlotte", "Benjamin", "Amelia",
		"Lucas", "Harper", "Henry", "Evelyn", "Alex", "Abigail", "Jackson",
		"Emily", "Sebastian", "Luna", "Aiden", "Chloe", "Owen", "Grace", "Samuel",
		"Ella", "Jacob", "Scarlett", "Michael", "Aria", "Daniel", "Lily", "Matthew",
		"Zoey", "Carter", "Riley", "Jayden", "Nora", "Wyatt", "Hannah", "Jack",
		"Hazel", "Julian", "Violet", "Luke", "Aurora", "Gabriel", "Savannah",
		"Caleb", "Audrey", "Ryan", "Brooklyn", "Nathan", "Penelope", "Elijah",
		"Stella", "Isaac", "Claire", "Lincoln", "Skylar", "Joshua", "Paisley",
		"Andrew", "Ellie", "Connor", "Anna", "Hudson", "Caroline", "Adam",
		"Genesis", "Thomas", "Madelyn", "Leo", "Aaliyah", "Miles", "Kennedy",
		"William", "Kinsley", "David", "Allison", "Joseph", "Maya", "John",
		"Sarah", "Dylan", "Madeline", "Eli", "Alice", "Adrian", "Hailey",
		"Jaxon", "Eva", "Asher", "Autumn", "Christopher", "Quinn", "Charles",
		"Nevaeh", "Ezra", "Piper", "Maverick", "Ruby", "Josiah", "Serenity",
		"Colton", "Willow", "Cooper", "Taylor", "Ian", "Emilia", "Carson",
		"Mackenzie", "Axel", "Isla", "Declan", "Brianna", "Easton", "Jordan",
		"Damian", "Alexa", "Hunter", "Peyton", "Kayden", "Layla", "Robert",
		"Kylie", "Angel", "Nicole", "Dominic", "Melanie", "Tristan", "Gianna",
		"Max", "Julia", "Nolan", "Olive", "Tucker", "Delilah", "Kai", "Ivy",
		"Blake", "Brielle", "Jace", "Valentina", "Bentley", "Diana", "Beckett",
		"Kara", "Levi", "Morgan", "Tyler", "Sydney", "Gavin", "Ariel",
		"Brandon", "Alexis", "Chase", "Maria", "Travis", "Vera", "Marcus",
		"Lila", "Cole", "Daisy", "Grant", "Summer", "Derek", "Paige",
		"Finn", "Keira", "Roman", "Elise", "Jesse", "Melody", "Vincent",
		"Sienna", "Elliot", "Camille", "Brody", "Reagan", "Dustin", "Vivian",
		"Griffin", "Fiona", "Reid", "Athena", "Felix", "Margot", "Jasper",
		"Esme", "Simon", "Wren", "Theo", "Iris", "Oscar", "Nadia",
		"Hugo", "Lydia", "Spencer", "Clara", "Malcolm", "Josie",
		"Milo", "Amber", "Harris", "Serena", "Porter", "Elena",
	}

	cnBases := []string{
		"小明", "小红", "小华", "小丽", "小刚", "小芳", "小强", "小敏",
		"小龙", "小凤", "小虎", "小雪", "小云", "小雨", "小鱼", "小白",
		"大白", "大伟", "大力", "大勇", "大鹏", "大雄", "大海",
		"天天", "甜甜", "乐乐", "安安", "萌萌", "豆豆", "果果", "星星",
		"浩然", "子涵", "梓轩", "博文", "俊熙", "昊天", "逸飞", "明哲",
		"诗涵", "雨桐", "欣怡", "梦瑶", "紫萱", "若曦", "思琪", "雅文",
		"橘子", "柠檬", "草莓", "西瓜", "芒果", "葡萄", "樱桃", "桃子",
		"奶茶", "咖啡", "抹茶", "可乐", "布丁", "饼干", "糖果",
		"清风", "明月", "朝阳", "晨光", "暮雪", "青竹", "红枫",
		"学霸", "咸鱼", "追风", "锦鲤", "佛系",
		"龙腾", "虎啸", "凤舞", "鹤鸣", "鹰飞", "蝶舞",
		"阿杰", "阿明", "阿飞", "阿宝", "阿亮",
		"志远", "嘉诚", "鹏飞", "文杰", "建国",
		"美玲", "秀英", "淑芬", "雪梅", "春花",
		"书生", "墨客", "琴心", "画仙",
	}

	seen := make(map[string]bool)
	users := make([]mockUser, 0, 1200)

	for len(users) < 1200 {
		username := firstNames[rand.Intn(len(firstNames))]
		if rand.Intn(2) == 0 {
			username += firstNames[rand.Intn(len(firstNames))]
		}
		if seen[username] || username == "rainson" || username == "june" {
			continue
		}
		seen[username] = true

		nickname := nickFirsts[rand.Intn(len(nickFirsts))]
		users = append(users, mockUser{Username: username, Nickname: nickname})
	}

	// Randomly assign Chinese nicknames to 200 users.
	cnSuffixes := []string{"同学", "老师", "大王", "宝宝", "达人", "少年"}
	cnPrefixes := []string{"超级", "快乐", "阳光", "无敌", "元气", "暴躁"}
	cnNums := []string{"一", "二", "三", "五", "六", "七", "八", "九", "十", "百", "千", "万"}
	cnPatterns := []func(string) string{
		func(b string) string { return b },
		func(b string) string { return fmt.Sprintf("%s%d", b, rand.Intn(999)+1) },
		func(b string) string { return fmt.Sprintf("%s_%d", b, rand.Intn(99)+1) },
		func(b string) string { return b + cnNums[rand.Intn(len(cnNums))] },
		func(b string) string { return b + cnSuffixes[rand.Intn(len(cnSuffixes))] },
		func(b string) string { return cnPrefixes[rand.Intn(len(cnPrefixes))] + b },
		func(b string) string { return b + "_" + cnBases[rand.Intn(len(cnBases))] },
	}
	for _, idx := range rand.Perm(len(users))[:200] {
		base := cnBases[rand.Intn(len(cnBases))]
		users[idx].Nickname = cnPatterns[rand.Intn(len(cnPatterns))](base)
	}

	return users
}
