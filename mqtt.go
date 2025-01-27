package libmqttunnel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type mqttBroker struct {
	client           mqtt.Client
	conf             Config
	mqttDisconnectCh chan bool
	controlTopic     string
	tunnelTopics     map[string]*Tunnel // topic: tunnel

	controlCh chan controlPacket
}

const mqttCommandsTimeout = 5 * time.Second
const topicQoS = 0

func NewMQTTBroker(conf Config, controlCh chan controlPacket) (*mqttBroker, error) {
	ret := mqttBroker{
		conf:             conf,
		mqttDisconnectCh: make(chan bool),
		tunnelTopics:     make(map[string]*Tunnel),

		controlTopic: conf.Control,

		controlCh: controlCh,
	}

	opts, err := getMQTTOptions(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to get MQTT options, %w", err)
	}

	// add callback
	opts.SetConnectionLostHandler(ret.onMqttConnectionLost)
	opts.SetOnConnectHandler(ret.onConnect)
	opts.SetReconnectingHandler(ret.onReconnect)

	mqtt.ERROR, _ = zap.NewStdLogAt(zap.L(), zap.ErrorLevel)
	mqtt.CRITICAL, _ = zap.NewStdLogAt(zap.L(), zap.ErrorLevel)
	// mqtt.WARN, _ = zap.NewStdLogAt(zap.L(), zap.WarnLevel)
	// mqtt.DEBUG, _ = zap.NewStdLogAt(logger.Desugar(), zap.DebugLevel)

	// connect to MQTT Broker
	client := mqtt.NewClient(opts)
	ret.client = client

	// connect first time
	if err := ret.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect broker, %w", err)
	}

	return &ret, nil
}

func (mqb *mqttBroker) start(ctx context.Context) error {
	for {
		select {
		case <-mqb.mqttDisconnectCh:
			zap.S().Error("mqtt disconnect message. try to reconnect")
			// do nothing. auto-reconnect should work
		case <-ctx.Done():
			zap.S().Warnf("MQTTConnection finished, %v", ctx.Err())
			return ctx.Err()
		}
	}
}

func (mqb *mqttBroker) publish(ctx context.Context, topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	zap.S().Debugw("mqtt publish", zap.String("topic", topic))

	return mqb.client.Publish(topic, qos, retained, payload)
}

func (mqb *mqttBroker) connect() error {
	zap.S().Debugf("connect start")
	token := mqb.client.Connect()
	if c := token.WaitTimeout(mqttCommandsTimeout); !c {
		return fmt.Errorf("connect timed out")
	}
	return token.Error()
}

// subscribeTunnelTopic subscribe topic
func (mqb *mqttBroker) subscribeTunnelTopic(topic string, tunnel *Tunnel) error {
	mqb.tunnelTopics[topic] = tunnel

	return mqb.subscribe()
}

func (mqb *mqttBroker) subscribe() error {
	topics := make(map[string]byte)

	if mqb.controlTopic != "" {
		topics[mqb.controlTopic] = 1
	}
	for t, _ := range mqb.tunnelTopics {
		topics[t] = tunnelQoS
	}

	if len(topics) == 0 {
		return nil
	}

	zap.S().Infow("topic subscribing", zap.Strings("topic", logTopic(topics)))

	subscribeToken := mqb.client.SubscribeMultiple(topics, mqb.onMessage)
	if c := subscribeToken.WaitTimeout(mqttCommandsTimeout); !c {
		return fmt.Errorf("subscribe timed out")
	}
	return subscribeToken.Error()
}

func (mqb *mqttBroker) unsubscribe(topic string) error {
	if topic == "" {
		return nil
	}

	zap.S().Debugw("topic unsubscribing", zap.String("topic", topic))

	token := mqb.client.Unsubscribe(topic)
	if !token.WaitTimeout(mqttCommandsTimeout) {
		return fmt.Errorf("unsubscribe timeout (%s)", topic)
	}
	if token.Error() != nil {
		return token.Error()
	}
	delete(mqb.tunnelTopics, topic)

	return nil
}

func (mqb *mqttBroker) onMessage(client mqtt.Client, msg mqtt.Message) {
	zap.S().Debugw("on message", zap.String("topic", msg.Topic()), zap.Int("size", len(msg.Payload())))

	if msg.Topic() == mqb.conf.Control {
		if err := mqb.controlPacketReceived(msg); err != nil {
			zap.S().Error(err)
		}
		return
	}
	tun, exists := mqb.tunnelTopics[msg.Topic()]
	if !exists {
		zap.S().Errorw("requested topic is not exists",
			zap.String("topic", msg.Topic()))
		return
	}
	tun.writeCh <- msg.Payload()
}

func (mqb *mqttBroker) controlPacketReceived(msg mqtt.Message) error {
	var control controlPacket
	if err := json.Unmarshal(msg.Payload(), &control); err != nil {
		return fmt.Errorf("unmarshal error, %v", err)
	}
	mqb.controlCh <- control
	return nil
}

func (mqb *mqttBroker) onConnect(client mqtt.Client) {
	zap.S().Info("connected")
	if err := mqb.subscribe(); err != nil {
		zap.S().Errorw("subscribe failed", zap.Error(err))
	}
}

func (mqb *mqttBroker) onReconnect(client mqtt.Client, opts *mqtt.ClientOptions) {
	zap.S().Info("reconnecting...")
}

func (mqb *mqttBroker) onMqttConnectionLost(client mqtt.Client, err error) {
	zap.S().Error("MQTT connection lost", zap.Error(err))
	mqb.mqttDisconnectCh <- true
}

func newTLSConfig(config Config) (*tls.Config, error) {
	rootCA, err := os.ReadFile(config.CaCert)
	if err != nil {
		return nil, err
	}
	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(rootCA)
	cert, err := tls.LoadX509KeyPair(config.ClientCert, config.PrivateKey)
	if err != nil {
		return nil, err
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		RootCAs:            certpool,
		InsecureSkipVerify: true,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		Certificates:       []tls.Certificate{cert},
	}, nil
}

func getMQTTOptions(conf Config) (*mqtt.ClientOptions, error) {
	if conf.ClientID == "" {
		u, err := uuid.NewRandom()
		if err != nil {
			return nil, fmt.Errorf("failed to make uuid, %w", err)
		}
		conf.ClientID = u.String()
	}

	opts := mqtt.NewClientOptions()

	if conf.Port == 1883 {
		opts.AddBroker(fmt.Sprintf("tcp://%s:%d", conf.Host, conf.Port))
	} else {
		opts.AddBroker(fmt.Sprintf("ssl://%s:%d", conf.Host, conf.Port))
		if conf.ClientCert != "" {
			tlsConfig, err := newTLSConfig(conf)
			if err != nil {
				return nil, fmt.Errorf("failed to construct tls config, %v", err)
			}
			opts.SetTLSConfig(tlsConfig)
		} else if conf.IgnoreSslErrors {
			tlsConfig := &tls.Config {
				InsecureSkipVerify: true,
			}
			opts.SetTLSConfig(tlsConfig)
		}
	}
	if conf.UserName != "" {
		opts.SetUsername(conf.UserName)
	}
	if conf.Password != "" {
		opts.SetPassword(conf.Password)
	}
	opts.SetClientID(conf.ClientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetryInterval(20 * time.Second)

	return opts, nil
}

// logTopic is a util function to log multiple topics
func logTopic(topics map[string]byte) []string {
	ret := make([]string, 0, len(topics))
	for k := range topics {
		ret = append(ret, k)
	}

	return ret
}
