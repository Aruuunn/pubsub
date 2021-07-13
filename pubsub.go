//Package pubsub implements a simple Publish Subscribe Pattern in pure golang
package pubsub

import (
	"context"
	"sync"
)

// Message is the format in which data is being passed from one channel to another
type Message struct {
	ChannelName string
	Value       interface{}
}

//Subscription will be returned after on calling NewSubscription. Should be unsubscribed on clean up
type Subscription struct {
	wg                 *sync.WaitGroup
	pubsub             *PubSub
	cancel             context.CancelFunc
	context            context.Context
	pipeChannel        chan Message
	subscribedChannels []string
}

//PubSub is type which encapsulates the data related to the publish-subscriber pattern implementation.
//It is to be instantiated only using New() function
type PubSub struct {
	context       context.Context
	subscriptions map[string][]*Subscription
}

//The function to instantiate PubSub
func New(ctx context.Context) *PubSub {
	return &PubSub{
		context:       ctx,
		subscriptions: make(map[string][]*Subscription),
	}
}

//Publish data to the specified channels
func (pubsub *PubSub) Publish(channelNames []string, value interface{}) {
	for _, channelName := range channelNames {
		subscriptions := pubsub.subscriptions[channelName]

		message := Message{
			ChannelName: channelName,
			Value:       value,
		}

		for _, subscription := range subscriptions {
			subscription.wg.Add(1)

			go func() {
				defer subscription.wg.Done()

				select {
				case subscription.pipeChannel <- message:
					return
				case <-pubsub.context.Done():
					return
				case <-subscription.context.Done():
					return
				}
			}()
		}
	}
}

//Create a NewSubscription. The returned subscription must be unsubscribed on clean up
func (pubsub *PubSub) NewSubscription(channelNames []string) *Subscription {
	pipeChannel := make(chan Message)

	ctx, cancel := context.WithCancel(context.Background())

	sub := Subscription{
		pipeChannel:        pipeChannel,
		cancel:             cancel,
		context:            ctx,
		wg:                 new(sync.WaitGroup),
		subscribedChannels: make([]string, 0),
		pubsub:             pubsub,
	}

	for _, channelName := range channelNames {
		sub.subscribedChannels = append(sub.subscribedChannels, channelName)
		pubsub.subscriptions[channelName] = append(pubsub.subscriptions[channelName], &sub)
	}

	return &sub
}

//The channel returned will recieve data from the published data to the subscribed channels
func (subscription *Subscription) Channel() <-chan Message {
	return subscription.pipeChannel
}

//Helper function to remove element at an index
func remove(s []*Subscription, i int) []*Subscription {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func indexOf(s *Subscription, subscriptions []*Subscription) (index int) {
	index = -1

	for i, sub := range subscriptions {
		if sub == s {
			index = i
		}
	}

	return
}

//Unsubscribe to the subscribed channels
func (subscription *Subscription) UnSubscribe() {
	pubsub := subscription.pubsub

	subscription.cancel()
	subscription.wg.Wait()

	for _, channelName := range subscription.subscribedChannels {
		idx := indexOf(subscription, pubsub.subscriptions[channelName])

		if idx != -1 {
			pubsub.subscriptions[channelName] = remove(pubsub.subscriptions[channelName], idx)
		}
	}
}
