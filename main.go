package main

import (
	"flag"
	"fmt"
	nats "github.com/nats-io/nats.go"
	"log"
	"time"
)

var host = flag.String("host", "nats", "nats host")

func wait(sub *nats.Subscription) error {
	msg, err := sub.NextMsg(10 * time.Second)
	if err != nil {
		return err
	}

	log.Printf("%s: %s", sub.Subject, string(msg.Data))

	return nil
}

func subjectBasedMessaging(nc *nats.Conn) {
	// time.us -> nats-server -> us1(time.us), us2(time.us)
	//
	// Subscribe
	us1, err := nc.SubscribeSync("time.us")
	if err != nil {
		log.Fatal(err)
	}

	us2, err := nc.SubscribeSync("time.us")
	if err != nil {
		log.Fatal(err)
	}

	// Publish
	if err := nc.Publish("time.us", []byte("All is Well")); err != nil {
		log.Fatal(err)
	}

	// Wait for a message
	if err := wait(us1); err != nil {
		log.Fatal(err)
	}
	if err := wait(us2); err != nil {
		log.Fatal(err)
	}

	// time.us.east -> nats-server -> usEast(time.us.east), usAnyEast(time.*.east)
	//
	// Subscribe
	usEast, err := nc.SubscribeSync("time.us.east")
	if err != nil {
		log.Fatal(err)
	}

	usAnyEast, err := nc.SubscribeSync("time.*.east")
	if err != nil {
		log.Fatal(err)
	}

	// Publish
	if err := nc.Publish("time.us.east", []byte("All is Well!")); err != nil {
		log.Fatal(err)
	}

	// Wait for a message
	if err := wait(usEast); err != nil {
		log.Fatal(err)
	}
	if err := wait(usAnyEast); err != nil {
		log.Fatal(err)
	}

	// time.us.east.atlanta -> nats-server -> usAnyPlus(time.us.>)
	//                                      x usAny(time.us.*)
	// Subscribe
	usAnyPlus, err := nc.SubscribeSync("time.us.>")
	if err != nil {
		log.Fatal(err)
	}

	usAny, err := nc.SubscribeSync("time.us.*")
	if err != nil {
		log.Fatal(err)
	}

	// Publish
	if err := nc.Publish("time.us.east.atlanta", []byte("All is Well!!")); err != nil {
		log.Fatal(err)
	}

	// Wait for a message
	if err := wait(usAnyPlus); err != nil {
		log.Fatal(err)
	}
	if err := wait(usAny); err == nil {
		log.Fatal("unreachable")
	}
}

func main() {
	flag.Parse()

	log.Println("hello")

	nc, err := nats.Connect(fmt.Sprintf("nats://%s:4222", *host), nats.Name("poc-go-nats"), nats.PingInterval(3*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	subjectBasedMessaging(nc)
}
