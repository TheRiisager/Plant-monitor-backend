package mqtt

import (
	"context"
	"fmt"
	"net/url"

	"riisager/backend_plant_monitor_go/internal/io/database"
	"riisager/backend_plant_monitor_go/internal/io/json"
	"riisager/backend_plant_monitor_go/internal/types"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

const SUB_FILE_PATH string = "./config/subscriptions.json"

type MqttOptions struct {
	Context             context.Context
	GlobalStore         *types.GlobalStore
	SubscriptionChannel chan types.SubscriptionInfo
	Database            database.DatabaseWrapper
}

func Run(options MqttOptions) {

	mqttContext, cancel := context.WithCancel(options.Context)

	//TODO replace with environment variable
	u, err := url.Parse("mqtt://localhost:1883")
	if err != nil {
		cancel()
	}

	mqttRouter := paho.NewStandardRouter()
	mqttRouter.DefaultHandler(func(p *paho.Publish) {
		handleMessageReceived(p, options.Database)
	})

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
				options.GlobalStore.Devices,
				mqttContext,
			)
			handleConnection(
				mqttContext,
				cm,
				c,
				options.GlobalStore.Devices,
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

func handleMessageReceived(p *paho.Publish, db database.DatabaseWrapper) {

}

func handleConnection(ctx context.Context, cm *autopaho.ConnectionManager, c *paho.Connack, subs types.Devices, channel chan types.SubscriptionInfo) {
	go func() {
		for {
			select {
			case newSub := <-channel:
				_, err := cm.Subscribe(ctx, &paho.Subscribe{
					Subscriptions: []paho.SubscribeOptions{
						{Topic: newSub.Device},
					},
				})

				if err != nil {
					fmt.Println(err)
				}
				subs.Add(newSub)
				fmt.Println(subs)
			case <-ctx.Done():
				SaveSubscriptionsToFile(&subs)
				return
			}
		}
	}()
}

func initSubscriptionsFromConfig(cm *autopaho.ConnectionManager, knownSubs types.Devices, ctx context.Context) {

	var subscribeList []paho.SubscribeOptions

	if knownSubs == nil {
		fmt.Println("No subscriptions in config")
		return
	}

	for _, sub := range knownSubs {
		subscribeList = append(subscribeList, paho.SubscribeOptions{Topic: sub.Device})
	}

	_, err := cm.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: subscribeList,
	})

	if err != nil {
		panic(err)
	}
}

func SaveSubscriptionsToFile(subs *types.Devices) {
	//todo replace with environment variables
	err := json.WriteToFile(
		SUB_FILE_PATH,
		&subs,
	)

	if err != nil {
		fmt.Println(err)
	}
}
