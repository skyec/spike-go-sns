package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type TestEvent struct {
	Name    string
	Action  int64
	Message string
}

func main() {
	svc := sns.New(session.New(), &aws.Config{Region: aws.String("us-west-2")})

	event := TestEvent{
		Name:    "Bob",
		Action:  42,
		Message: "This is a test",
	}

	jsn, err := json.Marshal(event)
	if err != nil {
		log.Fatalf("json.Marshal: %s", err)
	}

	params := &sns.PublishInput{
		Message: aws.String(base64.StdEncoding.EncodeToString(jsn)), // Required
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"EventType": { // Required
				DataType:    aws.String("String"), // Required
				StringValue: aws.String("TestEvent"),
			},
		},

		Subject:  aws.String("subject"),
		TopicArn: aws.String("arn:aws:sns:us-west-2:874194573056:a1zone-builder"),
	}
	resp, err := svc.Publish(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}
