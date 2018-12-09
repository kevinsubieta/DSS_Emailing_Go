package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"github.com/streadway/amqp"
)

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

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/documents", createRequestToWriteFile).Methods("POST")
	log.Fatal(http.ListenAndServe(":9001", router))

}

func openConnectionToQueue() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	return ch
}

func writeRequestOnQueue(messageRequest []byte) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	channel, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer channel.Close()

	storageRequestqueue, err := channel.QueueDeclare(
		"EmailingRequest", // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = channel.Publish(
		"",                       // exchange
		storageRequestqueue.Name, // routing key
		false,                    // mandatory
		false,                    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageRequest,
		})
	failOnError(err, "Failed to publish a message")
}

func returnValueToUser(w http.ResponseWriter, message Protocol) {
	if message.Status {
		http.Error(w, "Request OK", http.StatusOK)
	} else {
		http.Error(w, "Error on Request", http.StatusBadRequest)
	}
}

func createRequestToWriteFile(w http.ResponseWriter, r *http.Request) {
	users := r.FormValue("users")
	var usersRequest = Protocol{"EmailRequest", users, true}
	requestInJson, _ := json.Marshal(usersRequest)
	writeRequestOnQueue(requestInJson)
	readValuesFromQueue(w)
}

func readValuesFromQueue(w http.ResponseWriter) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	channel, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer channel.Close()

	q, err := channel.QueueDeclare(
		"EmailingResponses", // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments               // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	failOnError(err, "Failed to set QoS"+q.Name)

	msgs, err := channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	for d := range msgs {
		//if CORR_ID == d.CorrelationId {
		res := d.Body
		var messageFromQueue Protocol
		json.Unmarshal(res, &messageFromQueue)
		returnValueToUser(w, messageFromQueue)
		d.Ack(false)
		break
		//}
	}
	log.Printf("Reading queue from responses")
}


