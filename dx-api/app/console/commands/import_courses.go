package commands

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/lib/pq"
)

type ImportCourses struct{}

func (c *ImportCourses) Signature() string {
	return "app:import-courses"
}

func (c *ImportCourses) Description() string {
	return "Import English course JSON files into the database"
}

func (c *ImportCourses) Extend() command.Extend {
	return command.Extend{
		Arguments: []command.Argument{
			&command.ArgumentString{
				Name:     "path",
				Required: true,
				Usage:    "Path to the course directory",
			},
		},
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:  "force",
				Usage: "Force reimport by deleting existing games first",
			},
			&command.StringFlag{
				Name:  "category",
				Value: "实用英语",
				Usage: "Game category name to import under",
			},
		},
	}
}

func (c *ImportCourses) Handle(ctx console.Context) error {
	start := time.Now()
	dirPath := ctx.ArgumentString("path")
	force := ctx.OptionBool("force")

	// 1. Look up category (supports both parent and child categories)
	categoryName := ctx.Option("category")
	var category models.GameCategory
	if err := facades.Orm().Query().
		Where("name", categoryName).
		First(&category); err != nil || category.ID == "" {
		ctx.Error(fmt.Sprintf("category '%s' not found", categoryName))
		return fmt.Errorf("failed to find category: %w", err)
	}
	ctx.Info(fmt.Sprintf("category: %s (%s)", category.Name, category.ID))

	// Load game presses for matching against folder names
	pressMap := loadPressMap()
	ctx.Info(fmt.Sprintf("loaded %d presses", len(pressMap)))

	// 2. Load top 1202 user IDs
	userIDs, err := loadUserIDs(1202)
	if err != nil {
		ctx.Error(fmt.Sprintf("failed to load users: %v", err))
		return err
	}
	ctx.Info(fmt.Sprintf("loaded %d users", len(userIDs)))

	// 3. Read directory, filter subdirs, sort by name
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		ctx.Error(fmt.Sprintf("failed to read directory: %v", err))
		return err
	}

	var folders []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			folders = append(folders, e)
		}
	}
	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Name() < folders[j].Name()
	})
	ctx.Info(fmt.Sprintf("found %d course folders", len(folders)))

	// 4. Force cleanup if requested
	if force {
		names := make([]string, 0, len(folders))
		for _, f := range folders {
			names = append(names, cleanGameName(f.Name()))
		}
		cleaned, err := forceCleanup(category.ID, names)
		if err != nil {
			ctx.Error(fmt.Sprintf("force cleanup failed: %v", err))
			return err
		}
		if cleaned > 0 {
			ctx.Warning(fmt.Sprintf("force-cleaned %d existing games", cleaned))
		}
	}

	// 5. Process each folder
	var created, skipped int
	for folderIdx, folder := range folders {
		gameName := cleanGameName(folder.Name())

		// Skip if already exists
		if !force {
			var existing models.Game
			if err := facades.Orm().Query().
				Where("name", gameName).
				Where("game_category_id", category.ID).
				First(&existing); err == nil && existing.ID != "" {
				skipped++
				continue
			}
		}

		// Parse JSON files in folder
		folderPath := filepath.Join(dirPath, folder.Name())
		levels, totalItems, typeDist, err := parseLevels(folderPath)
		if err != nil {
			ctx.Error(fmt.Sprintf("[%d] %s: parse error: %v", folderIdx, gameName, err))
			continue
		}
		if len(levels) == 0 {
			skipped++
			continue
		}

		// Generate game description
		levelNames := make([]string, 0, len(levels))
		for _, l := range levels {
			levelNames = append(levelNames, l.Title)
		}
		desc := generateGameDescription(levelNames, totalItems, typeDist)

		// Begin transaction
		tx, err := facades.Orm().Query().Begin()
		if err != nil {
			ctx.Error(fmt.Sprintf("[%d] %s: failed to begin tx: %v", folderIdx, gameName, err))
			continue
		}

		// Create game
		gameID := uuid.Must(uuid.NewV7()).String()
		userID := userIDs[rand.IntN(len(userIDs))]
		pressID := matchPress(folder.Name(), pressMap)
		game := models.Game{
			ID:             gameID,
			Name:           gameName,
			Description:    &desc,
			UserID:         &userID,
			Mode:           "word-sentence",
			GameCategoryID: &category.ID,
			GamePressID:    pressID,
			Order:          float64(folderIdx * 1000),
			IsActive:       true,
			Status:         "published",
		}
		if err := tx.Create(&game); err != nil {
			_ = tx.Rollback()
			ctx.Error(fmt.Sprintf("[%d] %s: failed to create game: %v", folderIdx, gameName, err))
			continue
		}

		// Insert levels and content items
		if err := insertLevels(tx, gameID, levels); err != nil {
			_ = tx.Rollback()
			ctx.Error(fmt.Sprintf("[%d] %s: failed to insert levels: %v", folderIdx, gameName, err))
			continue
		}

		if err := tx.Commit(); err != nil {
			ctx.Error(fmt.Sprintf("[%d] %s: commit failed: %v", folderIdx, gameName, err))
			continue
		}

		created++
		ctx.Info(fmt.Sprintf("[%d/%d] %s — %d levels, %d items",
			folderIdx+1, len(folders), gameName, len(levels), totalItems))
	}

	// 6. Summary
	ctx.NewLine()
	ctx.Info(fmt.Sprintf("done in %s — created: %d, skipped: %d, total folders: %d",
		time.Since(start), created, skipped, len(folders)))
	return nil
}

// loadUserIDs returns up to limit user IDs ordered by created_at ASC.
func loadUserIDs(limit int) ([]string, error) {
	type userRow struct {
		ID string `gorm:"column:id"`
	}
	var rows []userRow
	if err := facades.Orm().Query().
		Model(&models.User{}).
		Order("created_at ASC").
		Limit(limit).
		Select("id").
		Get(&rows); err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	return ids, nil
}

// forceCleanup deletes games matching the given names under the category,
// along with their levels and content items.
func forceCleanup(categoryID string, names []string) (int, error) {
	query := facades.Orm().Query()
	var games []models.Game
	if err := query.
		Where("game_category_id", categoryID).
		Where("name IN ?", names).
		Get(&games); err != nil {
		return 0, fmt.Errorf("failed to query games for cleanup: %w", err)
	}

	for _, game := range games {
		var levels []models.GameLevel
		if err := query.Where("game_id", game.ID).Get(&levels); err != nil {
			return 0, fmt.Errorf("failed to query levels for game %s: %w", game.ID, err)
		}

		for _, level := range levels {
			if _, err := query.Where("game_level_id", level.ID).Delete(&models.ContentItem{}); err != nil {
				return 0, fmt.Errorf("failed to delete content items for level %s: %w", level.ID, err)
			}
		}

		if _, err := query.Where("game_id", game.ID).Delete(&models.GameLevel{}); err != nil {
			return 0, fmt.Errorf("failed to delete levels for game %s: %w", game.ID, err)
		}

		if _, err := query.Where("id", game.ID).Delete(&models.Game{}); err != nil {
			return 0, fmt.Errorf("failed to delete game %s: %w", game.ID, err)
		}
	}

	return len(games), nil
}

// parseLevels reads all .json files in a folder and returns parsed course data.
func parseLevels(folderPath string) ([]CourseFile, int, map[string]int, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to read folder: %w", err)
	}

	var jsonFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			jsonFiles = append(jsonFiles, e.Name())
		}
	}
	sort.Strings(jsonFiles)

	var levels []CourseFile
	totalItems := 0
	typeDist := make(map[string]int)

	for _, name := range jsonFiles {
		data, err := os.ReadFile(filepath.Join(folderPath, name))
		if err != nil {
			return nil, 0, nil, fmt.Errorf("failed to read %s: %w", name, err)
		}

		var cf CourseFile
		if err := json.Unmarshal(data, &cf); err != nil {
			return nil, 0, nil, fmt.Errorf("failed to parse %s: %w", name, err)
		}

		// Filter items with empty WordDetails
		var filtered []CourseItem
		for _, item := range cf.Sentences {
			if len(item.WordDetails) > 0 {
				filtered = append(filtered, item)
			}
		}
		cf.Sentences = filtered

		if len(cf.Sentences) == 0 {
			continue
		}

		for _, item := range cf.Sentences {
			typeDist[item.Type]++
		}
		totalItems += len(cf.Sentences)
		levels = append(levels, cf)
	}

	return levels, totalItems, typeDist, nil
}

// insertLevels creates game levels and their content items within a transaction.
func insertLevels(tx orm.Query, gameID string, levels []CourseFile) error {
	const batchSize = 100

	for levelIdx, level := range levels {
		levelID := uuid.Must(uuid.NewV7()).String()
		degrees := computeDegrees(level.Sentences)
		desc := generateLevelDescription(level.Sentences)

		gl := models.GameLevel{
			ID:           levelID,
			GameID:       gameID,
			Name:         level.Title,
			Description:  &desc,
			Order:        float64(levelIdx * 1000),
			PassingScore: 0,
			Degrees:      pq.StringArray(degrees),
			IsActive:     true,
		}
		if err := tx.Create(&gl); err != nil {
			return fmt.Errorf("failed to create level %s: %w", level.Title, err)
		}

		// Build content items in batches
		var batch []models.ContentItem
		for _, item := range level.Sentences {
			items, err := transformItems(item.Content, item.WordDetails)
			if err != nil {
				return fmt.Errorf("failed to transform items for %q: %w", item.Content, err)
			}

			structure, err := transformStructure(item.SentenceStructure)
			if err != nil {
				return fmt.Errorf("failed to transform structure for %q: %w", item.Content, err)
			}

			ci := models.ContentItem{
				ID:          uuid.Must(uuid.NewV7()).String(),
				GameLevelID: levelID,
				Content:     item.Content,
				ContentType: item.Type,
				Translation: &item.Chinese,
				Items:       &items,
				Structure:   structure,
				Order:       float64(item.SortOrder * 1000),
				IsActive:    true,
			}
			batch = append(batch, ci)

			if len(batch) >= batchSize {
				if err := tx.Create(&batch); err != nil {
					return fmt.Errorf("failed to batch create content items: %w", err)
				}
				batch = batch[:0]
			}
		}

		// Flush remaining
		if len(batch) > 0 {
			if err := tx.Create(&batch); err != nil {
				return fmt.Errorf("failed to batch create remaining content items: %w", err)
			}
		}
	}

	return nil
}

// loadPressMap loads all game presses into a name→ID map.
func loadPressMap() map[string]string {
	var presses []models.GamePress
	facades.Orm().Query().Get(&presses)
	m := make(map[string]string, len(presses))
	for _, p := range presses {
		m[p.Name] = p.ID
	}
	return m
}

// matchPress finds a press name within the folder name (longest match first).
func matchPress(folderName string, pressMap map[string]string) *string {
	cleaned := folderPrefixRe.ReplaceAllString(folderName, "")

	names := make([]string, 0, len(pressMap))
	for name := range pressMap {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return len([]rune(names[i])) > len([]rune(names[j]))
	})

	for _, name := range names {
		if strings.Contains(cleaned, name) {
			id := pressMap[name]
			return &id
		}
	}
	return nil
}
