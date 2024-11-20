package main

type heartbeatResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
}

type sumResponse struct {
	Sum int `json:"sum"`
}

type TestRavi struct {
	ID          int     `json:"id"`
	Sum         float64 `json:"sum"`
	DateCreated string  `json:"date_created"`
}

type SQSMessage struct {
	MessageId string `json:"message_id"`
	Body      string `json:"body"`
}

type RequestMessage struct {
	Message string `json:"message"`
}