//Package pubsub implements a simple Publish Subscribe Pattern in pure golang
package pubsub

import "context"

// Message is the format in which data is being passed from one channel to another
type Message struct {
	ChannelName string
	Value       interface{}
}

//PubSub is type which encapsulates the data related to the publish-subscriber pattern implementation.
//It is to be instantiated only using New() function
type PubSub struct {
	channels map[string]chan Message
	context  context.Context
}

//Subscription will be returned after a subscription. Should be unsubscribed on clean up
type Subscription struct {
	pipeChannel <-chan Message
	cancel      context.CancelFunc
}

//The function to instantiate PubSub
func New(ctx context.Context) *PubSub {
	return &PubSub{
		channels: make(map[string]chan Message),
		context:  ctx,
	}
}

//Returns go channel related to the channelName, creates one if not present
func (p *PubSub) getChannel(channelName string) chan Message {
	channel, ok := p.channels[channelName]

	if !ok {
		channel = make(chan Message)
		p.channels[channelName] = channel
	}

	return channel
}

//Publish data to the specified channels
func (p *PubSub) Publish(channelNames []string, value interface{}) {
	for _, channelName := range channelNames {
		channel := p.getChannel(channelName)

		go func() {

			message := Message{
				ChannelName: channelName,
				Value:       value,
			}

			for {
				select {
				case channel <- message:
					return
				case <-p.context.Done():
					return
				}
			}
		}()
	}
}

//Create a NewSubscription. The returned subscription must be unsubscribed on clean up
func (p *PubSub) NewSubscription(channelNames []string) Subscription {
	pipeChannel := make(chan Message)

	ctx, cancel := context.WithCancel(context.Background())

	for _, channelName := range channelNames {
		channel := p.getChannel(channelName)

		go func() {
			for {
				select {
				case value := <-channel:
					go func() {
						pipeChannel <- value
					}()
					return
				case <-ctx.Done():
					return
				case <-p.context.Done():
					return
				}
			}
		}()
	}

	return Subscription{
		pipeChannel: pipeChannel,
		cancel:      cancel,
	}
}

//The channel returned will recieve data from the published data to the subscribed channels
func (subscription *Subscription) Channel() <-chan Message {
	return subscription.pipeChannel
}

//Unsubscribe to the subscribed channels
func (subscription *Subscription) UnSubscribe() {
	subscription.cancel()
}
