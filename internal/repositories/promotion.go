package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"order/internal/models"
	pgGorm "order/internal/repositories/pg-gorm"
)

type PromotionRepository struct {
	db pgGorm.PGInterface
}

func NewPromotionRepository(newPgRepo pgGorm.PGInterface) *PromotionRepository {
	return &PromotionRepository{db: newPgRepo}
}

// --- Interface for DI (add methods used by service) ---
type PromotionRepoInterface interface {
	GetActivePromotion(ctx context.Context, at time.Time) (*models.PromotionConfig, error)
	HasCustomerReceived(ctx context.Context, promoID uuid.UUID, customerID uuid.UUID) (bool, error)
	CountDistinctCustomers(ctx context.Context, promoID uuid.UUID) (int64, error)
	CountRewards(ctx context.Context, promoID uuid.UUID) (int64, error)
	CreateReward(ctx context.Context, reward *models.PromotionReward) error
}

// implementations

func (r *PromotionRepository) GetActivePromotion(ctx context.Context, at time.Time) (*models.PromotionConfig, error) {
	tx, cancel := r.db.DBWithTimeout(ctx)
	defer cancel()
	var promo models.PromotionConfig
	if err := tx.Where("start_time <= ? AND end_time >= ? AND is_active = ?", at, at, true).Order("start_time desc").First(&promo).Error; err != nil {
		return nil, err
	}
	return &promo, nil
}

func (r *PromotionRepository) HasCustomerReceived(ctx context.Context, promoID uuid.UUID, customerID uuid.UUID) (bool, error) {
	tx, cancel := r.db.DBWithTimeout(ctx)
	defer cancel()
	var count int64
	if err := tx.Model(&models.PromotionReward{}).
		Where("promotion_config_id = ? AND customer_id = ?", promoID, customerID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PromotionRepository) CountDistinctCustomers(ctx context.Context, promoID uuid.UUID) (int64, error) {
	tx, cancel := r.db.DBWithTimeout(ctx)
	defer cancel()
	var count int64
	// count distinct customers who already received reward for this promotion
	if err := tx.Model(&models.PromotionReward{}).
		Select("customer_id").
		Where("promotion_config_id = ?", promoID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PromotionRepository) CountRewards(ctx context.Context, promoID uuid.UUID) (int64, error) {
	tx, cancel := r.db.DBWithTimeout(ctx)
	defer cancel()
	var count int64
	if err := tx.Model(&models.PromotionReward{}).
		Where("promotion_config_id = ?", promoID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PromotionRepository) CreateReward(ctx context.Context, reward *models.PromotionReward) error {
	var cancel context.CancelFunc
	tx, cancel := r.db.DBWithTimeout(ctx)
	defer cancel()
	if err := tx.Create(reward).Error; err != nil {
		return err
	}
	return nil
}
