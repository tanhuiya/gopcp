package main

import (
	"fmt"
	"time"
)

type Counter struct {
	count int
}
func main() {
	var mapChan = make(chan map[string]Counter, 1)
	syncChan := make(chan struct{}, 2)
	go func() {
		for {
			if ele, ok := <-mapChan; ok {
				counter := ele["count"]
				counter.count++
			} else {
				break
			}
		}
		fmt.Println("stop recieve")
		syncChan<- struct{}{}
	}()

	go func() {
		countMap := map[string]Counter{
			"count": Counter{},
		}
		for i := 0; i < 5; i++ {
			mapChan <- countMap
			time.Sleep(time.Second)
			fmt.Println("the count map: %v, sender",countMap)
		}
		close(mapChan)
		syncChan<- struct{}{}
	}()
	<- syncChan
	<- syncChan
}

