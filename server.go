package main

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"os"
)

var TYPE_READ_ALL_DOCUMENTS = 0
var TYPE_SAVE_DOCUMENT = 1
var TYPE_DELETE_DOCUMENT = 2

type Users struct {
	ID    string
	Name  string
	Email string
}

type Protocol struct {
	Name     string
	Content string
	Status   bool
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}


func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	storageRequestqueue, err := ch.QueueDeclare(
		"EmailingRequest", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		storageRequestqueue.Name, // queue
		"",                       // consumer
		false,                    // auto-ack
		false,                    // exclusive
		false,                    // no-local
		false,                    // no-wait
		nil,                      // args
	)
	failOnError(err, "Failed to register a consumer")

	storageResponsesqueue, err := ch.QueueDeclare(
		"EmailingResponses", // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare a queue")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			res := d.Body
			var messageFromQueue Protocol
			json.Unmarshal(res, &messageFromQueue)
			requestInJson, err := json.Marshal(messageFromQueue)
			failOnError(err, "Failed to publish a message")
			if !processUsersRequest(messageFromQueue) {
				messageFromQueue.Status = false
			}
			requestInJson, err = json.Marshal(messageFromQueue)
			failOnError(err, "Failed to publish a message")
			d.Ack(false)
			err = ch.Publish(
				"",                         // exchange
				storageResponsesqueue.Name, // routing key
				false,                      // mandatory
				false,                      // immediate
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          requestInJson,
				})
		}
	}()
	log.Printf(" [*] Awaiting RPC requests")
	<-forever
}


func processUsersRequest(messageFromQueue Protocol) bool {
	listOfUsers := messageFromQueue.Content
	var users []Users
	json.Unmarshal([]byte(listOfUsers), &users)

	for _, user := range users {
		err := createFileByRequest(user.Name)
		if err != nil {
			return false
		}
	}
	return true
}


func createFileByRequest(userName string) error {
	fileContent := "Request email from " + userName
	fileName := "./Requests/"+userName+".txt"

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		error := ioutil.WriteFile(fileName, []byte(fileContent), 777)
		return error
	} else {
		return nil
	}

}
