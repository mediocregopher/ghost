package main

import (
	"github.com/mediocregopher/ghost"
	"log"
	"time"
)

var SERVER = "localhost:4000"

type Hello struct {
	A string
	B int
}

func main() {
	ghost.Register(Hello{})
	if err := ghost.AddConn(&SERVER); err != nil {
		log.Fatal(err)
	}
	for {
		time.Sleep(1 * time.Second)

		log.Println("Sending to server")
		msg := Hello{"This is A", 2}
		err := ghost.Send(&SERVER, msg)
		if err != nil {
			log.Println("ERR", err)
		}
	}
}
