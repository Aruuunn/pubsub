package pubsub_test

import (
	"context"
	"github.com/arunmurugan78/pubsub"
	"log"
	"testing"
	"time"
)

type Func func(message pubsub.Message, expectedChannelName string, expectedValue interface{})

func getAsserter(t *testing.T) Func {
	return func(message pubsub.Message, expectedChannelName string, expectedValue interface{}) {
		if message.ChannelName != expectedChannelName {
			log.Fatalf("Expected %s channel, got %s \n", expectedChannelName, message.ChannelName)
		}

		if message.Value != expectedValue {
			log.Fatalf("Expected %s Value, got %s \n", expectedValue, message.Value)
		}
	}
}

func TestPubSub(t *testing.T) {
	assert := getAsserter(t)

	t.Run("Should be able to publish and subscribe to channels and should be able to receive published values", func(t *testing.T) {
		ctx := context.Background()

		p := pubsub.New(ctx)

		sub := p.NewSubscription([]string{"one", "two", "three"})

		p.Publish([]string{"one"}, 100)

		ch := sub.Channel()

		message := <-ch

		assert(message, "one", 100)

		p.Publish([]string{"two"}, 201)

		message = <-ch

		assert(message, "two", 201)

		p.Publish([]string{"four"}, 300)

		p.Publish([]string{"three"}, 400)

		message = <-ch

		assert(message, "three", 400)
	})

	t.Run("should not receive published changes after un subscribe", func(t *testing.T) {
		ctx := context.Background()
		p := pubsub.New(ctx)

		sub := p.NewSubscription([]string{"one", "two", "three", "four"})

		ch := sub.Channel()

		timeout := time.After(3 * time.Second)

		sub.UnSubscribe()

		p.Publish([]string{"one"}, 101)

		select {
		case <-ch:
			log.Fatalf("Got published updates")
		case <-timeout:
			return
		}

	})

	t.Run("should not receive published changes after context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		p := pubsub.New(ctx)

		sub := p.NewSubscription([]string{"one", "two", "three", "four"})

		ch := sub.Channel()

		timeout := time.After(3 * time.Second)

		cancel()

		p.Publish([]string{"one"}, 101)

		time.Sleep(1 * time.Second)

		for {
			select {
			case <-ch:
				log.Fatalf("Got published updates after calling cancel")
			case <-timeout:
				return
			}
		}

	})
}
