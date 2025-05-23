package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	model "kimistore/internal/models"
	"kimistore/internal/repo"
	pgGorm "kimistore/internal/repo/pg-gorm"
	"os"
)

type MediaService struct {
	repo      repo.MediaRepositoryInterface
	newPgRepo pgGorm.PGInterface
}

type MediaServiceInterface interface {
	SaveMedia(ctx context.Context, media model.Media) (*model.ImageSaveResponse, error)
	IsDuplicate(ctx context.Context, hash string) (bool, error)
	GenerateFileHash(filePath string) (string, error)
}

func NewMediaService(repo repo.MediaRepositoryInterface, newRepo pgGorm.PGInterface) *MediaService {
	return &MediaService{
		repo:      repo,
		newPgRepo: newRepo,
	}
}

func (s *MediaService) GenerateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *MediaService) SaveMedia(ctx context.Context, media model.Media) (*model.ImageSaveResponse, error) {

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()

	saveMedia, err := s.repo.SaveMedia(media, tx, ctx)
	if err != nil {
		return nil, err
	}

	tx.Commit()

	return saveMedia, nil
}

func (s *MediaService) IsDuplicate(ctx context.Context, hash string) (bool, error) {

	tx := s.newPgRepo.GetRepo().Begin()
	defer tx.Rollback()

	tx, cancel := s.newPgRepo.DBWithTimeout(ctx)
	defer cancel()
	return s.repo.IsDuplicate(hash, tx)
}
