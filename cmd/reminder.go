package cmd

import (
	"time"

	"github.com/orensimple/otus_events_reminder/config"
	"github.com/orensimple/otus_events_reminder/internal/logger"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
)

var addr string

var RootCmd = &cobra.Command{
	Use:   "reminder",
	Short: "Run reminder events",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init(addr)
		logger.InitLogger()
		timer1 := time.NewTimer(30 * time.Second)
		<-timer1.C
		startRecieve()
	},
}

func init() {
	RootCmd.Flags().StringVar(&addr, "config", "./config", "")
}

func startRecieve() {
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672/")
	if err != nil {
		logger.ContextLogger.Errorf("Failed to connect to RabbitMQ", err.Error())
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.ContextLogger.Errorf("Failed to open a channel", err.Error())
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"eventsByDate", // name
		"direct",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		logger.ContextLogger.Infof("Failed to declare an exchange", err.Error())
	}

	_, err = ch.QueueDeclare(
		"eventsByDay", // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)

	err = ch.QueueBind(
		"eventsByDay",  // name
		"day",          // key
		"eventsByDate", // exchange
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		logger.ContextLogger.Infof("Problem bind queue", err.Error())
	}

	msgs, err := ch.Consume(
		"eventsByDay",   // queue
		"ConsumerByDay", // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		logger.ContextLogger.Errorf("Failed to register a consumer", err.Error())
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			logger.ContextLogger.Infof("Received a message:", d.Body)
		}
	}()
	logger.ContextLogger.Infof(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
