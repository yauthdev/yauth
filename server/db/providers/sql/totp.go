package sql

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	"github.com/authorizerdev/authorizer/server/db/models"
)

func (p *provider) UpsertAuthenticator(ctx context.Context, authenticators models.Authenticators) (*models.Authenticators, error) {
	if authenticators.ID == "" {
		authenticators.ID = uuid.New().String()
	}
	authenticators.Key = authenticators.ID
	authenticators.CreatedAt = time.Now().Unix()
	authenticators.UpdatedAt = time.Now().Unix()
	res := p.db.Clauses(
		clause.OnConflict{
			UpdateAll: true,
			Columns:   []clause.Column{{Name: "id"}},
		}).Create(&authenticators)
	if res.Error != nil {
		return nil, res.Error
	}
	return &authenticators, nil
}

func (p *provider) GetAuthenticatorDetailsByUserId(ctx context.Context, userId string, authenticatorType string) (*models.Authenticators, error) {
	var authenticators models.Authenticators
	result := p.db.Where("user_id = ?", userId).Where("method = ?", authenticatorType).First(&authenticators)
	if result.Error != nil {
		return nil, result.Error
	}
	return &authenticators, nil
}
