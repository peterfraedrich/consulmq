package main

import (
	"fmt"

	"github.com/peterfraedrich/kvmq"
)

func main() {

	mq, err := kvmq.NewMQ(&kvmq.Config{
		Backend: "memory",
	})
	if err != nil {
		panic(err)
	}

	i := 0
	for i <= 25 {
		// Put and item on the queue
		qo, err := mq.Push([]byte("Hello, is it me you're looking for?"))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(qo.ID)
		i++
	}
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	x := 0
	for x <= 25 {
		// Pop an item off the queue
		_, qo, err := mq.Pop()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(qo.ID)
		x++
	}
}
