package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/anuarkuanysh/dental_project/infra/config"
	exceladapter "github.com/anuarkuanysh/dental_project/internal/adapters/driven/excel"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

func providePredictionExamples(cfg config.Config, log *slog.Logger) (port.PredictionExampleRepository, error) {
	path := resolvePredictionExamplesPath(cfg.PredictionExamplesPath)
	repo, err := exceladapter.NewRepository(path)
	if err != nil {
		if isPredictionFileMissing(err) {
			log.Warn("prediction examples file not found; POST /predict will return unavailable",
				"path", path,
				"configured", cfg.PredictionExamplesPath,
			)
			return exceladapter.NewEmptyRepository(), nil
		}
		return nil, err
	}

	examples, err := repo.ListExamples(context.Background())
	if err != nil {
		return nil, err
	}
	log.Info("prediction examples loaded", "path", path, "count", len(examples))
	return repo, nil
}

func resolvePredictionExamplesPath(cfgPath string) string {
	candidates := []string{cfgPath}
	if filepath.IsAbs(cfgPath) {
		return firstExistingPath(candidates, cfgPath)
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, cfgPath),
			filepath.Join(exeDir, "data", "prediction_examples.xlsx"),
		)
	}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, cfgPath),
			filepath.Join(wd, "data", "prediction_examples.xlsx"),
		)
	}

	return firstExistingPath(candidates, cfgPath)
}

func firstExistingPath(candidates []string, fallback string) string {
	seen := make(map[string]struct{}, len(candidates))
	for _, p := range candidates {
		if p == "" {
			continue
		}
		clean := filepath.Clean(p)
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		if _, err := os.Stat(clean); err == nil {
			return clean
		}
	}
	return fallback
}

func isPredictionFileMissing(err error) bool {
	if errors.Is(err, os.ErrNotExist) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "no such file or directory") ||
		strings.Contains(msg, "not found")
}
