package mutation

import (
	"context"
	"go-telemetry-server/common"
	"go-telemetry-server/gqlhandler/schema"
	"go-telemetry-server/logger"
	"go-telemetry-server/models"

	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var AddTelemetryEventMutation = &graphql.Field{
	Name:        "addTelemetryEvents",
	Type:        graphql.Boolean,
	Description: "Add telemetry events",
	Args: graphql.FieldConfigArgument{
		"events": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.NewList(schema.TelemetryEventInputSchema)),
		},
	},
	Resolve: func(p graphql.ResolveParams) (i interface{}, e error) {

		if !common.IsValidUser(p) {
			return nil, common.ErrUnauthorized
		}

		_, err := common.Sanitize(p.Args)
		if err != nil {
			return nil, err
		}

		logger.Log.Info("Mutation: Add Telemetry Events called by " + common.GetUserName(p))

		var telemetryEvents []models.TelemetryEvent

		err = mapstructure.Decode(p.Args["events"], &telemetryEvents)
		if err != nil {
			logger.Log.Errorf("Error in decoding telemetry events input: %v", err)
		}

		user := models.UserInfo{
			ID:   common.GetUserName(p),
			Type: common.GetUserType(p),
		}

		for index, event := range telemetryEvents {
			if event.User.ID == "" || !common.IsInternalUser(p) {
				telemetryEvents[index].User = user
			}
			if event.ID == "" {
				telemetryEvents[index].ID = uuid.NewString()
			}
		}

		// Create []interface{} from telemetryEvents
		var telemetryEventsDocs []interface{}
		err = mapstructure.Decode(telemetryEvents, &telemetryEventsDocs)
		if err != nil {
			logger.Log.Errorf("Error in decoding telemetry events: %v", err)
			return nil, err
		}

		//Insert user
		err = models.InsertMany(
			context.Background(), // Note: Don't use p.Context here as in case of tab close, the call will be cancelled
			models.TelemetryCollection,
			telemetryEventsDocs,
			options.InsertMany().SetOrdered(false), // Needed so that all new documents can be inserted in case some of them are duplicates
		)

		if err != nil && !mongo.IsDuplicateKeyError(err) {
			logger.Log.Errorf("Error inserting telemetry events: %v", err)
			return nil, err
		}
		return true, nil
	},
}
