package services

import (
    "errors"
    "my-api/models"
    "my-api/repositories"
)

type AssetService struct {
    repo *repositories.AssetRepository
}

type CreateAssetDTO struct {
    Name     string  `json:"name"`
    Type     string  `json:"type"`
    Balance  float64 `json:"balance"`
    Currency string  `json:"currency"`
    BankName string  `json:"bank_name"`
    AccountNo string `json:"account_no"`
}

type UpdateAssetDTO struct {
    Name      *string  `json:"name"`
    Type      *string  `json:"type"`
    Balance   *float64 `json:"balance"`
    Currency  *string  `json:"currency"`
    BankName  *string  `json:"bank_name"`
    AccountNo *string  `json:"account_no"`
}

func (dto *CreateAssetDTO) validate() error {
    if dto.Name == "" {
        return errors.New("name is required")
    }
    if dto.Currency == "" {
        return errors.New("currency is required")
    }
    if dto.Balance < 0 {
        return errors.New("balance cannot be negative")
    }
    return nil
}

func NewAssetService(repo *repositories.AssetRepository) *AssetService {
    return &AssetService{repo: repo}
}

func (s *AssetService) CreateAsset(userID uint64, dto CreateAssetDTO) (*models.Asset, error) {
    if err := dto.validate(); err != nil {
        return nil, err
    }
    asset := &models.Asset{
        UserID:    userID,
        Name:      dto.Name,
        Type:      dto.Type,
        Balance:   dto.Balance,
        Currency:  dto.Currency,
        BankName:  dto.BankName,
        AccountNo: dto.AccountNo,
    }
    if err := s.repo.CreateAsset(asset); err != nil {
        return nil, err
    }
    return asset, nil
}

func (s *AssetService) ListAssets(userID uint64) ([]models.Asset, error) {
    return s.repo.GetAssetsByUser(userID)
}

func (s *AssetService) GetAsset(userID uint64, id uint64) (*models.Asset, error) {
    asset, err := s.repo.GetAssetByID(id)
    if err != nil {
        return nil, err
    }
    if asset.UserID != userID {
        return nil, errors.New("unauthorized")
    }
    return asset, nil
}

func (s *AssetService) UpdateAsset(userID uint64, id uint64, dto UpdateAssetDTO) (*models.Asset, error) {
    asset, err := s.repo.GetAssetByID(id)
    if err != nil {
        return nil, err
    }
    if asset.UserID != userID {
        return nil, errors.New("unauthorized")
    }
    if dto.Name != nil { asset.Name = *dto.Name }
    if dto.Type != nil { asset.Type = *dto.Type }
    if dto.Balance != nil {
        if *dto.Balance < 0 { return nil, errors.New("balance cannot be negative") }
        asset.Balance = *dto.Balance
    }
    if dto.Currency != nil { asset.Currency = *dto.Currency }
    if dto.BankName != nil { asset.BankName = *dto.BankName }
    if dto.AccountNo != nil { asset.AccountNo = *dto.AccountNo }
    if err := s.repo.UpdateAsset(asset); err != nil {
        return nil, err
    }
    return asset, nil
}

func (s *AssetService) DeleteAsset(userID uint64, id uint64) error {
    asset, err := s.repo.GetAssetByID(id)
    if err != nil {
        return err
    }
    if asset.UserID != userID {
        return errors.New("unauthorized")
    }
    return s.repo.DeleteAsset(id)
}

func (s *AssetService) Summary(userID uint64) (map[string]float64, error) {
    assets, err := s.ListAssets(userID)
    if err != nil {
        return nil, err
    }
    summary := make(map[string]float64)
    for _, a := range assets {
        summary[a.Currency] += a.Balance
    }
    return summary, nil
}
