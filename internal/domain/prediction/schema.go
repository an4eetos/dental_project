package prediction

import "strings"

// Field describes one input or output column shared by Excel, API, and UI.
type Field struct {
	Key    string
	Column string
	Label  string
}

const (
	KeyAge                          = "age"
	KeyPregnancyWeeks               = "pregnancy_weeks"
	KeyKPUIndex                     = "kpu_index"
	KeyBrushingPerDay               = "brushing_per_day"
	KeyDentistVisitDuringPregnancy  = "dentist_visit_during_pregnancy"
	KeyParentCaries                 = "parent_caries"
	KeySalivaPH                     = "saliva_ph"

	KeyChildCariesProbability = "child_caries_probability"
	KeyRiskGroup              = "risk_group"
	KeyAction                   = "action"
	KeyRecommendations        = "recommendations"
)

// InputFields defines the seven survey questions (Excel column headers).
var InputFields = []Field{
	{Key: KeyAge, Column: "Ваш возраст (полных лет)?", Label: "Ваш возраст (полных лет)?"},
	{Key: KeyPregnancyWeeks, Column: "Срок беременности (недель)?", Label: "Срок беременности (недель)?"},
	{
		Key:    KeyKPUIndex,
		Column: "Индекс КПУ (количество кариозных, пломбированных и удаленных зубов) — по данным вашего стоматолога:",
		Label:  "Индекс КПУ (количество кариозных, пломбированных и удаленных зубов) — по данным вашего стоматолога:",
	},
	{Key: KeyBrushingPerDay, Column: "Сколько раз в день вы чистите зубы?", Label: "Сколько раз в день вы чистите зубы?"},
	{
		Key:    KeyDentistVisitDuringPregnancy,
		Column: "Посещали ли вы стоматолога во время беременности?",
		Label:  "Посещали ли вы стоматолога во время беременности?",
	},
	{Key: KeyParentCaries, Column: "Был ли кариес у ваших родителей?", Label: "Был ли кариес у ваших родителей?"},
	{
		Key:    KeySalivaPH,
		Column: "рН вашей слюны  (по результатам анализа)",
		Label:  "рН вашей слюны (по результатам анализа)",
	},
}

// OutputFields defines the four predicted values (Excel column headers).
var OutputFields = []Field{
	{
		Key:    KeyChildCariesProbability,
		Column: "Вероятность развития кариеса у ребенка",
		Label:  "Вероятность развития кариеса у ребенка",
	},
	{Key: KeyRiskGroup, Column: "Группа риска", Label: "Группа риска"},
	{Key: KeyAction, Column: "Действие", Label: "Действие"},
	{Key: KeyRecommendations, Column: "Назначение (Рекомендации)", Label: "Назначение (Рекомендации)"},
}

// InputKeys returns stable API keys for required request fields.
func InputKeys() []string {
	keys := make([]string, len(InputFields))
	for i, f := range InputFields {
		keys[i] = f.Key
	}
	return keys
}

// OutputKeys returns stable API keys for response fields.
func OutputKeys() []string {
	keys := make([]string, len(OutputFields))
	for i, f := range OutputFields {
		keys[i] = f.Key
	}
	return keys
}

// ColumnToInputKey maps normalized Excel header to input API key.
func ColumnToInputKey(column string) (string, bool) {
	norm := normalizeColumn(column)
	for _, f := range InputFields {
		if normalizeColumn(f.Column) == norm {
			return f.Key, true
		}
	}
	return "", false
}

// ColumnToOutputKey maps normalized Excel header to output API key.
func ColumnToOutputKey(column string) (string, bool) {
	norm := normalizeColumn(column)
	for _, f := range OutputFields {
		if normalizeColumn(f.Column) == norm {
			return f.Key, true
		}
	}
	return "", false
}

func normalizeColumn(column string) string {
	parts := strings.Fields(strings.TrimSpace(column))
	return strings.Join(parts, " ")
}
