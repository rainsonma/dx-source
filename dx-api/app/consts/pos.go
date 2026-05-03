package consts

// Part-of-speech keys used in ContentVocab.Definition JSON entries.
//
// Each entry in the definition JSON is a single-key object whose key
// is one of these constants and whose value is the Chinese gloss, e.g.:
//
//	[{"adj": "快的"}, {"v": "斋戒"}]
//
// Validation: IsValidPos rejects unknown keys to keep the wiki canonical.
const (
	PosNoun        = "n"
	PosVerb        = "v"
	PosAdjective   = "adj"
	PosAdverb      = "adv"
	PosPreposition = "prep"
	PosConjunction = "conj"
	PosPronoun     = "pron"
	PosArticle     = "art"
	PosNumeral     = "num"
	PosInterject   = "int"
	PosAuxiliary   = "aux"
	PosDeterminer  = "det"
)

// AllPos lists every supported POS key.
var AllPos = []string{
	PosNoun, PosVerb, PosAdjective, PosAdverb,
	PosPreposition, PosConjunction, PosPronoun, PosArticle,
	PosNumeral, PosInterject, PosAuxiliary, PosDeterminer,
}

// PosLabels maps POS keys to their Chinese labels for UI rendering.
var PosLabels = map[string]string{
	PosNoun:        "名词",
	PosVerb:        "动词",
	PosAdjective:   "形容词",
	PosAdverb:      "副词",
	PosPreposition: "介词",
	PosConjunction: "连词",
	PosPronoun:     "代词",
	PosArticle:     "冠词",
	PosNumeral:     "数词",
	PosInterject:   "感叹词",
	PosAuxiliary:   "助动词",
	PosDeterminer:  "限定词",
}

var posSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(AllPos))
	for _, p := range AllPos {
		m[p] = struct{}{}
	}
	return m
}()

// IsValidPos returns true if s is one of the canonical POS keys.
func IsValidPos(s string) bool {
	_, ok := posSet[s]
	return ok
}
