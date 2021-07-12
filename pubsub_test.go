package pubsub_test

import (
	"context"
	"log"
	"pubsub"
	"testing"
)

type Func func (message pubsub.Message, expectedChannelName string, expectedValue interface{})


func getAsserter(t *testing.T) Func {
	return func (message pubsub.Message, expectedChannelName string, expectedValue interface{}) {
		if message.ChannelName != expectedChannelName {
			log.Fatalf("Expected %s channel, got %s \n", expectedChannelName, message.ChannelName)
			t.Fail()
		}

		if message.Value != expectedValue {
			log.Fatalf("Expected %s Value, got %s \n", expectedValue, message.Value)
			t.Fail()
		}
	}
}


func TestPubSub(t *testing.T) {
	assert := getAsserter(t)

	ctx, _ := context.WithCancel(context.Background())
	p := pubsub.New(ctx)

	sub := p.NewSubscription([] string { "one", "two", "three" })

	p.Publish([]string {"one"}, 100)

	ch := sub.Channel()
	message := <-ch

	assert(message, "one", 100)
}

