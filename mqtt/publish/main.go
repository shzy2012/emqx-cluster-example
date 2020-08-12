package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func randomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func main() {

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	//MQTT
	mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	clientID := fmt.Sprintf("emqx_subscribe_client_%s", randomHex(3))
	opts := mqtt.NewClientOptions().AddBroker("tcp://172.16.0.186:8888").SetClientID(clientID)

	opts.SetKeepAlive(60 * time.Second)
	// 设置消息回调处理函数
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	go func() {

		i := 0
		for {
			// 发布消息
			token := c.Publish("testtopic/1", 1, true, fmt.Sprintf("HAproxy sent %v\n", i))
			token.Wait()
			fmt.Printf("主题:testtopic/1 发布成功\n")
			i = i + 1
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		sig := <-sigs
		log.Println(sig)
		// 断开连接
		c.Disconnect(250)
		done <- true
	}()

	<-done
	log.Println("[IAM]=>stop service.")
}
