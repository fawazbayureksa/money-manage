package repositories

import (
    "errors"
    "gorm.io/gorm"
    "my-api/models"
)

type AssetRepository struct {
    DB *gorm.DB
}

func NewAssetRepository(db *gorm.DB) *AssetRepository {
    return &AssetRepository{DB: db}
}

func (r *AssetRepository) CreateAsset(asset *models.Asset) error {
    return r.DB.Create(asset).Error
}

func (r *AssetRepository) GetAssetsByUser(userID uint64) ([]models.Asset, error) {
    var assets []models.Asset
    if err := r.DB.Where("user_id = ?", userID).Find(&assets).Error; err != nil {
        return nil, err
    }
    return assets, nil
}

func (r *AssetRepository) GetAssetByID(id uint64) (*models.Asset, error) {
    var asset models.Asset
    if err := r.DB.First(&asset, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, err
        }
        return nil, err
    }
    return &asset, nil
}

func (r *AssetRepository) UpdateAsset(asset *models.Asset) error {
    return r.DB.Save(asset).Error
}

func (r *AssetRepository) DeleteAsset(id uint64) error {
    return r.DB.Delete(&models.Asset{}, id).Error
}
