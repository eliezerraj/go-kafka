package adapter

import (
	"log"
	"context"
	"time"
	"os"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	"github.com/go-kafka/internal/core"

)

const consumer_timeout = 10
var lag_consumer = 0

type ConsumerService struct{
	configurations	*core.Configurations
	reader			*kafka.Reader
}

func NewConsumerService(configurations *core.Configurations) *ConsumerService {	
	l := log.New(os.Stdout, "Consumer: ", 0)

	lag_consumer = configurations.KafkaConfig.Lag

	kafkaBrokerUrls := []string {configurations.KafkaConfig.Brokers1,
								configurations.KafkaConfig.Brokers2,
								configurations.KafkaConfig.Brokers3,}

	sramMech, err := scram.Mechanism(scram.SHA512, configurations.KafkaConfig.Username, configurations.KafkaConfig.Password)
	if err != nil {
					log.Printf("failed to SASL:", err)
	}
						
	mechanism := plain.Mechanism{
									Username: configurations.KafkaConfig.Username,
									Password: configurations.KafkaConfig.Password,
								}

	log.Print(mechanism)
	log.Print(sramMech)

	dialer := &kafka.Dialer{	Timeout:  consumer_timeout * time.Second,
						 //SASLMechanism: mechanism,//sramMech,
							}

	config := kafka.ReaderConfig{	Brokers:  kafkaBrokerUrls,
									GroupID:  configurations.KafkaConfig.Groupid,
									Dialer:   dialer,
									Topic:    configurations.KafkaConfig.Topic,
									MaxWait:  consumer_timeout * time.Second,
									MinBytes:  900e3, // 1KB  - Batch Message
									MaxBytes:  10e6, // 10MB  - Batch Message
									CommitInterval: time.Second, // flushes commits to Kafka every second
									StartOffset: -2, // - 1 The most recent offset available for a partition or -2 The least recent offset available for a partition.
									//Partition: configurations.KafkaConfig.Partition,
									Logger: l, 
								}
	
	reader := kafka.NewReader(config)
	
	return &ConsumerService{ configurations : configurations,
							reader: reader,
						}
}

func (c *ConsumerService) Close(ctx context.Context) error{
	if err := c.reader.Close(); err != nil {
		log.Printf("failed to close reader:", err)
		return err
	}
	return nil
}

func (c *ConsumerService) Consumer(ctx context.Context) {
	log.Printf("kafka Consumer")

	r := c.reader

	for {
			msg, err := r.ReadMessage(ctx)
			if err != nil {
					// Comment below just for testing
					//panic("could not read message " + err.Error())
					log.Println("Error Reading Kafka Stream : ",err)
			}

			log.Println("----------------------------------------")
			// log.Printf("Message Incoming Topic %s Partition %d Offset [%d]  %s MSG = %s\n", msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
			// log.Println("----------------------------------------")

			if ( lag_consumer > 0){
					//log.Println("Waiting for %s", lag_consumer)
					time.Sleep(time.Second * time.Duration(lag_consumer))
			}

			if err = r.CommitMessages(ctx, msg); err != nil {
					log.Println("failed to commit messages:", err)
			}

			log.Printf("***** Key: %s  | Message: %s ****\n ",  string(msg.Key) ,string(msg.Value))
	}
}

