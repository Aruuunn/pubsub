# 🚀 PubSub


[![GoDoc](https://godoc.org/github.com/kkdai/maglev?status.svg)](https://godoc.org/github.com/arunmurugan78/pubsub) 
[![Go](https://github.com/ArunMurugan78/pubsub/actions/workflows/go.yml/badge.svg)](https://github.com/ArunMurugan78/pubsub/actions/workflows/go.yml)

A small package which implements publisher - subscriber pattern in pure Golang.

## Installation 

```bash
go get github.com/arunmurugan78/pubsub
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"github.com/arunmurugan78/pubsub"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	ctx := context.Background()

	//Create a new PubSub instance
	pbsb := pubsub.New(ctx)

	channelNames := []string{"channelOne", "channelTwo"}

	//Subscribe to required "channels"
	sub := pbsb.NewSubscription(channelNames)
	defer sub.UnSubscribe()

	//Will get the published data through this go channel
	channel := sub.Channel()

	wg.Add(1)

	go func() {
		defer wg.Done()
		select {
		case message := <-channel:
			// Received Data will be of type pubsub.Message
			fmt.Println("Received ", message.Value, " from ", message.ChannelName)
			//Output "Received  100  from  channelTwo"
		case <-ctx.Done():
			return
		}
	}()

	//Publish data to the specified channels
	//Here we are publishing value 100 to channelTwo
	pbsb.Publish([]string{"channelTwo"}, 100)
	wg.Wait()
}

```



