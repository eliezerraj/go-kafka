package adapter

import (
        "log"
        "context"
        "time"
        "strconv"
        "math/rand"
        "os"
        "encoding/json"

        "github.com/segmentio/kafka-go"
        "github.com/segmentio/kafka-go/sasl/plain"

        "github.com/go-kafka/internal/core"

)

const producer_timeout = 10

type Message struct {
    ID          int     `json:"id"`
    Description string  `json:"description"`
    Status      bool    `json:"status"`
}

type ProducerService struct{
        configurations  *core.Configurations
        writer                  *kafka.Writer
}

type Logger interface {
        Printf(string, ...interface{})
}

func NewProducerService(configurations *core.Configurations) *ProducerService {
        l := log.New(os.Stdout, "Producer: ", 0)

        kafkaBrokerUrls := []string {configurations.KafkaConfig.Brokers1,
                                                                configurations.KafkaConfig.Brokers2,
                                                                configurations.KafkaConfig.Brokers3,
																}

//      mechanism, err := scram.Mechanism(scram.SHA512,  configurations.KafkaConfig.Username,  configurations.KafkaConfig.Password)
//      if err != nil {
//              panic(err)
//      }


        mechanism := plain.Mechanism{
                                                                        Username: configurations.KafkaConfig.Username,
                                                                        Password: configurations.KafkaConfig.Password,
                                                                }

        s1 := rand.NewSource(time.Now().UnixNano())
        r1 := rand.New(s1)

        log.Print(mechanism)

        dialer := &kafka.Dialer{Timeout:  producer_timeout * time.Second,
                                                        ClientID: configurations.KafkaConfig.Clientid + "--" +strconv.Itoa(r1.Intn(100)),
//                                                      SASLMechanism: mechanism,
                                                }

        // Number of acknowledges from partition replicas required before receiving
        // a response to a produce request. The default is -1, which means to wait for
        // all replicas, and a value above 0 is required to indicate how many replicas
        // should acknowledge a message to be considered successful.
	
        config := kafka.WriterConfig{Brokers:           kafkaBrokerUrls,
                                                                Topic:                  configurations.KafkaConfig.Topic,
                                                                Dialer:                 dialer,
                                                                RequiredAcks:   configurations.KafkaConfig.RequiredAcks,
                                                                WriteTimeout:   producer_timeout * time.Second,
                                                                ReadTimeout:    producer_timeout * time.Second,
                                                                Logger:                 l, }

        writer := kafka.NewWriter(config)

        return &ProducerService{ configurations : configurations,
                                                        writer: writer,
                                                }
}

func (c *ProducerService) Close(ctx context.Context) error{
        if err := c.writer.Close(); err != nil {
                log.Printf("failed to close writer:", err)
                return err
        }
        return nil
}

func (c *ProducerService) Producer(ctx context.Context, i int) {
        log.Printf("kafka Producer")

        w :=  c.writer

        rand.Seed(time.Now().UnixNano())
        min := 1
        max := 4

        key := "key-"+ strconv.Itoa(rand.Intn(max-min+1) + min)
        //message := "teste-(" + strconv.Itoa(rand.Intn(max-min+1) + min)  + ")-" + strconv.Itoa(i)

        message := Message{}
        message.ID = rand.Intn(max-min+1) + min
        message.Description = "teste-(" + strconv.Itoa(rand.Intn(max-min+1) + min)  + ")-" + strconv.Itoa(i)
        message.Status = true

        res, _ := json.Marshal(message)
        msg := kafka.Message{
                Key:    []byte(key),
                Value:  []byte(string(res)),
                Time:  time.Now(),
        }

        log.Println("----------------------------------------")
        log.Printf("==> Message Topic %s Key %s Value %s \n", c.configurations.KafkaConfig.Topic ,string(msg.Key), string(msg.Value))
        log.Println("----------------------------------------")

        err := w.WriteMessages(ctx, msg)
        if err != nil {
                log.Printf("failed to write messages:", err)
        }
}
