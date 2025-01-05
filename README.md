# libmqttunnel: tunnel via MQTT broker

This library tunnels TCP Connection through the MQTT Broker.

Based on https://github.com/shirou/mqtunnel

# How to use


# Config file 

You can use client certs as well as username/password in the Config file.

```
{
    "host": "mqttbroker.example",
    "port": 8883,
	"username", "",
	"password", "",
    "caCert": "certs/root-CA.crt",
    "clientCert": "certs/5a880e326f-certificate.pem.crt",
    "privateKey": "certs/5a880e326f-private.pem.key",
    "control": "device/1/control"
}
```

Other options

- `clientId`: MQTT ClientID. If empty, random string is generated


# Architecture

Example: Local port = 2022, Remote port = 22,

```mermaid
sequenceDiagram

LocalTCP ->> LocalMQTunnel: conn read from port 2022
LocalMQTunnel ->> MQTTBroker: Publish to local port topic '/2022'
MQTTBroker ->> RemoteMQTunnel: Recieve from local port topic '/2022'
RemoteMQTunnel ->> RemoteTCP: conn write to port 22
RemoteTCP ->> RemoteMQTunnel: conn read from port 22
RemoteMQTunnel ->> MQTTBroker: Publish to local port topic '/22'
MQTTBroker ->> LocalMQTunnel: Recieve from local port topic '/22'
LocalMQTunnel ->> LocalTCP: conn write to port 2022
```

## More internal architecture

```mermaid
sequenceDiagram

participant Remote
participant RemoteTCP
participant RemoteTCPConnection
participant RemoteMQTunnel

RemoteMQTunnel ->> RemoteMQTunnel: subscribe control topic
LocalMQTunnel ->> LocalMQTunnel: make a Tunnel instance which includes local/remote port pair
LocalMQTunnel ->> LocalTCP: NewTCPConnection()
LocalTCP ->> LocalTCP: start listening
Local ->> LocalTCP: connect
LocalTCP ->> LocalMQTunnel: OpenTunnel()
LocalMQTunnel ->> RemoteMQTunnel: Publish control packet
RemoteMQTunnel ->> RemoteTCPConnection: NewTCPConnection()
RemoteTCPConnection ->> RemoteTCP: connect()
RemoteTCP ->> Remote: connect()
```


# License

- Apache License
