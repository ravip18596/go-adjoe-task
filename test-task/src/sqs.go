package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/sirupsen/logrus"
)

var ctx = context.Background()

func InitAWSSQS() *sqs.Client {
	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "http://localstack:4566",
			SigningRegion: "eu-central-1",
		}, nil
	})

	sdkConfig, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("eu-central-1"),
		config.WithEndpointResolver(customResolver),
	)
	if err != nil {
		logrus.Error("Couldn't load default configuration. Have you set up your AWS account?")
		logrus.Error(err)
		return nil
	}

	creds := credentials.NewStaticCredentialsProvider("test", "test", "")
	sdkConfig.Credentials = creds
	sqsClient := sqs.NewFromConfig(sdkConfig)
	return sqsClient
}

var sqsClient = InitAWSSQS()

func sendSQSMessage(message string) string {
	queueName := "my-queue"
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	}
	urlResult, err := sqsClient.GetQueueUrl(ctx, gQInput)
	if err != nil {
		logrus.Error("Got an error getting the queue URL:", err)
		return ""
	}
	queueURL := urlResult.QueueUrl

	gMInput := &sqs.SendMessageInput{
		QueueUrl:    queueURL,
		MessageBody: aws.String(message),
	}
	output, err := sqsClient.SendMessage(ctx, gMInput)
	if err != nil {
		logrus.Error("Got an error sending messages:")
		logrus.Error(err)
		return ""
	}
	logrus.Info("Message sent:", *output.MessageId)
	return *output.MessageId
}

func listSQSQueueMessage() []SQSMessage {
	queueName := "my-queue"
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	}
	message := make([]SQSMessage, 0)
	urlResult, err := sqsClient.GetQueueUrl(ctx, gQInput)
	if err != nil {
		logrus.Error("Got an error getting the queue URL:", err)
		return message
	}
	queueURL := urlResult.QueueUrl

	gMInput := &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            queueURL,
		MaxNumberOfMessages: 1,
		VisibilityTimeout:   int32(12 * 60 * 60),
	}

	msgResult, err := sqsClient.ReceiveMessage(ctx, gMInput)
	if err != nil {
		logrus.Error("Got an error receiving messages:")
		logrus.Error(err)
		return message
	}
	for _, msg := range msgResult.Messages {
		message = append(message, SQSMessage{
			Body:      *msg.Body,
			MessageId: *msg.MessageId,
		})
	}
	return message
}
