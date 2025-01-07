package graphql

import (
	"context"
	"time"

	"github.com/authorizerdev/authorizer/internal/constants"
	"github.com/authorizerdev/authorizer/internal/graph/model"
	"github.com/authorizerdev/authorizer/internal/utils"
)

// DeactivateAccount is the method for the deactivate_account field.
// Permissions: authorized user
func (g *graphqlProvider) DeactivateAccount(ctx context.Context) (*model.Response, error) {
	log := g.Log.With().Str("func", "DeactivateAccount").Logger()
	var res *model.Response
	gc, err := utils.GinContextFromContext(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get GinContext")
		return res, err
	}

	tokenData, err := g.TokenProvider.GetUserIDFromSessionOrAccessToken(gc)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get user id from session or access token")
		return res, err
	}
	log = log.With().Str("userID", tokenData.UserID).Logger()
	user, err := g.StorageProvider.GetUserByID(ctx, tokenData.UserID)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get user by id")
		return res, err
	}
	now := time.Now().Unix()
	user.RevokedTimestamp = &now
	user, err = g.StorageProvider.UpdateUser(ctx, user)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to update user")
		return res, err
	}
	go func() {
		g.MemoryStoreProvider.DeleteAllUserSessions(user.ID)
		g.EventsProvider.RegisterEvent(ctx, constants.UserDeactivatedWebhookEvent, "", user)
	}()
	res = &model.Response{
		Message: `user account deactivated successfully`,
	}
	return res, nil
}