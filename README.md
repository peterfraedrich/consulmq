![ConsulMq](consulmq.png)

Docs -- [https://pkg.go.dev/github.com/peterfraedrich/consulmq](https://pkg.go.dev/github.com/peterfraedrich/consulmq)

![Go](https://github.com/peterfraedrich/consulmq/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/peterfraedrich/consulmq)](https://goreportcard.com/report/github.com/peterfraedrich/consulmq)
[![Coverage Status](https://coveralls.io/repos/github/peterfraedrich/consulmq/badge.svg?branch=master)](https://coveralls.io/github/peterfraedrich/consulmq?branch=master)

ConsulMQ allows you to use [Hashicorp Consul](https://consul.io) as a messaging queue. The idea is that you're already using Consul for configuration, monitoring, key-value DB, service-mesh, and a host of other functions, it would be nice to be able to have a simple message queue also running in Consul and eliminate the need for extra infrastrucutre (like RabbitMQ or Kafka, etc.).

## Features
* Durable, distributed task/message queue
* ConsulMQ nodes register with Consul, providing real-time visibility into how many nodes are connected
* Simple, easy-to-use API
* Based on well-established devops/infrastructure tools

## TL;DR

```go
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
```

## How it works
ConsulMQ uses Consul's key/value store as a messaging queue. Each queue consists of an `index` and the queued messages. The `_index` record  is a JSON list holds the order and mapping for all of the messages in the queue. The queued messages are represented by a unique ID and stored under that ID as the message's key. When an operation is requested (`Push`, `Pop`, etc.), the index is updated with the changes and the message ID used to locate the appropriate message.

## Pro/Con

### Pros
* Using a tool that's already in production eliminates the need for spinning up yet-another-tool
* Consul is a well-known entity
* Changes are distributed to all Consul nodes which eliminates a single point of failure

### Cons
* Consul wasn't really designed to do this
* The use of locks means that only one node can write to an index at a time
* Will not be as performant as a dedicated message broker or stream platform (AMQP, Kafka, etc.)

## Roadmap

* Enforce TTL's
* Consul Enterprise Namespace compatibility
* Logging & Monitoring
* Additional operations
    * Search
    * PushAtIndex
    * PopFromindex
    * Drain
    * ClearQueue


### License: MIT
