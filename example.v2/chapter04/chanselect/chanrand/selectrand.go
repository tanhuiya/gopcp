package main

import "fmt"

func main()  {
	chanNum := 5
	chanInt := make(chan int, chanNum)
	for i := 0; i < chanNum; i++  {
		select {
		case chanInt<- 0:
		case chanInt<- 1:
		case chanInt<- 2:
		}
	}
	for i := 0; i < chanNum; i++  {
		fmt.Println(<-chanInt)
	}
}
