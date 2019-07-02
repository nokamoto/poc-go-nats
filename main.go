package main

import (
	"flag"
	"fmt"
	nats "github.com/nats-io/nats.go"
	"log"
	"time"
)

var host = flag.String("host", "nats", "nats host")

func main() {
	flag.Parse()

	log.Println("hello")

	nc, err := nats.Connect(fmt.Sprintf("nats://%s:4222", *host), nats.Name("poc-go-nats"), nats.PingInterval(3*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Subscribe
	sub, err := nc.SubscribeSync("updates")
	if err != nil {
		log.Fatal(err)
	}

	// Publish
	if err := nc.Publish("updates", []byte("All is Well")); err != nil {
		log.Fatal(err)
	}

	// Wait for a message
	msg, err := sub.NextMsg(10 * time.Second)
	if err != nil {
		log.Fatal(err)
	}

	// Use the response
	log.Printf("Reply: %s", msg.Data)
}
