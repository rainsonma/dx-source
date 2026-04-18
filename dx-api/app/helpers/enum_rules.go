package helpers

import (
	"strings"

	"dx-api/app/consts"
)

var enumValues = map[string][]string{
	"degree":         {consts.GameDegreeBeginner, consts.GameDegreeIntermediate, consts.GameDegreeAdvanced},
	"pattern":        {consts.GamePatternListen, consts.GamePatternSpeak, consts.GamePatternRead, consts.GamePatternWrite},
	"mode":           {consts.GameModeWordSentence, consts.GameModeVocabBattle, consts.GameModeVocabMatch, consts.GameModeVocabElimination},
	"feedback_type":  {consts.FeedbackTypeFeature, consts.FeedbackTypeContent, consts.FeedbackTypeUX, consts.FeedbackTypeBug, consts.FeedbackTypeOther},
	"content_type":   {consts.ContentTypeWord, consts.ContentTypeBlock, consts.ContentTypePhrase, consts.ContentTypeSentence},
	"image_role":     {consts.ImageRoleAdmUserAvatar, consts.ImageRoleUserAvatar, consts.ImageRoleCategoryCover, consts.ImageRoleTemplateCover, consts.ImageRoleGameCover, consts.ImageRolePressCover, consts.ImageRoleGameGroupCover, consts.ImageRolePostImage},
	"source_from":    {consts.SourceFromManual, consts.SourceFromAI},
	"source_type":    {consts.SourceTypeSentence, consts.SourceTypeVocab},
	"grade":          {consts.UserGradeMonth, consts.UserGradeSeason, consts.UserGradeYear, consts.UserGradeLifetime},
	"payment_method": {consts.PaymentMethodWechat, consts.PaymentMethodAlipay},
	"bean_package":   consts.BeanPackageSlugs,
	"pk_difficulty":  consts.PkDifficultySlugs,
}

// InEnum returns an "in:val1,val2,..." rule string for the named enum.
// Panics on unknown enum names to catch typos at startup.
func InEnum(name string) string {
	vals, ok := enumValues[name]
	if !ok {
		panic("unknown enum: " + name)
	}
	return "in:" + strings.Join(vals, ",")
}
