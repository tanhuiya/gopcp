package main

import "fmt"

func main()  {
	intChan := make(chan int, 10)
	for i := 0; i < 10; i++ {
		intChan <- i
	}
	close(intChan)
	syncChan := make(chan struct{})
	go func() {
		LOOP:
		for {
			select {
			case val, ok := <-intChan:
				if !ok {
					fmt.Print("End")
					break LOOP
				} else {
					fmt.Println(val)
				}
			}
		}
		syncChan<- struct{}{}
	}()
	<-syncChan
}
