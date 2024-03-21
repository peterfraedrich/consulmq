package main

import (
	"fmt"

	"github.com/peterfraedrich/consulmq"
)

func main() {

	mq, err := consulmq.Connect(consulmq.Config{
		Address:    "172.17.0.2:8500",
		Datacenter: "dc1",
		Token:      "",
		MQName:     "cmq",
	})
	if err != nil {
		panic(err)
	}

	i := 0
	for i <= 100 {
		// Put and item on the queue
		qo, err := mq.Push([]byte("Hello, is it me you're looking for?"))
		if err != nil {
			panic(err)
		}
		fmt.Println(qo.ID)
		i++
	}
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	x := 0
	for x <= 100 {
		// Pop an item off the queue
		_, qo, err := mq.Pop()
		if err != nil {
			panic(err)
		}
		fmt.Println(qo.ID)
		x++
	}
}
