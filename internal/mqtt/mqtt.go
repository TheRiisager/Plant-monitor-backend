package mqtt

import (
	"context"
	"fmt"
	"net/url"

	"riisager/backend_plant_monitor_go/internal/io/json"
	"riisager/backend_plant_monitor_go/internal/types"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type MqttOptions struct {
	Context             context.Context
	SubscriptionChannel chan types.SubscriptionInfo
	DatabaseChannel     chan int
}

const SUB_FILE_PATH string = "./config/subscriptions.json"

func Run(options MqttOptions) {

	mqttContext, cancel := context.WithCancel(options.Context)

	u, err := url.Parse("mqtt://localhost:1883")
	if err != nil {
		cancel()
	}

	mqttRouter := paho.NewStandardRouter()
	knownsubs, err := json.ReadfromFile[[]types.SubscriptionInfo](SUB_FILE_PATH)

	if err != nil {
		cancel()
	}

	clientCfg := autopaho.ClientConfig{
		ServerUrls:            []*url.URL{u},
		KeepAlive:             20,
		SessionExpiryInterval: 60,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, c *paho.Connack) {
			initSubscriptionsFromConfig(
				cm,
				knownsubs,
				mqttContext,
			)
			handleConnection(
				mqttContext,
				cm,
				c,
				knownsubs,
				options.SubscriptionChannel,
			)
		},
		OnConnectError: func(err error) {
			cancel()
		},
		ClientConfig: paho.ClientConfig{
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					fmt.Println("received message!")
					mqttRouter.Route(pr.Packet.Packet())
					return true, nil
				}},
		},
	}

	connection, err := autopaho.NewConnection(mqttContext, clientCfg)
	if err != nil {
		fmt.Println("an error occured! " + err.Error())
		cancel()
		panic(err)
	}

	if err = connection.AwaitConnection(mqttContext); err != nil {
		fmt.Println("an error occured! " + err.Error())
		cancel()
		panic(err)
	}

	<-mqttContext.Done()
}

func handleConnection(ctx context.Context, cm *autopaho.ConnectionManager, c *paho.Connack, subs []types.SubscriptionInfo, channel chan types.SubscriptionInfo) {
	go func() {
		subList := subs
		for {
			select {
			case newSub := <-channel:
				_, err := cm.Subscribe(ctx, &paho.Subscribe{
					Subscriptions: []paho.SubscribeOptions{
						{Topic: newSub.Topic},
					},
				})

				if err != nil {
					fmt.Println(err)
				}
				subList = append(subList, newSub)
				fmt.Println(subList)
			case <-ctx.Done():
				SaveSubscriptionsToFile(&subList)
				return
			}
		}
	}()
}

func initSubscriptionsFromConfig(cm *autopaho.ConnectionManager, knownSubs []types.SubscriptionInfo, ctx context.Context) {

	var subscribeList []paho.SubscribeOptions

	if knownSubs == nil {
		fmt.Println("No subscriptions in config")
		return
	}

	for _, sub := range knownSubs {
		subscribeList = append(subscribeList, paho.SubscribeOptions{Topic: sub.Topic})
	}

	_, err := cm.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: subscribeList,
	})

	if err != nil {
		panic(err)
	}
}

func SaveSubscriptionsToFile(subs *[]types.SubscriptionInfo) {
	err := json.WriteToFile(
		SUB_FILE_PATH,
		subs,
	)

	if err != nil {
		fmt.Println(err)
	}
}
