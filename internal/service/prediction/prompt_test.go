package prediction_test

import (
	"strings"
	"testing"

	domain "github.com/anuarkuanysh/dental_project/internal/domain/prediction"
	promptsvc "github.com/anuarkuanysh/dental_project/internal/service/prediction"
)

func TestBuildFewShotPrompt_includesExamplesAndTarget(t *testing.T) {
	t.Parallel()

	examples := []domain.Example{
		{
			Input: domain.Row{
				domain.KeyAge:            "28",
				domain.KeyPregnancyWeeks: "20",
				domain.KeyKPUIndex:       "5",
				domain.KeyBrushingPerDay: "2",
				domain.KeyDentistVisitDuringPregnancy: "да",
				domain.KeyParentCaries:                "да",
				domain.KeySalivaPH:                  "6.8",
			},
			Expected: domain.Row{
				domain.KeyChildCariesProbability: "45%",
				domain.KeyRiskGroup:              "средняя",
				domain.KeyAction:                 "наблюдение",
				domain.KeyRecommendations:        "контроль через 3 месяца",
			},
		},
	}
	target := domain.Row{
		domain.KeyAge:            "30",
		domain.KeyPregnancyWeeks: "24",
		domain.KeyKPUIndex:       "3",
		domain.KeyBrushingPerDay: "2",
		domain.KeyDentistVisitDuringPregnancy: "нет",
		domain.KeyParentCaries:                "нет",
		domain.KeySalivaPH:                  "7.0",
	}

	prompt := promptsvc.BuildFewShotPrompt(examples, target)

	for _, want := range []string{
		"Example 1:",
		"Ваш возраст (полных лет)?: 28",
		"Вероятность развития кариеса у ребенка: 45%",
		"Now predict for:",
		"Ваш возраст (полных лет)?: 30",
		"JSON:",
		"child_caries_probability",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q\n%s", want, prompt)
		}
	}
}

func TestParseOutputJSON(t *testing.T) {
	t.Parallel()

	raw := `{"child_caries_probability":"30%","risk_group":"низкая","action":"профилактика","recommendations":"фторирование"}`
	out, err := promptsvc.ParseOutputJSON(raw)
	if err != nil {
		t.Fatalf("ParseOutputJSON: %v", err)
	}
	if out[domain.KeyRiskGroup] != "низкая" {
		t.Fatalf("unexpected risk group: %q", out[domain.KeyRiskGroup])
	}
}
