package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

/*
CREATE TABLE test_ravi (

	id INT AUTO_INCREMENT PRIMARY KEY,
	sum DECIMAL(10, 2),  -- Adjust the precision and scale as needed
	date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP

);
*/
func todoHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if mysql == nil {
		mysql = InitMysql()
	}
	rows, err := mysql.Query("select id,description,dueDate from to_do_item order by dueDate desc limit 1;")
	if err != nil {
		logrus.Error("mysql query statement error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		response := make(map[string]string)
		response["error"] = err.Error()
		_ = json.NewEncoder(w).Encode(response)
	}
	defer rows.Close()
	var results []TodoItem
	for rows.Next() {
		var tr TodoItem
		err := rows.Scan(&tr.ID, &tr.Description, &tr.DueDateStr)
		if err != nil {
			logrus.Error("Error scanning row: ", err)
			continue
		}
		results = append(results, tr)
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error": err,
		}).Error("Error json marshal response")
	}
}

func addHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	a, aErr := strconv.Atoi(vars["first"])
	b, bErr := strconv.Atoi(vars["second"])
	if aErr != nil || bErr != nil {
		log.Println(a, aErr, b, bErr)
		w.WriteHeader(http.StatusBadRequest)
		response := make(map[string]string)
		response["error"] = "Invalid Request"
		_ = json.NewEncoder(w).Encode(response)
	}
	sum := a + b
	if mysql == nil {
		mysql = InitMysql()
	}
	insertStm, err := mysql.Prepare("insert into test_ravi(sum) values (?);")
	if err != nil {
		logrus.Error("mysql prepare statement error", err)
	}
	defer insertStm.Close()
	_, err = insertStm.Exec(sum)
	if err != nil {
		logrus.Error("mysql prepare statement execution error", err)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(sumResponse{Sum: sum})
}

func fetchSumFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if mysql == nil {
		mysql = InitMysql()
	}
	rows, err := mysql.Query("select * from test_ravi;")
	if err != nil {
		logrus.Error("mysql query statement error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		response := make(map[string]string)
		response["error"] = err.Error()
		_ = json.NewEncoder(w).Encode(response)
	}
	defer rows.Close()
	var results []TestRavi
	for rows.Next() {
		var tr TestRavi
		err := rows.Scan(&tr.ID, &tr.Sum, &tr.DateCreated)
		if err != nil {
			logrus.Error("Error scanning row: ", err)
			continue
		}
		results = append(results, tr)
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error": err,
		}).Error("Error json marshal response")
	}
}

func listSQSMessageFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	messageMap := listSQSQueueMessage()
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(messageMap)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error": err,
		}).Error("Error marshall response")
	}
}

func sendSQSMessageFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var msgReq RequestMessage
	if err := json.Unmarshal(body, &msgReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	messageId := sendSQSMessage(msgReq.Message)
	w.WriteHeader(http.StatusOK)
	response := make(map[string]string)
	response["message_id"] = messageId
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error": err,
		}).Error("Error marshall response")
	}
}
