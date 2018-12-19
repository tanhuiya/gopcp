package main

import (
	"fmt"
	"time"
)

func main() {
	var mapChan = make(chan map[string]int, 1)
	syncChan := make(chan struct{}, 2)
	go func() {
		for {
			if ele, ok := <-mapChan; ok {
				ele["count"]++
			} else {
				break
			}
		}
		fmt.Println("stop recieve")
		syncChan<- struct{}{}
	}()

	go func() {
		countMap := make(map[string]int)
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
