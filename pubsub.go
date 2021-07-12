package pubsub

import "context"

type Message struct {
	ChannelName string
	Value interface{}
}


type PubSub struct {
	channels map[string] chan Message
	context context.Context
}

type Subscription struct {
	pipeChannel <-chan Message
	cancel context.CancelFunc
}


func New(ctx context.Context) *PubSub {
	return &PubSub {
		channels: make(map[string]chan Message),
		context: ctx,
	}
}


func (p *PubSub) getChannel(channelName string) chan Message {
	channel, ok := p.channels[channelName]

	if !ok {
		channel = make(chan Message)
		p.channels[channelName] = channel
	}

	return channel
}


func (p *PubSub) Publish(channelNames []string, value interface{}) {
	for _, channelName := range channelNames {
		channel := p.getChannel(channelName)

		go func () {

			message :=  Message {
				ChannelName: channelName,
				Value: value,
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

func (p *PubSub) NewSubscription(channelNames []string) Subscription {
	pipeChannel := make(chan Message)

	ctx, cancel := context.WithCancel(context.Background())

	for _, channelName := range channelNames {
		channel := p.getChannel(channelName)

		go func () {
			for {
				select {
					case value := <- channel:
						go func () {
							pipeChannel <- value
						}()
						return
					case <- ctx.Done():
						return
					case <- p.context.Done():
						return 
				}
			}
		}()
	}

	return Subscription {
		pipeChannel: pipeChannel,
		cancel: cancel,
	}
}


func  (subscription *Subscription) Channel() <-chan Message {
	return subscription.pipeChannel
}

func (subscription *Subscription) UnSubscribe() {
	subscription.cancel()
}
