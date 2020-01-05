package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"os"
)

func main() {
	lambda.Start(interactiveMessageHandler)
}

func interactiveMessageHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	res := events.APIGatewayProxyResponse{}

	signingSecrets := os.Getenv("SIGNING_SECRETS")
	// Use interfaces instead of structures for loose coupling
	interactiveMessageUsecase := NewInteractionUsecase(signingSecrets)
	// Create a response for user's action
	res, err := interactiveMessageUsecase.MakeSlackResponse(req)
	if err != nil {
		return res, err
	}
	return res, nil
}