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

	lastNames := []string{
		"smith", "johnson", "williams", "brown", "jones", "garcia", "miller",
		"davis", "rodriguez", "martinez", "hernandez", "lopez", "gonzalez",
		"wilson", "anderson", "thomas", "taylor", "moore", "jackson", "martin",
		"lee", "perez", "thompson", "white", "harris", "sanchez", "clark",
		"ramirez", "lewis", "robinson", "walker", "young", "allen", "king",
		"wright", "scott", "torres", "nguyen", "hill", "flores", "green",
		"adams", "nelson", "baker", "hall", "rivera", "campbell", "mitchell",
		"carter", "roberts", "gomez", "phillips", "evans", "turner", "diaz",
		"parker", "cruz", "edwards", "collins", "reyes", "stewart", "morris",
		"morales", "murphy", "cook", "rogers", "gutierrez", "ortiz", "morgan",
		"cooper", "peterson", "bailey", "reed", "kelly", "howard", "ramos",
		"kim", "cox", "ward", "richardson", "watson", "brooks", "chavez",
		"wood", "james", "bennett", "gray", "mendoza", "ruiz", "hughes",
		"price", "alvarez", "castillo", "sanders", "patel", "myers", "long",
		"ross", "foster", "jimenez", "powell", "jenkins", "perry", "russell",
		"sullivan", "bell", "coleman", "butler", "henderson", "barnes", "gonzales",
		"fisher", "vasquez", "simmons", "griffin", "mcdonald", "hayes", "murray",
		"ford", "graham", "duncan", "stone", "logan", "hart", "webb",
		"fields", "chambers", "burns", "hardy", "west", "burke", "walsh",
		"lyons", "ramsey", "steele", "barton", "howe", "bishop", "larson",
		"klein", "bauer", "lindgren", "sato", "tanaka", "nakamura", "ito",
		"fischer", "weber", "schmidt", "meyer", "wagner", "becker", "schulz",
		"hansen", "olsen", "berg", "lund", "dahl", "holm", "strand",
		"dubois", "laurent", "moreau", "durand", "petit", "blanc", "garnier",
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

	nickLasts := []string{
		"S.", "J.", "W.", "B.", "D.", "G.", "M.", "T.", "R.", "L.",
		"H.", "K.", "P.", "N.", "C.", "F.", "E.", "A.", "V.", "O.",
	}

	// Patterns for username generation.
	patterns := []func(first, last string, idx int) string{
		// first.last
		func(f, l string, i int) string { return fmt.Sprintf("%s.%s", f, l) },
		// first_last
		func(f, l string, i int) string { return fmt.Sprintf("%s_%s", f, l) },
		// firstlast
		func(f, l string, i int) string { return fmt.Sprintf("%s%s", f, l) },
		// first.last + digits
		func(f, l string, i int) string { return fmt.Sprintf("%s.%s%d", f, l, rand.Intn(99)+1) },
		// first + digits
		func(f, l string, i int) string { return fmt.Sprintf("%s%d", f, rand.Intn(999)+1) },
		// first_last + digits
		func(f, l string, i int) string { return fmt.Sprintf("%s_%s%d", f, l, rand.Intn(99)+1) },
	}

	seen := make(map[string]bool)
	users := make([]mockUser, 0, 1200)

	for len(users) < 1200 {
		fi := rand.Intn(len(firstNames))
		li := rand.Intn(len(lastNames))
		pi := rand.Intn(len(patterns))

		username := patterns[pi](firstNames[fi], lastNames[li], len(users))
		if seen[username] || username == "rainson" || username == "june" {
			continue
		}
		seen[username] = true

		nfi := rand.Intn(len(nickFirsts))
		nli := rand.Intn(len(nickLasts))
		nickname := fmt.Sprintf("%s %s", nickFirsts[nfi], nickLasts[nli])

		users = append(users, mockUser{Username: username, Nickname: nickname})
	}

	return users
}
