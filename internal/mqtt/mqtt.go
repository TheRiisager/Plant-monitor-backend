package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	utils "riisager/backend_plant_monitor_go/internal"
	"riisager/backend_plant_monitor_go/internal/io/database"
	"riisager/backend_plant_monitor_go/internal/types"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

const SUB_FILE_PATH string = "./config/subscriptions.json"

type MqttOptions struct {
	Context             context.Context
	GlobalStore         *types.GlobalStore
	SubscriptionChannel chan types.DeviceInfo
	RealtimeChannel     chan types.Reading
	Database            database.DatabaseWrapper
}

func Run(options MqttOptions) {

	mqttContext, cancel := context.WithCancel(options.Context)

	u, err := url.Parse(os.Getenv("MQTT_URL"))
	if err != nil {
		fmt.Printf("error parsing URL: %#v\n", err)
		cancel()
		return
	}

	mqttRouter := paho.NewStandardRouter()
	mqttRouter.DefaultHandler(func(p *paho.Publish) {
		handleMessageReceived(p, options)
	})

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
				options.GlobalStore,
				options.SubscriptionChannel,
			)
		},
		OnConnectError: func(err error) {
			fmt.Println(err)
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
		fmt.Printf("Failed to connect to MQTT: %#v\n", err)
		cancel()
		return
	}

	if err = connection.AwaitConnection(mqttContext); err != nil {
		fmt.Printf("Failed to connect to MQTT: %#v\n", err)
		cancel()
		return
	}

	<-mqttContext.Done()
}

func handleMessageReceived(p *paho.Publish, options MqttOptions) {
	var payload types.Reading
	err := json.Unmarshal(p.Payload, &payload)
	if err != nil {
		fmt.Println(err)
	}

	err = utils.TrySend(payload, options.RealtimeChannel)
	if err != nil {
		fmt.Println(err)
	}
	options.Database.SaveReading(payload)
}

func handleConnection(ctx context.Context, cm *autopaho.ConnectionManager, _ *paho.Connack, store *types.GlobalStore, channel chan types.DeviceInfo) {
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

				if !store.DeviceExists(newSub.Device) {
					store.Mutex.Lock()
					store.Devices.Add(newSub)
					store.Mutex.Unlock()
				}
				fmt.Println(store.Devices)
			case <-ctx.Done():
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
