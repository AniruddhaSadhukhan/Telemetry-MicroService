package query

import (
	"fmt"
	"go-telemetry-server/common"
	"go-telemetry-server/gqlhandler/schema"
	"go-telemetry-server/logger"
	"go-telemetry-server/models"
	"go-telemetry-server/telemetry"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var TelemetryEventQuery = &graphql.Field{
	Name: "telemetryEvents",
	Type: schema.JSON,
	Description: `Returns the result of the MongoDB Aggregation Pipeline provided as input.

	As new Date(...) operator is not supported in JSON, use the following syntax to create a date object:
	"new Date([<Date RFC3339 String>])[.AddDate(<Year>, <Month>, <Day>)][.AddDuration(<Duration String>)]"

	Example:
	1. "new Date()" : Will give current date time
	2. "new Date(2020-01-01T15:30:00Z)" : Will convert the given date string to date time
	3. "new Date().AddDate(0, -1, 0)" : Will return the previous month date time
	4. "new Date().AddDuration(-5m)" : Will return the date time 5 minutes ago
	5. "new Date().AddDate(0, -1, 0).AddDuration(-5m)" : Will return the date time 1 month 5 minutes ago
	6. "new Date(2020-01-01T15:30:00Z).AddDate(0, -1, 0).AddDuration(-5m)" : Will return the date time 1 month 5 minutes ago of the given date string

	Note: Some of the MongoDB stages are restricted due to security reasons.


	
	`,
	Args: graphql.FieldConfigArgument{
		"aggregationPipeline": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(schema.JSON),
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

		defer telemetry.LogGraphQlCall(p, e)

		_, err = resolveDate(p.Args["aggregationPipeline"])
		if err != nil {
			return nil, err
		}

		var aggregationPipeline []primitive.M
		err = mapstructure.Decode(p.Args["aggregationPipeline"], &aggregationPipeline)
		if err != nil {
			logger.Log.Errorf("Error in decoding aggregation pipeline input")
			return nil, err
		}

		err = checkForRestrictedOperations(aggregationPipeline)
		if err != nil {
			return nil, err
		}

		res := make([]primitive.M, 0)

		err = models.Aggregate(
			p.Context,
			models.TelemetryCollection,
			aggregationPipeline,
			&res,
		)

		if err != nil {
			return nil, err
		}

		return res, nil

	},
}

func resolveDate(rawInput interface{}) (interface{}, error) {

	input := reflect.ValueOf(rawInput)

	switch input.Kind() {

	// Recursively check all values of the map
	case reflect.Map:
		for _, key := range input.MapKeys() {

			newValue, err := resolveDate(input.MapIndex(key).Interface())
			if err != nil {
				return rawInput, err
			}
			input.SetMapIndex(key, reflect.ValueOf(newValue))
		}

	// Check all elements of the array
	case reflect.Slice:
		for index := 0; index < input.Len(); index++ {

			newValue, err := resolveDate(input.Index(index).Interface())
			if err != nil {
				return rawInput, err
			}
			input.Index(index).Set(reflect.ValueOf(newValue))
		}

	// Check for string and resolve date values
	case reflect.String:
		rawInputString := rawInput.(string)
		if strings.HasPrefix(rawInputString, "new Date(") {
			date, err := convertCustomStringToDate(rawInputString)
			if err != nil {
				return rawInput, err
			}
			rawInput = date
		}
	}

	return rawInput, nil

}

func convertCustomStringToDate(rawInputString string) (date time.Time, err error) {

	rawInputString = strings.TrimPrefix(rawInputString, "new Date(")

	dateStringSegments := strings.SplitN(rawInputString, ")", 2)
	dateString := strings.TrimSpace(dateStringSegments[0])
	additionalDateString := strings.TrimSpace(dateStringSegments[1])

	if dateString == "" {
		date = time.Now()
	} else {
		date, err = time.Parse(time.RFC3339, dateString)
		if err != nil {
			logger.Log.Errorf("Error parsing date string: %v", err)
			return date, fmt.Errorf("unable to parse date : date must be in RFC3339 format ( 2006-01-02T15:04:05Z07:00 )")
		}
	}

	if additionalDateString == "" {
		return
	}

	additionalOperationSegments := strings.Split(additionalDateString, ".")
	for _, operationString := range additionalOperationSegments {

		operationString = strings.TrimSpace(operationString)

		if operationString == "" {
			continue
		}

		if strings.HasPrefix(operationString, "AddDuration(") {
			operationString = strings.TrimPrefix(operationString, "AddDuration(")
			operationString = strings.TrimSuffix(operationString, ")")
			operationString = strings.TrimSpace(operationString)

			parsedDuration, err := time.ParseDuration(operationString)
			if err != nil {
				logger.Log.Errorf("Error parsing duration string: %v", err)
				return date, fmt.Errorf(`unable to parse duration for AddDuration : A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". `)
			}

			date = date.Add(parsedDuration)

		} else if strings.HasPrefix(operationString, "AddDate(") {

			operationString = strings.TrimPrefix(operationString, "AddDate(")
			operationString = strings.TrimSuffix(operationString, ")")
			operationString = strings.TrimSpace(operationString)

			dateParams := strings.SplitN(operationString, ",", 3)

			years, err := strconv.Atoi(strings.TrimSpace(dateParams[0]))
			if err != nil {
				logger.Log.Errorf("Error parsing year date params: %v", err)
				return date, fmt.Errorf("unable to parse years for AddDate : %v", err)
			}

			months, err := strconv.Atoi(strings.TrimSpace(dateParams[1]))
			if err != nil {
				logger.Log.Errorf("Error parsing month date params: %v", err)
				return date, fmt.Errorf("unable to parse months for AddDate : %v", err)
			}

			days, err := strconv.Atoi(strings.TrimSpace(dateParams[2]))
			if err != nil {
				logger.Log.Errorf("Error parsing day date params: %v", err)
				return date, fmt.Errorf("unable to parse days for AddDate : %v", err)
			}

			date = date.AddDate(years, months, days)

		} else {
			return date, fmt.Errorf("got invalid operation : %v . Supported operations are AddDate, AddDuration", operationString)
		}
	}

	return
}

var restrictedOperations = map[string]bool{
	"$lookup":            true,
	"$unionWith":         true,
	"$out":               true,
	"$merge":             true,
	"$changeStream":      true,
	"$callStats":         true,
	"$currentOp":         true,
	"$graphLookup":       true,
	"$indexStats":        true,
	"$listLocalSessions": true,
	"$listSessions":      true,
	"$planCacheStats":    true,
}

func checkForRestrictedOperations(pipeline []primitive.M) error {
	for _, stage := range pipeline {
		for key := range stage {
			if restrictedOperations[key] {
				return fmt.Errorf("operation %v is restricted", key)
			}
		}
	}
	return nil
}
