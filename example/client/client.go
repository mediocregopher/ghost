package main

import (
	"github.com/mediocregopher/ghost"
	"log"
	"time"
)

const SERVER = "localhost:4000"

type Hello struct {
	A string
	B int
}

func main() {
	ghost.Register(Hello{})
	ghost.AddConn(SERVER)
	for {
		time.Sleep(1 * time.Second)

		log.Println("Sending to server")
		msg := Hello{"This is A", 2}
		err := ghost.Send(SERVER, msg)
		if err != nil {
			log.Println("ERR", err)
		}
	}
}
