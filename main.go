package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
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

type JsonMarsaller func(input interface{}) ([]byte, error)
type MapMarsaller func(input interface{}) (map[string]*dynamodb.AttributeValue, error)
type SessionGenerator func(options session.Options) (*session.Session, error)

const tableName = "shop_ping_log" // TODO Load from .env etc

var ErrJsonMarshalling = errors.New("encountered an error marshalling json data")
var ErrAwsSessionFailed = errors.New("failed to create aws session")
var ErrAwsDynamoDBMarshalling = errors.New("dynamodb encountered an error marshalling data")
var ErrDeviceID = errors.New("error device id required")
var ErrAwsSessionConfigIncorrect = errors.New("aws session configuration incorrect please check logs")

// dependencies to inject into the handle request so we can mock for testing.
type dependencies struct {
	ddb            dynamodbiface.DynamoDBAPI // The dynamodb
	tableName      string
	sessionOptions session.Options // Dependancy so we can create and test with bad options
	marshalJson    JsonMarsaller
	marshalMap     MapMarsaller
	createSession  SessionGenerator
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
		return events.APIGatewayProxyResponse{}, ErrDeviceID
	}

	// Create db connection here - could go in main but apparently not much more overhead to do hear and will always be in the request then.
	if d.ddb == nil {
		// setup the session
		sess, err := d.createSession(d.sessionOptions)
		if err != nil {
			return events.APIGatewayProxyResponse{}, ErrAwsSessionFailed
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
	item, err := marshalMap(logEntry) //dynamodbattribute.MarshalMap(logEntry)
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
	body, err := d.marshalJson(response.Attributes)
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
		marshalJson:    marshalJson,
		marshalMap:     marshalMap,
		createSession:  createAwsSession,
		sessionOptions: session.Options{SharedConfigState: session.SharedConfigEnable},
	}
	lambda.Start(d.HandleRequest)
}

func createAwsSession(options session.Options) (*session.Session, error) {
	// setup the session
	sess, err := session.NewSessionWithOptions(options)
	if err != nil {
		return nil, ErrAwsSessionFailed
	}

	return sess, nil
}

func marshalMap(input interface{}) (map[string]*dynamodb.AttributeValue, error) {
	// Marshal our logEntry to a dynamodb object
	item, err := dynamodbattribute.MarshalMap(input)
	if err != nil {
		return nil, ErrAwsDynamoDBMarshalling
	}
	return item, nil
}

func marshalJson(input interface{}) ([]byte, error) {
	output, err := json.Marshal(input)
	if err != nil {
		log.Printf("marshalling json error: %+v\n", err)
		return nil, ErrJsonMarshalling
	}
	return output, nil
}
