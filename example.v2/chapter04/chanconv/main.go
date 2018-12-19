package main

import "fmt"

func main()  {
	var ok bool
	ch := make(chan int)
	_, ok = interface{}(ch).(<-chan int)
	fmt.Println("chan to <-chan int ", ok)
	_, ok = interface{}(ch).(chan<- int)
	fmt.Println("chan to chan<- int ", ok)

	sch := make(chan<- int)
	_, ok = interface{}(sch).(chan int)
	fmt.Println("chan<- to chan int ", ok)

	rch := make(<-chan int)
	_, ok = interface{}(rch).(chan int)
	fmt.Println("<-chan to chan int ", ok)
}