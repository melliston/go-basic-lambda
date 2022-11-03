package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func TestLambdaHandler(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {

		d := dependencies{
			tableName:      tableName,
			marshalJson:    marshalJson,
			marshalMap:     marshalMap,
			createSession:  createAwsSession,
			sessionOptions: session.Options{Config: aws.Config{}},
		}

		mockedDevice := DeviceRequest{
			Device: "abc123",
		}
		_, err := d.HandleRequest(context.TODO(), mockedDevice)
		if err != nil {
			t.Logf("FAILED: testing HandleRequest failed: %s\n", err)
		}
	})
}

func TestLambdaHandlerNoDevice(t *testing.T) {
	t.Run("device not present request", func(t *testing.T) {

		d := dependencies{}
		mockedDevice := DeviceRequest{}
		_, err := d.HandleRequest(context.TODO(), mockedDevice)

		if err.Error() != ErrDeviceID.Error() {
			t.Errorf("FAILED: testing HandleRequest failed with no device id present: %s\n", err)
		}
	})
}

func TestLambdaHandlerFailedAWSSession(t *testing.T) {
	t.Run("incorrect aws session setup request", func(t *testing.T) {

		d := dependencies{
			sessionOptions: session.Options{Config: aws.Config{Region: aws.String("us-fake-region-2")}}, // Empty session with no login details
			marshalJson:    marshalJson,
			marshalMap:     marshalMap,
			createSession:  mock_createSession,
		}

		// Create a new context and pass to HandleRequest
		_, err := d.createSession(d.sessionOptions)

		if err == nil {
			t.Errorf("FAILED: expexted error, got none\n")
		} else {
			if err == ErrAwsSessionFailed {
				t.Logf("PASSED: creation of failed error\n")
			} else {
				t.Errorf("FAILED: expected error %v, got %v\n", ErrAwsSessionFailed, err)
			}
		}
	})
}

func TestDynamoDBMarshalling(t *testing.T) {

	t.Run("dynamodb marshalling", func(t *testing.T) {
		d := dependencies{
			marshalMap: marshalMap,
		}
		x := "foo"
		want, _ := dynamodbattribute.MarshalMap(x)
		got, err := d.marshalMap(x)
		if err != nil {
			t.Errorf("FAILED: expected no error, error got %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("FAILED: marshalling got %v, want %v", got, want)
		}
	})

	t.Run("dynamodb marshalling error", func(t *testing.T) {
		d := dependencies{
			marshalMap: mock_marshalMap,
		}
		x := "foo"
		_, err := d.marshalMap(x)
		if err == nil {
			t.Errorf("FAILED: marshilling error should have occured with data passed\n")
		} else {
			if err != ErrAwsDynamoDBMarshalling {
				t.Errorf("FAILED: expected error %v, got %v\n", ErrAwsDynamoDBMarshalling, err)
			} else {
				t.Logf("PASSED: dynamodb marshalling error")
			}
		}
	})
}

func TestJsonMarshalling(t *testing.T) {

	d := dependencies{
		marshalJson: marshalJson,
	}

	t.Run("json marshalling", func(t *testing.T) {
		x := "foo"
		want := "\"" + x + "\""
		got, err := d.marshalJson(x)
		if err != nil {
			t.Errorf("FAILED: expected no error, error got %v", err)
		}
		if string(got) != want {
			t.Errorf("FAILED: json marshalling got %v, want %v", got, want)
		}
	})

	t.Run("json marshalling error", func(t *testing.T) {
		x := map[string]interface{}{
			"foo": make(chan int),
		}
		_, err := d.marshalJson(x)
		if err == nil {
			t.Errorf("FAILED: expected error got nil")
		}
	})
}

func mock_marshalMap(input interface{}) (map[string]*dynamodb.AttributeValue, error) {
	return nil, ErrAwsDynamoDBMarshalling
}

func mock_createSession(input session.Options) (*session.Session, error) {
	return nil, ErrAwsSessionFailed
}
