package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"tsserver/m/v2/client"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var newClientChannel = make(chan *websocket.Conn)

func createMongoClient() *mongo.Client {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable.")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().
		ApplyURI(uri))

	if err != nil {
		panic(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err == nil {
		return client
	} else {
		panic(err)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade this http request to a websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	newClientChannel <- conn
}

func handleNewTsClient(ctx context.Context, db *mongo.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case newClient := <-newClientChannel:
			client := client.CreateTsClient(newClient, db)
			client.Start(ctx)
		}
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	var db_client = createMongoClient()

	// Create database connection
	result, _ := db_client.ListDatabaseNames(context.TODO(), bson.M{})
	fmt.Println(result)

	// Direct http requests to have WS clients created
	http.HandleFunc("/wsclient", wsHandler)
	go handleNewTsClient(context.TODO(), db_client)

	// Create http server
	const port = "9000"
	fmt.Printf("Creating http server on port: %s\n", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
