# DSS EMAILING
Project contains two executables: Client and server

### Client.go
Listen http requests, write request on queue broker to server and read responses from 
responses broker. 

### Server.go
Listen RequestQueue, process request and write response in responses queue

### Functions

* POST method to create requests to server per user, BODY => users - JSONString
```
http://localhost:9001/email
```

### Installing

To run the program, please install previous requirement:

* Install Go compiler for official web site: https://golang.org/
* Install RabbitMQ for official web site: https://www.rabbitmq.com/

### Install GoLand(Optional - Recomendation)
* Install GoLand: https://www.jetbrains.com/go/

### If you have installed GoLand, Do not follow these steps:
* After clone repository, inside the project folder, open a command prompt and type:
```
go get -u github.com/golang/dep/cmd/dep
```
```
dep init -v
```
```
dep ensure -v 
```
### Required
* Execute the server:
```
go run server.go
```

* Execute the client:
```
go run client.go
```
