package schema

import "github.com/graphql-go/graphql"

var TelemetryEventInputSchema = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "TelemetryEventInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"uniqueId": &graphql.InputObjectFieldConfig{
				Type:        graphql.ID,
				Description: "Provide a unique ID string (preferably UUID v4) which can be used for duplication check. If the ID matches any previous ID, this event will be considered a duplicate and discarded. In case no ID is provided, they will not be considered in the duplication checking ie. they will be given new unique UUID v4 IDs.",
			},
			"timeStamp": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.DateTime),
			},
			"eventType": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(TelemetryEventTypeEnum),
			},
			"user": &graphql.InputObjectFieldConfig{
				Type: graphql.NewInputObject(graphql.InputObjectConfig{
					Name: "UserInfo",
					Fields: graphql.InputObjectConfigFieldMap{
						"id": &graphql.InputObjectFieldConfig{
							Type:        graphql.String,
							Description: "ID of the actor. For Example: UID, Email, etc. in case of an user",
						},
						"type": &graphql.InputObjectFieldConfig{
							Type:        graphql.String,
							Description: "User, System, Internal Service, etc",
						},
					},
				}),
			},
			"platform": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.NewInputObject(graphql.InputObjectConfig{
					Name: "PlatformInfo",
					Fields: graphql.InputObjectConfigFieldMap{
						"platformName": &graphql.InputObjectFieldConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
						"env": &graphql.InputObjectFieldConfig{
							Type: graphql.String,
						},
						"componentName": &graphql.InputObjectFieldConfig{
							Type: graphql.String,
						},
					},
				})),
			},
			"device": &graphql.InputObjectFieldConfig{
				Type: graphql.NewInputObject(graphql.InputObjectConfig{
					Name: "DeviceInfo",
					Fields: graphql.InputObjectConfigFieldMap{
						"browserVersion": &graphql.InputObjectFieldConfig{
							Type: graphql.String,
						},
						"deviceType": &graphql.InputObjectFieldConfig{
							Type: graphql.String,
						},
						"orientation": &graphql.InputObjectFieldConfig{
							Type: graphql.String,
						},
						"os": &graphql.InputObjectFieldConfig{
							Type: graphql.String,
						},
						"browser": &graphql.InputObjectFieldConfig{
							Type: graphql.String,
						},
						"device": &graphql.InputObjectFieldConfig{
							Type: graphql.String,
						},
						"osVersion": &graphql.InputObjectFieldConfig{
							Type: graphql.String,
						},
					},
				}),
			},
			"eventData": &graphql.InputObjectFieldConfig{
				Type: JSON,
			},
		},
	},
)

var TelemetryEventTypeEnum = graphql.NewEnum(
	graphql.EnumConfig{
		Name: "TelemetryEventTypeEnum",
		Values: graphql.EnumValueConfigMap{
			"START": &graphql.EnumValueConfig{
				Value:       "Start",
				Description: "This indicates the user logged in or started using the application",
			},
			"END": &graphql.EnumValueConfig{
				Value:       "End",
				Description: "This indicates the user logged out or closed the application",
			},
			"PAGE_VISIT": &graphql.EnumValueConfig{
				Value:       "PageVisit",
				Description: "This is used to capture user visits to a specific page",
			},
			"FEATURE_USE": &graphql.EnumValueConfig{
				Value:       "FeatureUse",
				Description: "This is used to capture user usage of a specific feature",
			},
			"SEARCH": &graphql.EnumValueConfig{
				Value:       "Search",
				Description: "This is used to capture user search activity",
			},
			"BACKEND_GRAPHQL_CALL": &graphql.EnumValueConfig{
				Value:       "BackendGraphQLCall",
				Description: "This is used to capture backend GraphQL calls",
			},
			"ERROR": &graphql.EnumValueConfig{
				Value:       "Error",
				Description: "This is used to capture errors",
			},
			"OTHER": &graphql.EnumValueConfig{
				Value:       "Other",
				Description: "This is used to capture other events which can not be categorized in existing categories",
			},
		},
	},
)
