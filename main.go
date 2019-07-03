package main

import (
	"flag"
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"sync"
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
	log.Println()
	log.Println("https://nats-io.github.io/docs/developer/concepts/subjects.html")
	log.Println("time.us -> nats-server -> us1(time.us), us2(time.us)")
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

	log.Println()

	log.Println("time.us.east -> nats-server -> usEast(time.us.east), usAnyEast(time.*.east)")

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

	log.Println()

	log.Println("time.us.east.atlanta -> nats-server -> usAnyPlus(time.us.>)")
	log.Println("                                     x usAny(time.us.*)")

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

func requestReply(nc *nats.Conn) {
	log.Println()
	log.Println("https://nats-io.github.io/docs/developer/concepts/reqreply.html")
	log.Println("publisher -> subject -> subscriber -> reply -> publisher")

	request, err := nc.SubscribeSync("subject")
	if err != nil {
		log.Fatal(err)
	}

	reply, err := nc.SubscribeSync("reply")
	if err != nil {
		log.Fatal(err)
	}

	// Send the request
	err = nc.PublishRequest("subject", "reply", []byte("All is Well"))
	if err != nil {
		log.Fatal(err)
	}

	// Wait for a single request
	msg, err := request.NextMsg(10 * time.Second)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s: %s", request.Subject, string(msg.Data))

	// Send the response
	err = msg.Respond([]byte(fmt.Sprintf("echo %s", string(msg.Data))))
	if err != nil {
		log.Fatal(err)
	}

	// Wait for a single response
	if err := wait(reply); err != nil {
		log.Fatal(err)
	}
}

func queueSubscribersScalability(nc *nats.Conn) {
	log.Println()
	log.Println("https://nats-io.github.io/docs/developer/concepts/queue.html")

	size := 10

	wg := sync.WaitGroup{}
	wg.Add(size)

	_, err := nc.QueueSubscribe("updates", "worker", func(m *nats.Msg) {
		log.Printf("0: updates(worker): %s", string(m.Data))
		wg.Done()
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = nc.QueueSubscribe("updates", "worker", func(m *nats.Msg) {
		log.Printf("1: updates(worker): %s", string(m.Data))
		wg.Done()
	})
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < size; i++ {
		if err := nc.Publish("updates", []byte(fmt.Sprintf("%d: All is Well", i))); err != nil {
			log.Fatal(err)
		}
	}

	wg.Wait()
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

	requestReply(nc)

	queueSubscribersScalability(nc)
}
