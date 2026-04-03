# Import Courses Design Spec

## Overview

Import 47 game courses from `/dx-courses-copy/实用英语` JSON files into the existing game system (`games`, `game_levels`, `content_items` tables). All courses belong to the 实用英语 category.

## Data Scale

| Metric | Count |
|--------|-------|
| Game folders (→ games) | 47 |
| JSON files (→ game_levels) | 2,971 |
| Content items total | ~82,983 |
| Items with sentenceStructure | ~17,910 (~22%) |
| Type distribution | word: 24,986 / phrase: 21,351 / sentence: 34,293 / block: 2,353 |

## Implementation: Go CLI Command

New Goravel console command: `app:import-courses <directory-path>`

**File:** `app/console/commands/import_courses.go`
**Registration:** `bootstrap/app.go`

```
go run . artisan app:import-courses /path/to/实用英语
```

### Processing Flow

1. Look up 实用英语 category ID from `game_categories`
2. Load top 1202 user IDs from `users` table (ordered by `created_at ASC`, limit 1202)
3. Sort and iterate folders → create `games` records (random user from pool)
4. Sort and iterate JSON files per folder → create `game_levels` records
5. Iterate items per file → create `content_items` records (skip items with empty `wordDetails`)
6. All inserts wrapped in per-game transaction; batch insert content_items (100 per batch)

### Idempotency

- Default: skip games that already exist (matching name + category)
- `--force` flag: delete all content_items and game_levels for games under 实用英语 category that were created by this command (matched by name), delete those games, then reimport

## Table Mappings

### games

| Field | Value |
|-------|-------|
| `id` | UUID v7 |
| `name` | Folder name with `^\d+_` prefix stripped and `【】` removed |
| `description` | Auto-generated ~200 chars from real data |
| `user_id` | Random UUID from top 1202 users |
| `mode` | `"word-sentence"` |
| `game_category_id` | 实用英语 category ID |
| `game_press_id` | null |
| `icon` | null |
| `cover_id` | null |
| `order` | Sorted folder index * 1000 |
| `is_active` | true |
| `status` | `"published"` |

### game_levels

| Field | Value |
|-------|-------|
| `id` | UUID v7 |
| `game_id` | Parent game ID |
| `name` | JSON `title` field |
| `description` | Auto-generated ~200 chars from real data |
| `order` | Sorted file index * 1000 |
| `passing_score` | 0 |
| `degrees` | Computed from content types (see below) |
| `is_active` | true |

### content_items (no content_metas)

| Field | Value |
|-------|-------|
| `id` | UUID v7 |
| `game_level_id` | Parent level ID |
| `content_meta_id` | null |
| `content` | JSON item `content` |
| `content_type` | JSON item `type` (word/phrase/sentence/block) |
| `uk_audio_id` | null |
| `us_audio_id` | null |
| `definition` | null |
| `translation` | JSON item `chinese` |
| `explanation` | null |
| `items` | Transformed wordDetails JSONB |
| `structure` | Transformed sentenceStructure JSONB (or null) |
| `order` | `sortOrder * 1000` |
| `tags` | null |
| `is_active` | true |

## Degrees Calculation

Scan all item types in a level's JSON file:

- Has any `"word"` type → `{"beginner", "intermediate", "advanced"}`
- Has `"block"` or `"phrase"` (no word) → `{"intermediate", "advanced"}`
- Only `"sentence"` → `{"advanced"}`

Reference:
- beginner allows: word, block, phrase, sentence
- intermediate allows: block, phrase, sentence
- advanced allows: sentence only

## Data Transformations

### Game Name Cleaning

Strip `^\d+_` prefix, remove `【` and `】` characters.

Examples:
- `07_【DK】基础3000词` → `DK基础3000词`
- `01_日常英语对话100句` → `日常英语对话100句`
- `11_【新东方】100个句子记完4500个四级单词` → `新东方100个句子记完4500个四级单词`

### Auto-generated Descriptions (~200 chars)

**Game description** template:
```
共{levelCount}个学习单元，{itemCount}个学习内容。
{typeBreakdown}，涵盖{level1Name}、{level2Name}、{level3Name}等主题。
```

**Game level description** template:
```
收录{itemCount}个{primaryType}，
如「{sample1}」「{sample2}」等，{supplementInfo}。
```

Type names: word→"词汇", phrase→"短语", sentence→"句子", block→"语段", mixed→"词句"

Truncate to ~200 chars if exceeded.

### wordDetails → items JSONB

| Source Field | Target Field | Transform |
|-------------|-------------|-----------|
| (array index) | `position` | Sequential across words + punctuation |
| `word` | `item` | Direct |
| `phonetic.uk` | `phonetic.uk` | Wrap with `/` if non-empty |
| `phonetic.us` | `phonetic.us` | Wrap with `/` if non-empty |
| `pos` | `pos` | Direct (nullable) |
| `definition` | `translation` | Direct |
| — | `answer` | `true` for words, `false` for punctuation |

### Punctuation Injection

Tokenize the `content` string by whitespace. For each token, strip trailing punctuation and match word part to next `wordDetails` entry. Punctuation characters become separate items with `answer: false`.

Punctuation translations:
- `.` → 句号
- `,` → 逗号
- `!` → 感叹号
- `?` → 问号
- `;` → 分号
- `:` → 冒号
- `"` → 引号
- `'` → 撇号
- `-` → 连字符
- `(` → 左括号
- `)` → 右括号

### sentenceStructure → structure JSONB

| Source Field | Target Field | Transform |
|-------------|-------------|-----------|
| `start` | `start` | `+ 1` (0-based → 1-based) |
| `end` | `end` | `+ 1` (0-based → 1-based) |
| `text` | `content` | Direct |
| `role` | `role` | Direct |
| `type` | `role_en` | Direct |
| `explanation` | `explanation` | Direct |
| — | `color` | Mapped from role (see below) |

### Structure Color Map (Light Pastels for Background)

| Role | Color |
|------|-------|
| 主语 (subject) | `#FFF3E0` |
| 谓语 (predicate) | `#FCE4EC` |
| 宾语 (object) | `#E3F2FD` |
| 表语 (predicative) | `#E8F5E9` |
| 定语 (attributive) | `#F3E5F5` |
| 状语 (adverbial) | `#E0F7FA` |
| 补语 (complement) | `#FFF9C4` |
| 同位语 (appositive) | `#EFEBE9` |
| 插入语 (parenthetical) | `#F1F8E9` |
| 标点符号 (punctuation) | null |
| Unknown role | `#F5F5F5` |

### Phonetic Normalization

- Non-empty string: wrap with `/` → `/ɪkˈskjuːs/`
- Empty string `""`: keep as `""`
- Null: keep as null

## Error Handling

**Transaction scope:** One transaction per game. If any level fails, entire game is rolled back and skipped.

**Skip rules:**
- Items with empty `wordDetails` array → skip (not inserted)
- Non-JSON files in folder → skip silently
- Empty folders → skip

**Progress output:**
```
Importing 47 games from /path/to/实用英语...
[1/47] DK基础3000词 (120 levels, 3000 items) ✓
[2/47] 日常英语对话100句 (17 levels, 100 items) ✓
[3/47] 新概念英语第一册 (72 levels, 1440 items) ✗ error: ...
...
Done: 46 succeeded, 1 failed.
```

## Not In Scope

- `content_metas` table — not populated (content_meta_id is null on all items)
- Audio files — uk_audio_id and us_audio_id are null
- Multi-sentence block splitting — blocks are imported as-is
- Cover images — cover_id is null
- Game press assignment — game_press_id is null
