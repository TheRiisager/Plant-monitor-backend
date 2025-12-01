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
	SubscriptionChannel chan types.SubscriptionInfo
	DatabaseChannel     chan int
}

const SUB_FILE_PATH string = "./config/subscriptions.json"

func Run(rootContext context.Context, options MqttOptions) {

	mqttContext, cancel := context.WithCancel(rootContext)

	u, err := url.Parse("mqtt://localhost:1883")
	if err != nil {
		cancel()
		panic(err)
	}

	mqttRouter := paho.NewStandardRouter()
	knownsubs, err := json.ReadfromFile[types.SubscriptionInfo](SUB_FILE_PATH)

	if err != nil {
		panic(err)
	}

	clientCfg := autopaho.ClientConfig{
		ServerUrls:            []*url.URL{u},
		KeepAlive:             20,
		SessionExpiryInterval: 60,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, c *paho.Connack) {
			initSubscriptionsFromConfig(cm, knownsubs, mqttContext)
			go func() {
				for {
					select {
					case newSub := <-options.SubscriptionChannel:
						fmt.Println(newSub)
						_, err := cm.Subscribe(mqttContext, &paho.Subscribe{
							Subscriptions: []paho.SubscribeOptions{
								{Topic: newSub.Topic},
							},
						})

						if err != nil {
							fmt.Println(err)
						}
						writeErr := addSubscriptionToSubList(newSub, knownsubs)
						if writeErr != nil {
							fmt.Println(writeErr)
						}
					case <-mqttContext.Done():
						return
					}
				}
			}()
		},
		OnConnectError: func(err error) {
			cancel()
			panic(err)
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
}

func initSubscriptionsFromConfig(cm *autopaho.ConnectionManager, knownSubs []types.SubscriptionInfo, ctx context.Context) {

	var subscribeList []paho.SubscribeOptions

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

func addSubscriptionToSubList(newSub types.SubscriptionInfo, knownSubs []types.SubscriptionInfo) error {
	return json.AppendToFile(
		SUB_FILE_PATH,
		append(knownSubs, newSub),
	)
}
