// Command generate-prediction-examples writes a sample Excel file for few-shot prediction.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/xuri/excelize/v2"

	domain "github.com/anuarkuanysh/dental_project/internal/domain/prediction"
)

func main() {
	out := "data/prediction_examples.xlsx"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		log.Fatalf("create dir: %v", err)
	}

	headers := make([]string, 0, len(domain.InputFields)+len(domain.OutputFields))
	for _, field := range domain.InputFields {
		headers = append(headers, field.Column)
	}
	for _, field := range domain.OutputFields {
		headers = append(headers, field.Column)
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			log.Fatalf("set header: %v", err)
		}
	}

	rows := [][]string{
		{"28", "20", "8", "1", "нет", "да", "6.2", "65%", "высокая", "срочный осмотр", "профессиональная гигиена и фторирование"},
		{"30", "24", "5", "2", "да", "да", "6.5", "45%", "средняя", "наблюдение", "контроль через 3 месяца"},
		{"26", "16", "3", "2", "да", "нет", "7.0", "25%", "низкая", "профилактика", "поддерживающая гигиена"},
		{"34", "32", "10", "1", "нет", "да", "6.0", "70%", "высокая", "лечение кариеса", "санация полости рта до родов"},
		{"29", "22", "4", "3", "да", "нет", "6.9", "30%", "низкая", "профилактика", "рекомендованы фторсодержащие пасты"},
	}

	for r, row := range rows {
		for c, value := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				log.Fatalf("set cell: %v", err)
			}
		}
	}

	if err := f.SaveAs(out); err != nil {
		log.Fatalf("save workbook: %v", err)
	}

	fmt.Printf("wrote %s with %d examples\n", out, len(rows))
}
