package consts

// Source origin values.
const (
	SourceFromManual = "manual"
	SourceFromAI     = "ai"
	SourceFromImport = "import"
)

// SourceFromLabels maps each source origin to its Chinese label.
var SourceFromLabels = map[string]string{
	SourceFromManual: "手动添加",
	SourceFromAI:     "AI 生成",
	SourceFromImport: "导入",
}
