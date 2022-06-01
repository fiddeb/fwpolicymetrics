package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/influxdata/go-syslog/rfc5424"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type counter struct {
	name string
	help string
}

var (
	parser     = rfc5424.NewParser()
	splitAt    = []byte(" <")
	bestEffort = true
	action     string
)

func main() {

	exporterHost := flag.String("host", "127.0.0.1", "Exporter host")
	exporterPort := flag.String("port", "8080", "Exporter port")
	bootstrapServers := flag.String("bootstapservers", "localhost", "Kafka servers")
	kafkaGroup := flag.String("group", "localgroup", "Kafka consumer group")
	kafkaOffset := flag.String("offset", "latest", "Kafka offet strategy")
	kafkaTopic := flag.String("topic", "mytopic", "Kafka topic to consume")
	flag.Parse()

	exporter := *exporterHost + ":" + *exporterPort

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(exporter, nil)
	}()

	messages := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_total",
			Help: "How many messages recived from kafka, by patrtition",
		},
		[]string{"partition"},
	)
	errormessage := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_error_total",
			Help: "How many messages failed from kafka, by patrtition",
		},
		[]string{"partition"},
	)

	parseError := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "syslog_messages_error_total",
			Help: "How many messages failed in parser",
		},
		[]string{"type"},
	)

	flowsession := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rt_flow_session_total",
			Help: "How many policy decisions made by the firewall",
		}, []string{"hostname", "policy", "action", "source", "destination"},
	)
	prometheus.MustRegister(messages)
	prometheus.MustRegister(errormessage)
	prometheus.MustRegister(flowsession)
	prometheus.MustRegister(parseError)
	mall := messages.WithLabelValues("all")

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": string(*bootstrapServers),
		"group.id":          string(*kafkaGroup),
		"auto.offset.reset": string(*kafkaOffset),
	})

	if err != nil {
		panic(err)
	} else {
		fmt.Printf("connected to: %s, on %s", *bootstrapServers, *kafkaTopic)
	}

	c.SubscribeTopics([]string{string(*kafkaTopic)}, nil)

	for {
		msg, err := c.ReadMessage(-1)
		if err != nil {
			fmt.Println(err)

			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)

			bump := errormessage.WithLabelValues(fmt.Sprint(msg.TopicPartition.Partition))
			bump.Inc()
			continue
		}
		//fmt.Printf("%+v\n", data)
		m := messages.WithLabelValues(fmt.Sprint(msg.TopicPartition.Partition))
		m.Inc()
		mall.Inc()

		data, e := unpack(msg.Value)
		//rfc5424message, perr := simpleParser(string(msg.Value))
		if e != nil {
			parserr := parseError.WithLabelValues("rfc5424")
			parserr.Inc()
			//fmt.Printf("%s, %s", e,string(msg.Value))
			continue
		}
		if data.StructuredData() != nil {
			for _, v := range *data.StructuredData() {

				switch *data.MsgID() {
				case "RT_FLOW_SESSION_CREATE":
					action = "permit"
				case "RT_FLOW_SESSION_DENY":
					action = "deny"
				case "RT_FLOW_SESSION_CLOSE":
					action = "close"
				default:
					action = "unknown"
				}

				policyName, ok := v["policy-name"]
				sourceZoneName, ok := v["source-zone-name"]
				desinationZoneName, ok := v["destination-zone-name"]
				if ok {
					//fmt.Printf("'%s'\n", policyName)

					fw := flowsession.WithLabelValues(*data.Hostname(), fmt.Sprint(policyName), action, sourceZoneName, desinationZoneName)
					fw.Inc()
				}
			}
		}
	}

}

func unpack(v []byte) (*rfc5424.SyslogMessage, error) {

	// find the start of the syslog rfc5424 payload
	idx := bytes.Index(v, splitAt)
	if idx == -1 {
		return nil, errors.New("invalid syslog message")
	}

	// read the payload from the index to the end
	data := v[idx+1:]

	// parse the syslog payload
	m, err := parser.Parse(data, &bestEffort)
	if err != nil {
		return nil, err
	}

	return m, nil
}
