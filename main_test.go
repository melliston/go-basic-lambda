package main

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockedPutItem struct {
	dynamodbiface.DynamoDBAPI
	Response dynamodb.PutItemOutput
}

func (d mockedPutItem) PutItem(in *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return &d.Response, nil
}

func TestLambdaHandler(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {

		m := mockedPutItem{
			Response: dynamodb.PutItemOutput{},
		}

		d := dependencies{
			ddb:       m,
			tableName: tableName,
		}

		mockedDevice := DeviceRequest{
			Device: "abc123",
		}
		_, err := d.HandleRequest(context.TODO(), mockedDevice)
		if err != nil {
			t.Fatalf("testing HandleRequest failed: %s", err)
		}
	})
}

func TestLambdaHandlerNoDevice(t *testing.T) {
	t.Run("device not present request", func(t *testing.T) {

		d := dependencies{}

		mockedDevice := DeviceRequest{}

		_, err := d.HandleRequest(context.TODO(), mockedDevice)

		if err.Error() != ERR_DEVICE_ID {
			t.Fatalf("testing HandleRequest failed with no device id present: %s", err)
		}
	})
}

func TestLambdaHandlerAWSSession(t *testing.T) {
	t.Run("incorrect aws session setup request", func(t *testing.T) {

		d := dependencies{
			sessionOptions: &session.Options{}, // Empty session with no login details
		}

		mockedDevice := DeviceRequest{
			Device: "foobar123",
		}

		// Create a new context and pass to HandleRequest
		_, err := d.HandleRequest(context.TODO(), mockedDevice)
		// TODO Handle this test result
		if err != nil {
			//fmt.Println(err)
		}
	})
}
