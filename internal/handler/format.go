package handler

import (
	"fmt"
	"strings"

	"github.com/anuarkuanysh/dental_project/internal/model"
)

const telegramMaxMessageLen = 4090

// FormatTelegramReply builds a plain-text reply for Telegram (no parse mode), на русском языке.
func FormatTelegramReply(a *model.Analysis) string {
	var b strings.Builder

	b.WriteString("Только справочная информация — не медицинская и не стоматологическая консультация.\n\n")
	b.WriteString("Наблюдения могут быть ошибочными; одного фото недостаточно для оценки.\n\n")

	label, hint := trafficLightBlockRU(a.TrafficLight)
	b.WriteString(fmt.Sprintf("%s\n", label))
	if summary := strings.TrimSpace(a.TrafficLightSummary); summary != "" {
		b.WriteString(fmt.Sprintf("%s\n", summary))
	}
	b.WriteString(fmt.Sprintf("%s\n\n", hint))

	b.WriteString("Возможные заметные особенности (без диагноза, с оговорками):\n")
	if len(a.VisibleIssues) == 0 {
		b.WriteString("- надёжно не выявлено\n")
	} else {
		for _, item := range a.VisibleIssues {
			b.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}

	conf := confidenceLabelRU(a.Confidence)
	b.WriteString(fmt.Sprintf("\nУверенность модели / качества снимка: %s\n", conf))

	b.WriteString("\nОбщие советы по гигиене (не план лечения):\n")
	if len(a.Recommendations) == 0 {
		b.WriteString("- поддерживайте регулярную чистку зубов и использование нити; персональные рекомендации — у врача\n")
	} else {
		for _, item := range a.Recommendations {
			b.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}

	disc := strings.TrimSpace(a.Disclaimer)
	if disc == "" {
		disc = "Это не медицинская консультация."
	}
	b.WriteString(fmt.Sprintf("\n%s\n", disc))
	b.WriteString("\nПри симптомах, боли, отёке или тревоге за здоровье обратитесь к стоматологу.")

	s := b.String()
	if len(s) > telegramMaxMessageLen {
		s = s[:telegramMaxMessageLen] + "\n…(сообщение обрезано)"
	}
	return s
}

func trafficLightBlockRU(level string) (headline string, hint string) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "green":
		return "🟢 Уровень: зелёный (справочно)",
			"На снимке не видно явных поводов для беспокойства; продолжайте гигиену и плановые визиты к стоматологу."
	case "red":
		return "🔴 Уровень: красный (справочно)",
			"Рекомендуется записаться к стоматологу в ближайшее время. При сильной боли, кровотечении или отёке — срочная помощь."
	default:
		return "🟡 Уровень: жёлтый (справочно)",
			"Имеет смысл показать снимок стоматологу на плановом приёме и следить за гигиеной."
	}
}

func confidenceLabelRU(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "low":
		return "низкая"
	case "medium":
		return "средняя"
	case "high":
		return "высокая"
	default:
		return "неизвестно"
	}
}
