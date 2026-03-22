package seeders

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/goravel/framework/facades"
	"dx-api/app/models"
)

type ContentItemSeeder struct{}

func (s *ContentItemSeeder) Signature() string {
	return "ContentItemSeeder"
}

type itemDef struct {
	LevelName      string
	MetaSourceData string
	Content        string
	ContentType    string
	Translation    string
	Order          float64
	Items          []map[string]any
}

func contentItemDefs() []itemDef {
	return []itemDef{
		// --- Level 1 (第一关), Meta "The food is ready." ---
		{"第一关", "The food is ready.", "The food", "phrase", "食物", 1010, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "food", "answer": true, "phonetic": map[string]any{"uk": "/fuːd/", "us": "/fuːd/"}, "position": 2, "translation": "食物"},
		}},
		{"第一关", "The food is ready.", "is", "word", "是", 1020, []map[string]any{
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 1, "translation": "是"},
		}},
		{"第一关", "The food is ready.", "The food is", "block", "食物是", 1030, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "food", "answer": true, "phonetic": map[string]any{"uk": "/fuːd/", "us": "/fuːd/"}, "position": 2, "translation": "食物"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 3, "translation": "是"},
		}},
		{"第一关", "The food is ready.", "ready", "word", "准备好了", 1040, []map[string]any{
			{"pos": "形容词", "item": "ready", "answer": true, "phonetic": map[string]any{"uk": "/ˈredi/", "us": "/ˈredi/"}, "position": 1, "translation": "准备好的"},
		}},
		{"第一关", "The food is ready.", "The food is ready.", "sentence", "食物准备好了。", 1050, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "food", "answer": true, "phonetic": map[string]any{"uk": "/fuːd/", "us": "/fuːd/"}, "position": 2, "translation": "食物"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 3, "translation": "是"},
			{"pos": "形容词", "item": "ready", "answer": true, "phonetic": map[string]any{"uk": "/ˈredi/", "us": "/ˈredi/"}, "position": 4, "translation": "准备好的"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 5, "translation": ""},
		}},

		// --- Level 1 (第一关), Meta "I am very hungry." ---
		{"第一关", "I am very hungry.", "I", "word", "我", 2010, []map[string]any{
			{"pos": "代词", "item": "I", "answer": true, "phonetic": map[string]any{"uk": "/aɪ/", "us": "/aɪ/"}, "position": 1, "translation": "我"},
		}},
		{"第一关", "I am very hungry.", "am", "word", "是", 2020, []map[string]any{
			{"pos": "助动词", "item": "am", "answer": true, "phonetic": map[string]any{"uk": "/æm/", "us": "/æm/"}, "position": 1, "translation": "是"},
		}},
		{"第一关", "I am very hungry.", "I am", "block", "我是", 2030, []map[string]any{
			{"pos": "代词", "item": "I", "answer": true, "phonetic": map[string]any{"uk": "/aɪ/", "us": "/aɪ/"}, "position": 1, "translation": "我"},
			{"pos": "助动词", "item": "am", "answer": true, "phonetic": map[string]any{"uk": "/æm/", "us": "/æm/"}, "position": 2, "translation": "是"},
		}},
		{"第一关", "I am very hungry.", "very hungry", "phrase", "非常饿", 2040, []map[string]any{
			{"pos": "副词", "item": "very", "answer": true, "phonetic": map[string]any{"uk": "/ˈveri/", "us": "/ˈveri/"}, "position": 1, "translation": "非常"},
			{"pos": "形容词", "item": "hungry", "answer": true, "phonetic": map[string]any{"uk": "/ˈhʌŋɡri/", "us": "/ˈhʌŋɡri/"}, "position": 2, "translation": "饥饿的"},
		}},
		{"第一关", "I am very hungry.", "I am very hungry.", "sentence", "我非常饿。", 2050, []map[string]any{
			{"pos": "代词", "item": "I", "answer": true, "phonetic": map[string]any{"uk": "/aɪ/", "us": "/aɪ/"}, "position": 1, "translation": "我"},
			{"pos": "助动词", "item": "am", "answer": true, "phonetic": map[string]any{"uk": "/æm/", "us": "/æm/"}, "position": 2, "translation": "是"},
			{"pos": "副词", "item": "very", "answer": true, "phonetic": map[string]any{"uk": "/ˈveri/", "us": "/ˈveri/"}, "position": 3, "translation": "非常"},
			{"pos": "形容词", "item": "hungry", "answer": true, "phonetic": map[string]any{"uk": "/ˈhʌŋɡri/", "us": "/ˈhʌŋɡri/"}, "position": 4, "translation": "饥饿的"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 5, "translation": ""},
		}},

		// --- Level 1 (第一关), Meta "It is a good day." ---
		{"第一关", "It is a good day.", "It", "word", "它", 3010, []map[string]any{
			{"pos": "代词", "item": "It", "answer": true, "phonetic": map[string]any{"uk": "/ɪt/", "us": "/ɪt/"}, "position": 1, "translation": "它"},
		}},
		{"第一关", "It is a good day.", "is", "word", "是", 3020, []map[string]any{
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 1, "translation": "是"},
		}},
		{"第一关", "It is a good day.", "It is", "block", "它是", 3030, []map[string]any{
			{"pos": "代词", "item": "It", "answer": true, "phonetic": map[string]any{"uk": "/ɪt/", "us": "/ɪt/"}, "position": 1, "translation": "它"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 2, "translation": "是"},
		}},
		{"第一关", "It is a good day.", "a good day", "phrase", "一个好日子", 3040, []map[string]any{
			{"pos": "冠词", "item": "a", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 1, "translation": "一个"},
			{"pos": "形容词", "item": "good", "answer": true, "phonetic": map[string]any{"uk": "/ɡʊd/", "us": "/ɡʊd/"}, "position": 2, "translation": "好的"},
			{"pos": "名词", "item": "day", "answer": true, "phonetic": map[string]any{"uk": "/deɪ/", "us": "/deɪ/"}, "position": 3, "translation": "天"},
		}},
		{"第一关", "It is a good day.", "It is a good day.", "sentence", "它是一个好日子。", 3050, []map[string]any{
			{"pos": "代词", "item": "It", "answer": true, "phonetic": map[string]any{"uk": "/ɪt/", "us": "/ɪt/"}, "position": 1, "translation": "它"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 2, "translation": "是"},
			{"pos": "冠词", "item": "a", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 3, "translation": "一个"},
			{"pos": "形容词", "item": "good", "answer": true, "phonetic": map[string]any{"uk": "/ɡʊd/", "us": "/ɡʊd/"}, "position": 4, "translation": "好的"},
			{"pos": "名词", "item": "day", "answer": true, "phonetic": map[string]any{"uk": "/deɪ/", "us": "/deɪ/"}, "position": 5, "translation": "天"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 6, "translation": ""},
		}},

		// --- Level 2 (第二关), Meta "A car is on the road." ---
		{"第二关", "A car is on the road.", "A car", "phrase", "一辆汽车", 1010, []map[string]any{
			{"pos": "冠词", "item": "A", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 1, "translation": "一个"},
			{"pos": "名词", "item": "car", "answer": true, "phonetic": map[string]any{"uk": "/kɑː(r)/", "us": "/kɑːr/"}, "position": 2, "translation": "汽车"},
		}},
		{"第二关", "A car is on the road.", "is", "word", "是", 1020, []map[string]any{
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 1, "translation": "是"},
		}},
		{"第二关", "A car is on the road.", "A car is", "block", "一辆汽车是", 1030, []map[string]any{
			{"pos": "冠词", "item": "A", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 1, "translation": "一个"},
			{"pos": "名词", "item": "car", "answer": true, "phonetic": map[string]any{"uk": "/kɑː(r)/", "us": "/kɑːr/"}, "position": 2, "translation": "汽车"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 3, "translation": "是"},
		}},
		{"第二关", "A car is on the road.", "on the road", "phrase", "在路上", 1040, []map[string]any{
			{"pos": "介词", "item": "on", "answer": true, "phonetic": map[string]any{"uk": "/ɒn/", "us": "/ɑːn/"}, "position": 1, "translation": "在...上"},
			{"pos": "冠词", "item": "the", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 2, "translation": "这"},
			{"pos": "名词", "item": "road", "answer": true, "phonetic": map[string]any{"uk": "/rəʊd/", "us": "/roʊd/"}, "position": 3, "translation": "道路"},
		}},
		{"第二关", "A car is on the road.", "A car is on the road.", "sentence", "一辆汽车在路上。", 1050, []map[string]any{
			{"pos": "冠词", "item": "A", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 1, "translation": "一个"},
			{"pos": "名词", "item": "car", "answer": true, "phonetic": map[string]any{"uk": "/kɑː(r)/", "us": "/kɑːr/"}, "position": 2, "translation": "汽车"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 3, "translation": "是"},
			{"pos": "介词", "item": "on", "answer": true, "phonetic": map[string]any{"uk": "/ɒn/", "us": "/ɑːn/"}, "position": 4, "translation": "在...上"},
			{"pos": "冠词", "item": "the", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 5, "translation": "这"},
			{"pos": "名词", "item": "road", "answer": true, "phonetic": map[string]any{"uk": "/rəʊd/", "us": "/roʊd/"}, "position": 6, "translation": "道路"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 7, "translation": ""},
		}},

		// --- Level 2 (第二关), Meta "It is a red car." ---
		{"第二关", "It is a red car.", "It", "word", "它", 2010, []map[string]any{
			{"pos": "代词", "item": "It", "answer": true, "phonetic": map[string]any{"uk": "/ɪt/", "us": "/ɪt/"}, "position": 1, "translation": "它"},
		}},
		{"第二关", "It is a red car.", "is", "word", "是", 2020, []map[string]any{
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 1, "translation": "是"},
		}},
		{"第二关", "It is a red car.", "It is", "block", "它是", 2030, []map[string]any{
			{"pos": "代词", "item": "It", "answer": true, "phonetic": map[string]any{"uk": "/ɪt/", "us": "/ɪt/"}, "position": 1, "translation": "它"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 2, "translation": "是"},
		}},
		{"第二关", "It is a red car.", "a red car", "phrase", "一辆红色的汽车", 2040, []map[string]any{
			{"pos": "冠词", "item": "a", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 1, "translation": "一个"},
			{"pos": "形容词", "item": "red", "answer": true, "phonetic": map[string]any{"uk": "/red/", "us": "/red/"}, "position": 2, "translation": "红色的"},
			{"pos": "名词", "item": "car", "answer": true, "phonetic": map[string]any{"uk": "/kɑː(r)/", "us": "/kɑːr/"}, "position": 3, "translation": "汽车"},
		}},
		{"第二关", "It is a red car.", "It is a red car.", "sentence", "它是一辆红色的汽车。", 2050, []map[string]any{
			{"pos": "代词", "item": "It", "answer": true, "phonetic": map[string]any{"uk": "/ɪt/", "us": "/ɪt/"}, "position": 1, "translation": "它"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 2, "translation": "是"},
			{"pos": "冠词", "item": "a", "answer": true, "phonetic": map[string]any{"uk": "/ə/", "us": "/ə/"}, "position": 3, "translation": "一个"},
			{"pos": "形容词", "item": "red", "answer": true, "phonetic": map[string]any{"uk": "/red/", "us": "/red/"}, "position": 4, "translation": "红色的"},
			{"pos": "名词", "item": "car", "answer": true, "phonetic": map[string]any{"uk": "/kɑː(r)/", "us": "/kɑːr/"}, "position": 5, "translation": "汽车"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 6, "translation": ""},
		}},

		// --- Level 2 (第二关), Meta "The driver is happy." ---
		{"第二关", "The driver is happy.", "The driver", "phrase", "司机", 3010, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "driver", "answer": true, "phonetic": map[string]any{"uk": "/ˈdraɪvə(r)/", "us": "/ˈdraɪvər/"}, "position": 2, "translation": "司机"},
		}},
		{"第二关", "The driver is happy.", "is", "word", "是", 3020, []map[string]any{
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 1, "translation": "是"},
		}},
		{"第二关", "The driver is happy.", "The driver is", "block", "司机是", 3030, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "driver", "answer": true, "phonetic": map[string]any{"uk": "/ˈdraɪvə(r)/", "us": "/ˈdraɪvər/"}, "position": 2, "translation": "司机"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 3, "translation": "是"},
		}},
		{"第二关", "The driver is happy.", "happy", "word", "高兴的", 3040, []map[string]any{
			{"pos": "形容词", "item": "happy", "answer": true, "phonetic": map[string]any{"uk": "/ˈhæpi/", "us": "/ˈhæpi/"}, "position": 1, "translation": "高兴的"},
		}},
		{"第二关", "The driver is happy.", "The driver is happy.", "sentence", "司机很高兴。", 3050, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "driver", "answer": true, "phonetic": map[string]any{"uk": "/ˈdraɪvə(r)/", "us": "/ˈdraɪvər/"}, "position": 2, "translation": "司机"},
			{"pos": "动词", "item": "is", "answer": true, "phonetic": map[string]any{"uk": "/ɪz/", "us": "/ɪz/"}, "position": 3, "translation": "是"},
			{"pos": "形容词", "item": "happy", "answer": true, "phonetic": map[string]any{"uk": "/ˈhæpi/", "us": "/ˈhæpi/"}, "position": 4, "translation": "高兴的"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 5, "translation": ""},
		}},

		// --- Level 3 (第三关), Meta "The children go to school." ---
		{"第三关", "The children go to school.", "The children", "phrase", "孩子们", 1010, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "children", "answer": true, "phonetic": map[string]any{"uk": "/ˈtʃɪldrən/", "us": "/ˈtʃɪldrən/"}, "position": 2, "translation": "孩子们"},
		}},
		{"第三关", "The children go to school.", "go", "word", "去", 1020, []map[string]any{
			{"pos": "动词", "item": "go", "answer": true, "phonetic": map[string]any{"uk": "/ɡəʊ/", "us": "/ɡoʊ/"}, "position": 1, "translation": "去"},
		}},
		{"第三关", "The children go to school.", "The children go", "block", "孩子们去", 1030, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "children", "answer": true, "phonetic": map[string]any{"uk": "/ˈtʃɪldrən/", "us": "/ˈtʃɪldrən/"}, "position": 2, "translation": "孩子们"},
			{"pos": "动词", "item": "go", "answer": true, "phonetic": map[string]any{"uk": "/ɡəʊ/", "us": "/ɡoʊ/"}, "position": 3, "translation": "去"},
		}},
		{"第三关", "The children go to school.", "to school", "phrase", "去上学", 1040, []map[string]any{
			{"pos": "介词", "item": "to", "answer": true, "phonetic": map[string]any{"uk": "/tuː/", "us": "/tuː/"}, "position": 1, "translation": "到"},
			{"pos": "名词", "item": "school", "answer": true, "phonetic": map[string]any{"uk": "/skuːl/", "us": "/skuːl/"}, "position": 2, "translation": "学校"},
		}},
		{"第三关", "The children go to school.", "The children go to school.", "sentence", "孩子们去上学。", 1050, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "children", "answer": true, "phonetic": map[string]any{"uk": "/ˈtʃɪldrən/", "us": "/ˈtʃɪldrən/"}, "position": 2, "translation": "孩子们"},
			{"pos": "动词", "item": "go", "answer": true, "phonetic": map[string]any{"uk": "/ɡəʊ/", "us": "/ɡoʊ/"}, "position": 3, "translation": "去"},
			{"pos": "介词", "item": "to", "answer": true, "phonetic": map[string]any{"uk": "/tuː/", "us": "/tuː/"}, "position": 4, "translation": "到"},
			{"pos": "名词", "item": "school", "answer": true, "phonetic": map[string]any{"uk": "/skuːl/", "us": "/skuːl/"}, "position": 5, "translation": "学校"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 6, "translation": ""},
		}},

		// --- Level 3 (第三关), Meta "The bell rings." ---
		{"第三关", "The bell rings.", "The bell", "phrase", "铃", 2010, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "bell", "answer": true, "phonetic": map[string]any{"uk": "/bel/", "us": "/bel/"}, "position": 2, "translation": "铃"},
		}},
		{"第三关", "The bell rings.", "rings", "word", "响", 2020, []map[string]any{
			{"pos": "动词", "item": "rings", "answer": true, "phonetic": map[string]any{"uk": "/rɪŋz/", "us": "/rɪŋz/"}, "position": 1, "translation": "响"},
		}},
		{"第三关", "The bell rings.", "The bell rings.", "sentence", "铃响了。", 2030, []map[string]any{
			{"pos": "冠词", "item": "The", "answer": true, "phonetic": map[string]any{"uk": "/ðə/", "us": "/ðə/"}, "position": 1, "translation": "这"},
			{"pos": "名词", "item": "bell", "answer": true, "phonetic": map[string]any{"uk": "/bel/", "us": "/bel/"}, "position": 2, "translation": "铃"},
			{"pos": "动词", "item": "rings", "answer": true, "phonetic": map[string]any{"uk": "/rɪŋz/", "us": "/rɪŋz/"}, "position": 3, "translation": "响"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 4, "translation": ""},
		}},

		// --- Level 3 (第三关), Meta "They go home." ---
		{"第三关", "They go home.", "They", "word", "他们", 3010, []map[string]any{
			{"pos": "代词", "item": "They", "answer": true, "phonetic": map[string]any{"uk": "/ðeɪ/", "us": "/ðeɪ/"}, "position": 1, "translation": "他们"},
		}},
		{"第三关", "They go home.", "go", "word", "去", 3020, []map[string]any{
			{"pos": "动词", "item": "go", "answer": true, "phonetic": map[string]any{"uk": "/ɡəʊ/", "us": "/ɡoʊ/"}, "position": 1, "translation": "去"},
		}},
		{"第三关", "They go home.", "They go", "block", "他们去", 3030, []map[string]any{
			{"pos": "代词", "item": "They", "answer": true, "phonetic": map[string]any{"uk": "/ðeɪ/", "us": "/ðeɪ/"}, "position": 1, "translation": "他们"},
			{"pos": "动词", "item": "go", "answer": true, "phonetic": map[string]any{"uk": "/ɡəʊ/", "us": "/ɡoʊ/"}, "position": 2, "translation": "去"},
		}},
		{"第三关", "They go home.", "home", "word", "家", 3040, []map[string]any{
			{"pos": "名词", "item": "home", "answer": true, "phonetic": map[string]any{"uk": "/həʊm/", "us": "/hoʊm/"}, "position": 1, "translation": "家"},
		}},
		{"第三关", "They go home.", "They go home.", "sentence", "他们回家。", 3050, []map[string]any{
			{"pos": "代词", "item": "They", "answer": true, "phonetic": map[string]any{"uk": "/ðeɪ/", "us": "/ðeɪ/"}, "position": 1, "translation": "他们"},
			{"pos": "动词", "item": "go", "answer": true, "phonetic": map[string]any{"uk": "/ɡəʊ/", "us": "/ɡoʊ/"}, "position": 2, "translation": "去"},
			{"pos": "名词", "item": "home", "answer": true, "phonetic": map[string]any{"uk": "/həʊm/", "us": "/hoʊm/"}, "position": 3, "translation": "家"},
			{"pos": nil, "item": ".", "answer": false, "phonetic": nil, "position": 4, "translation": ""},
		}},
	}
}

func (s *ContentItemSeeder) Run() error {
	query := facades.Orm().Query()
	items := contentItemDefs()

	// Get only the 50 seeded games by name
	gameDefs := buildGameDefs()
	gameNames := make([]any, len(gameDefs))
	for i, g := range gameDefs {
		gameNames[i] = g.Name
	}
	var games []models.Game
	if err := query.WhereIn("name", gameNames).Get(&games); err != nil {
		return fmt.Errorf("failed to query games: %w", err)
	}

	count := 0
	for _, game := range games {
		// Build level name→ID map
		var levels []models.GameLevel
		if err := query.Where("game_id", game.ID).Get(&levels); err != nil {
			return fmt.Errorf("failed to query levels for game %s: %w", game.Name, err)
		}
		levelIDs := make(map[string]string)
		for _, l := range levels {
			levelIDs[l.Name] = l.ID
		}

		// Build meta (levelID:sourceData)→ID map
		levelIDList := make([]any, 0, len(levelIDs))
		for _, id := range levelIDs {
			levelIDList = append(levelIDList, id)
		}
		var metas []models.ContentMeta
		if err := query.WhereIn("game_level_id", levelIDList).Get(&metas); err != nil {
			return fmt.Errorf("failed to query metas for game %s: %w", game.Name, err)
		}
		metaIDs := make(map[string]string)
		for _, m := range metas {
			key := m.GameLevelID + ":" + m.SourceData
			metaIDs[key] = m.ID
		}

		for _, item := range items {
			levelID, ok := levelIDs[item.LevelName]
			if !ok {
				continue
			}

			metaKey := levelID + ":" + item.MetaSourceData
			metaID, ok := metaIDs[metaKey]
			if !ok {
				continue
			}

			// Serialize items to JSON string
			itemsJSON, err := json.Marshal(item.Items)
			if err != nil {
				return fmt.Errorf("failed to marshal items: %w", err)
			}
			itemsStr := string(itemsJSON)
			translation := item.Translation

			var existing models.ContentItem
			if err := query.Where("content", item.Content).Where("content_type", item.ContentType).Where("content_meta_id", metaID).First(&existing); err != nil || existing.ID == "" {
				if err := query.Create(&models.ContentItem{
					ID:            ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
					GameLevelID:   levelID,
					ContentMetaID: &metaID,
					Content:       item.Content,
					ContentType:   item.ContentType,
					Translation:   &translation,
					Items:         &itemsStr,
					Order:         item.Order,
					IsActive:      true,
				}); err != nil {
					return fmt.Errorf("failed to create content item '%s': %w", item.Content, err)
				}
			} else {
				if _, err := query.Where("content", item.Content).Where("content_type", item.ContentType).Where("content_meta_id", metaID).Update(&models.ContentItem{
					GameLevelID: levelID,
					Translation: &translation,
					Items:       &itemsStr,
					Order:       item.Order,
					IsActive:    true,
				}); err != nil {
					return fmt.Errorf("failed to update content item '%s': %w", item.Content, err)
				}
			}
			count++
		}
	}

	log.Printf("Seeded %d content items\n", count)
	return nil
}
