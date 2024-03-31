package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/peterfraedrich/kvmq"
)

func main() {

	mq, err := kvmq.NewMQ(&kvmq.Config{
		Backend: "filesystem",
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("PUSH")
	obj, err := mq.Push([]byte("This is a push"))
	printResult(obj, err)
	fmt.Println("PUSH_FIRST")
	obj, err = mq.PushFirst([]byte("Push First"))
	printResult(obj, err)
	fmt.Println("PUSH_INDEX 1")
	obj, err = mq.PushIndex([]byte("This is a push index at 1"), 1)
	printResult(obj, err)
	fmt.Println("LENGTH")
	l, _ := mq.Length()
	fmt.Println(l)

	fmt.Println("PEEK")
	_, obj, err = mq.Peek()
	printResult(obj, err)
	fmt.Println("PEEK_INDEX 1")
	_, obj, err = mq.PeekIndex(1)
	printResult(obj, err)
	fmt.Println("PEEK_LAST")
	_, obj, err = mq.PeekLast()
	printResult(obj, err)
	fmt.Println("FIND 'First'")
	_, _, obj, err = mq.Find([]byte("First"))
	printResult(obj, err)
	fmt.Println("PEEK_SCAN")
	_, objs, _ := mq.PeekScan()
	b, _ := json.MarshalIndent(objs, "", " ")
	fmt.Println(string(b[:]))

	fmt.Println("POP")
	_, obj, err = mq.Pop()
	printResult(obj, err)
	fmt.Println("POP_INDEX 1")
	_, obj, err = mq.PopIndex(1)
	printResult(obj, err)
	fmt.Println("POP_LAST")
	_, obj, err = mq.PopLast()
	printResult(obj, err)
	fmt.Println("LENGTH")
	l, _ = mq.Length()
	fmt.Println(l)
}

func printResult(obj *kvmq.QueueObject, errs error) {
	b, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b[:]))
	if errs != nil {
		fmt.Println(errs.Error())
	}
}
