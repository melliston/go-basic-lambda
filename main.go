package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/google/uuid"
)

const tableName = "shop_ping_log" // TODO Load from .env etc

const (
	ERR_DEVICE_ID = "error device id required"
)

// dependencies to inject into the handle request so we can mock for testing.
type dependencies struct {
	ddb            dynamodbiface.DynamoDBAPI // The dynamodb
	tableName      string
	sessionOptions *session.Options // Dependancy so we can create and test with bad options
}

// Data type to send to the dynamodb table complete with tags to map to columns
type PingLog struct {
	Id        uuid.UUID `dynamodbav:"id"`
	Device    string    `dynamodbav:"device"`
	Timestamp time.Time `dynamodbav:"timestamp"`
	Synced    bool      `dynamodbav:"synced"`
}

// A struct to hold the params passed with the request. The json tag will automatically parse passed in json from the post body.
type DeviceRequest struct {
	Device string `json:"device"`
}

func (d *dependencies) HandleRequest(ctx context.Context, device DeviceRequest) (events.APIGatewayProxyResponse, error) {
	if len(device.Device) < 1 {
		return events.APIGatewayProxyResponse{}, errors.New(ERR_DEVICE_ID)
	}

	// Create db connection here - could go in main but apparently not much more overhead to do hear and will always be in the request then.
	if d.ddb == nil {
		// setup the session
		sess, err := session.NewSessionWithOptions(*d.sessionOptions)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}
		// Init the dynamodb with the provided session
		d.ddb = dynamodb.New(sess)
		d.tableName = tableName
	}

	// create a new log entry
	id := uuid.New()
	logEntry := PingLog{
		Id:        id,
		Device:    device.Device,
		Timestamp: time.Now(),
		Synced:    false,
	}

	// Marshal our logEntry to a dynamodb object
	item, err := dynamodbattribute.MarshalMap(logEntry)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	// save to the DB
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(d.tableName),
	}

	response, err := d.ddb.PutItem(input)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	// encode the response
	body, err := json.Marshal(response.Attributes)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	// Return a 200 if we get to here it was a success
	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 200,
	}, nil
}

func main() {
	d := dependencies{
		ddb:            nil,
		tableName:      tableName,
		sessionOptions: &session.Options{SharedConfigState: session.SharedConfigEnable},
	}
	lambda.Start(d.HandleRequest)
}
