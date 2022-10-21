package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

type Response struct {
	Message string `json:"message"`
}

func HandleRequest(ctx context.Context) (*Response, error) {
	log.Println("Testing Lambda Debug")
	return &Response{
		Message: "Hello, world!",
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
