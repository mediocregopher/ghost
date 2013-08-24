package main

import (
	"github.com/mediocregopher/ghost"
	"log"
)

type Hello struct {
	A string
	B int
}

func main() {
	ghost.Register(Hello{})

	rcvCh,errCh,err := ghost.Listen(":4000")
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	log.Println("Listening")

	go func(){
		for err := range errCh {
			log.Println(err.Error())
		}
	}()

	go func(){
		for msg := range rcvCh {
			msgW := (*msg).(Hello)
			log.Printf("Got Hello message: %v",msgW)
		}
	}()

	select {}
}
