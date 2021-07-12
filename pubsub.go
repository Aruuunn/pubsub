package pubsub

import "context"

type Message struct {
	ChannelName string
	Value interface{}
}


type PubSub struct {
	channels map[string] chan Message
}

type Subscription struct {
	pipeChannel <-chan Message
	cancel context.CancelFunc
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
			channel <- Message {
				ChannelName: channelName,
				Value: value,
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
					case <- ctx.Done():
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


func (subscription *Subscription) UnSubscribe() {
	subscription.cancel()
}
