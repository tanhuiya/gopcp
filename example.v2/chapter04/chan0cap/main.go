package main

import (
	"log"
	"time"
)

func main(){
	sendInterval := time.Second
	receiveInterval := time.Second * 2
	intChan := make(chan int, 0)
	go func() {
		for i := 0; i <= 5; i++ {
			intChan <- i
			time.Sleep(sendInterval)
		}
		close(intChan)
	}()
	var ts1, ts0 int64
LOOP:
	for {
		select {
		case value, ok := <- intChan:
			if !ok {
				break LOOP
			}
			ts1 = time.Now().Unix()
			if ts0 == 0 {
				log.Println("received: ", value)
			} else {
				log.Printf("received: %d interval %d", value, ts1 - ts0)
			}
		}
		ts0 = ts1
		time.Sleep(receiveInterval)
	}
	log.Print("End")
}
