package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func heartbeatHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_ = json.NewEncoder(w).Encode(heartbeatResponse{
		Status: "OK",
		Code:   http.StatusOK,
	})
}

// Middleware should be able to handle panic
func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				//Middleware should be able to handle panic
				logrus.WithFields(logrus.Fields{
					"URL": r.RequestURI,
				}).Errorf("panic: %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func handler() http.Handler {
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.Use(recoverHandler)
	router.HandleFunc("/", heartbeatHandlerFunc)
	router.HandleFunc("/add/a/{first}/b/{second}", addHandlerFunc)
	router.HandleFunc("fetch/sum", fetchSumFunc)
	router.HandleFunc("/todo", todoHandlerFunc)
	router.HandleFunc("/sqs/sendmessage", sendSQSMessageFunc)
	router.HandleFunc("/sqs/message", listSQSMessageFunc)
	return router
}

func main() {
	query := "CREATE TABLE IF NOT EXISTS to_do_item(id INT AUTO_INCREMENT PRIMARY KEY, description VARCHAR(255) NOT NULL, dueDate DATETIME NOT NULL);"
	_, err := mysql.Exec(query)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":   err,
			"query": query,
		}).Error("Unable to create table to_do_item")
	} else {
		logrus.WithFields(logrus.Fields{
			"err":   err,
			"query": query,
		}).Info("Created table to_do_item")
	}
	// Getting SQS queue name in order to start reading message
	logrus.Info("Starting CLI")
	queueName := "my-queue"
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	}
	urlResult, err := sqsClient.GetQueueUrl(ctx, gQInput)
	if err != nil {
		logrus.Error("Got an error getting the queue URL:", err)
	}
	queueURL := urlResult.QueueUrl

	// running it in goroutine so that http server could start up
	go func() {
		// Insert a new TodoItem every 3 seconds
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		sqsReadChan := make(chan struct{}, 1)
		for {
			select {
			case <-ticker.C:
				go func() {
					todoItem := TodoItem{
						Description: "Sample Todo Item",
						DueDate:     time.Now().Add(24 * time.Hour), // Set due date to 24 hours from now
					}
					todoItem.DueDateStr = todoItem.DueDate.Format("2006-01-02 15:04:05")
					// Insert the TodoItem into the database
					insertStm, err := mysql.Prepare("INSERT INTO to_do_item(description, dueDate) VALUES (?, ?)")
					if err != nil {
						logrus.Error("mysql prepare statement error", err)
					}
					defer insertStm.Close()

					_, err = insertStm.Exec(todoItem.Description, todoItem.DueDate)
					if err != nil {
						logrus.Error("mysql prepare statement execution error", err)
					}

					logrus.WithFields(logrus.Fields{
						"description": todoItem.Description,
						"dueDate":     todoItem.DueDate,
					}).Info("Inserted into mysql to_do_item table")

					// Insert into SQS queue - my-queue
					byteArr, err := json.Marshal(todoItem)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"err":      err,
							"todoItem": todoItem,
						}).Error("Error in json marshalling toDoItem")
					}
					todoItemStr := string(byteArr)
					sendSQSMessage(todoItemStr)
					sqsReadChan <- struct{}{}
				}()
			case <-sqsReadChan:
				go func() {
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
					}
					for _, message := range msgResult.Messages {
						var todoItem TodoItem
						if err := json.Unmarshal([]byte(*message.Body), &todoItem); err != nil {
							logrus.Errorf("failed to unmarshal message body: %v", err)
							continue
						}
						logrus.WithFields(logrus.Fields{
							"messageId":   *message.MessageId,
							"description": todoItem.Description,
							"dueDate":     todoItem.DueDateStr,
						}).Info("SQS receive message")

						_, err := sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
							QueueUrl:      queueURL,
							ReceiptHandle: message.ReceiptHandle,
						})
						if err != nil {
							logrus.Errorf("failed to delete message: %v", err)
						}
					}
				}()
			}
		}
	}()

	logrus.Info("starting http web server started at port 3001")
	errHttp := http.ListenAndServe(":3001", handler())
	if errHttp != nil {
		logrus.WithFields(logrus.Fields{
			"addr": "3001",
			"err":  errHttp,
		}).Fatal("Unable to create HTTP Service ")
	}
}
