package pubsub_test

import (
	"context"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"
	"github.com/arunmurugan78/pubsub"
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


func TestPubSubWorstCase(t *testing.T) {
	for i :=1; i<30; i++  {

		var wg sync.WaitGroup
		ctx := context.Background()
		pbsb := pubsub.New(ctx)

		channels := make([]string, 0)
		subs := make([]*pubsub.Subscription, 0)

		for j:=0;j<i;j++ {
			channels = append(channels, strconv.Itoa(j))
			subs = append(subs, pbsb.NewSubscription(channels))
		}

		channels = make([]string, 0)

		for j:=0;j<i;j++ {
			channels = append(channels, strconv.Itoa(j))
			
			pbsb.Publish(channels, j)


			for k:=j; k<i;k ++ {
				wg.Add(1)
				go func (idx int, j int, l int){
					defer wg.Done()	

					
					for c := 0; c< l-len(subs[idx].GetSubscribedChannels());c++ {					
						message := <-subs[idx].Channel()
						if message.Value != j {
							log.Fatalf("Expected %d but go %d", j, message.Value)
						}
					}

				}(k, j, len(channels))
			}

			wg.Wait()
		}

		for _, sub := range subs {
			sub.UnSubscribe()
		}
	}
}