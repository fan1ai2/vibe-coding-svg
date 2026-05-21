package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/service"
	"github.com/hibiken/asynq"
)

type ConversionWorker struct {
	cfg     *config.Config
	repo    *repo.ConversionRepo
	storage *service.Storage
}

func NewConversionWorker(cfg *config.Config, r *repo.ConversionRepo, s *service.Storage) *ConversionWorker {
	return &ConversionWorker{cfg: cfg, repo: r, storage: s}
}

func (w *ConversionWorker) HandleProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload service.ConversionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if err := w.repo.UpdateStatus(payload.ConversionID, model.StatusProcessing, ""); err != nil {
		return fmt.Errorf("update status to processing: %w", err)
	}

	reader, err := w.storage.Download(service.BucketOriginals, payload.OriginalKey)
	if err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "download failed: "+err.Error())
		return fmt.Errorf("download original: %w", err)
	}
	defer reader.Close()

	tmpDir := os.TempDir()
	inPath := filepath.Join(tmpDir, payload.ConversionID+"_in."+payload.FormatIn)
	outPath := filepath.Join(tmpDir, payload.ConversionID+"_out.svg")

	inFile, err := os.Create(inPath)
	if err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "temp file: "+err.Error())
		return fmt.Errorf("create temp input: %w", err)
	}
	if _, err := io.Copy(inFile, reader); err != nil {
		inFile.Close()
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "write temp: "+err.Error())
		return fmt.Errorf("write temp input: %w", err)
	}
	inFile.Close()
	defer os.Remove(inPath)
	defer os.Remove(outPath)

	if err := ConvertRasterToSVG(inPath, outPath); err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "conversion failed: "+err.Error())
		return fmt.Errorf("convert: %w", err)
	}

	svgData, err := os.ReadFile(outPath)
	if err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "read result: "+err.Error())
		return fmt.Errorf("read svg result: %w", err)
	}

	resultKey := payload.OriginalKey + ".svg"
	resultFile, err := os.Open(outPath)
	if err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "open result: "+err.Error())
		return fmt.Errorf("open svg file: %w", err)
	}
	defer resultFile.Close()

	fi, _ := resultFile.Stat()
	if err := w.storage.Upload(service.BucketResults, resultKey, "image/svg+xml", resultFile, fi.Size()); err != nil {
		w.repo.UpdateStatus(payload.ConversionID, model.StatusFailed, "upload result: "+err.Error())
		return fmt.Errorf("upload svg result: %w", err)
	}

	pathCount := CountSVGPaths(svgData)
	fileSizeOut := len(svgData)

	if err := w.repo.UpdateResult(payload.ConversionID, resultKey, "", fileSizeOut, pathCount, 0); err != nil {
		return fmt.Errorf("update result in db: %w", err)
	}

	return nil
}
