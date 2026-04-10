package seeders

import (
	"fmt"
	"log"
	"math/rand"

	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// UserExtraSeeder is a one-shot throwaway seeder that appends 10,800
// mock users on top of whatever is already in the users table,
// without touching existing rows. Delete this file after running
// it once via: go run . artisan db:seed --seeder=UserExtraSeeder
type UserExtraSeeder struct{}

func (s *UserExtraSeeder) Signature() string {
	return "UserExtraSeeder"
}

// extraMockUser is the in-memory representation of a generated user
// before it's turned into a models.User. Named distinctly from
// user_seeder.go's private mockUser type to avoid a package-level
// type collision.
type extraMockUser struct {
	Username string
	Nickname string
}

func (s *UserExtraSeeder) Run() error {
	const (
		targetCount  = 10800
		chineseCount = 5000
		englishCount = targetCount - chineseCount
		chunkSize    = 500
	)

	// Phase 1 — hash the shared mock password once.
	mockPw, err := helpers.HashPassword("Mock!@#Pass")
	if err != nil {
		return fmt.Errorf("failed to hash mock password: %w", err)
	}

	// Phase 2 — load all existing usernames so the generator can
	// avoid colliding with the 1,202 users already in DB.
	seen, err := loadExistingUsernames()
	if err != nil {
		return fmt.Errorf("failed to load existing usernames: %w", err)
	}
	// Defensive fallbacks in case the DB is empty (fresh seed).
	seen["rainson"] = true
	seen["june"] = true

	// Phase 3 — generate targetCount unique users with English +
	// Chinese nicknames randomly interleaved.
	generated := buildExtraMockUsers(seen)

	// Phase 4 — assemble models.User structs.
	grades := []string{"month", "season", "year", "lifetime"}
	built := make([]models.User, 0, targetCount)
	for _, m := range generated {
		grade := grades[rand.Intn(len(grades))]
		nickname := m.Nickname
		built = append(built, models.User{
			ID:         uuid.Must(uuid.NewV7()).String(),
			Username:   m.Username,
			Nickname:   &nickname,
			Grade:      grade,
			Password:   mockPw,
			InviteCode: helpers.GenerateInviteCode(8),
			IsActive:   true,
			IsMock:     true,
		})
	}

	// Phase 5 — chunked batch insert inside one transaction.
	// All-or-nothing: if any chunk fails, everything rolls back.
	err = facades.Orm().Transaction(func(tx orm.Query) error {
		for i := 0; i < len(built); i += chunkSize {
			end := min(i+chunkSize, len(built))
			chunk := built[i:end]
			if err := tx.Create(&chunk); err != nil {
				return fmt.Errorf(
					"failed to insert chunk %d (rows %d-%d): %w",
					i/chunkSize, i, end-1, err,
				)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to seed extra mock users: %w", err)
	}

	log.Printf(
		"Seeded %d extra mock users (%d Chinese + %d English)",
		targetCount, chineseCount, englishCount,
	)
	return nil
}

// loadExistingUsernames returns a set of every username currently in
// the users table, keyed for O(1) collision checks.
func loadExistingUsernames() (map[string]bool, error) {
	type row struct {
		Username string `gorm:"column:username"`
	}
	var rows []row
	if err := facades.Orm().Query().
		Model(&models.User{}).
		Select("username").
		Get(&rows); err != nil {
		return nil, fmt.Errorf("failed to query existing usernames: %w", err)
	}
	seen := make(map[string]bool, len(rows)+2)
	for _, r := range rows {
		seen[r.Username] = true
	}
	return seen, nil
}

// buildExtraMockUsers generates targetCount unique mock users whose
// usernames don't collide with anything in `seen`, then overwrites
// chineseCount random indices with Chinese nicknames. The `seen`
// map is mutated as usernames are claimed.
func buildExtraMockUsers(seen map[string]bool) []extraMockUser {
	const (
		targetCount  = 10800
		chineseCount = 5000
	)

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

	users := make([]extraMockUser, 0, targetCount)

	// Phase A — generate unique English-nickname users.
	for len(users) < targetCount {
		username := firstNames[rand.Intn(len(firstNames))]
		if rand.Intn(2) == 0 {
			username += firstNames[rand.Intn(len(firstNames))]
		}
		if seen[username] {
			continue
		}
		seen[username] = true

		nickname := nickFirsts[rand.Intn(len(nickFirsts))]
		users = append(users, extraMockUser{
			Username: username,
			Nickname: nickname,
		})
	}

	// Phase B — overwrite chineseCount random indices with Chinese
	// nicknames built from cnBases × cnPatterns, interleaving them
	// across the full slice.
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

	for _, idx := range rand.Perm(len(users))[:chineseCount] {
		base := cnBases[rand.Intn(len(cnBases))]
		users[idx].Nickname = cnPatterns[rand.Intn(len(cnPatterns))](base)
	}

	return users
}
