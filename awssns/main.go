/*
Copyright (c) 2019 TriggerMesh, Inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/knative/pkg/cloudevents"
	log "github.com/sirupsen/logrus"
)

const port = ":8081"

var (
	topicEnv     string
	domainEnv    string
	protocolEnv  string
	channelEnv   string
	namespaceEnv string
	awsRegionEnv string
	awsCredsFile string
	sink         string
)

//Client struct represent all clients
type Client struct {
	CloudEvents *cloudevents.Client
}

type SNSEventRecord struct {
	EventVersion         string    `json:"EventVersion"`
	EventSubscriptionArn string    `json:"EventSubscriptionArn"`
	EventSource          string    `json:"EventSource"`
	SNS                  SNSEntity `json:"Sns"`
}

type SNSEntity struct {
	Signature         string                 `json:"Signature"`
	MessageID         string                 `json:"MessageId"`
	Type              string                 `json:"Type"`
	TopicArn          string                 `json:"TopicArn"`
	MessageAttributes map[string]interface{} `json:"MessageAttributes"`
	SignatureVersion  string                 `json:"SignatureVersion"`
	Timestamp         time.Time              `json:"Timestamp"`
	SigningCertURL    string                 `json:"SigningCertUrl"`
	Message           string                 `json:"Message"`
	UnsubscribeURL    string                 `json:"UnsubscribeUrl"`
	Subject           string                 `json:"Subject"`
}

func init() {
	flag.StringVar(&sink, "sink", "", "where to sink events to")
}

func main() {

	flag.Parse()

	//Set logging output levels
	_, varPresent := os.LookupEnv("DEBUG")
	if varPresent {
		log.SetLevel(log.DebugLevel)
	}

	domainEnv = os.Getenv("DOMAIN")
	topicEnv = os.Getenv("TOPIC")
	protocolEnv = os.Getenv("PROTOCOL")
	channelEnv = os.Getenv("CHANNEL")
	namespaceEnv = os.Getenv("NAMESPACE")
	awsRegionEnv = os.Getenv("AWS_REGION")
	awsCredsFile = os.Getenv("AWS_CREDS")

	//Setup subscription in the background. Will keep us from having chicken/egg between server
	//being ready to respond and us having the info we need for the subscription request
	go topicSubscribe()

	client := Client{
		CloudEvents: cloudevents.NewClient(
			sink,
			cloudevents.Builder{
				Source: "aws:sns",
			},
		),
	}

	//Start server
	http.HandleFunc("/", client.ReceiveMsg)
	http.HandleFunc("/healthz", health)
	log.Info("Beginning to serve on port " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

//ReceiveMsg implements the receive interface for sns
func (c Client) ReceiveMsg(w http.ResponseWriter, r *http.Request) {

	//Fish out notification body
	var notification interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Failed to parse notification: ", err)
	}
	err = json.Unmarshal(body, &notification)
	if err != nil {
		log.Error("Failed to parse notification: ", err)
	}
	data := notification.(map[string]interface{})

	//If the message is about our subscription, curl the confirmation endpoint.
	if data["Type"].(string) == "SubscriptionConfirmation" {

		subcribeURL := data["SubscribeURL"].(string)
		_, err := http.Get(subcribeURL)
		if err != nil {
			log.Fatal("Unable to confirm SNS subscription: ", err)
		}
		log.Info("Successfully confirmed SNS subscription")

		//If it's a legit notification, push the event
	} else if data["Type"].(string) == "Notification" {

		record := SNSEventRecord{
			EventVersion:         "1.0",
			EventSubscriptionArn: "",
			EventSource:          "aws:sns",
			SNS: SNSEntity{
				Signature:         data["Signature"].(string),
				MessageID:         data["MessageId"].(string),
				Type:              data["Type"].(string),
				TopicArn:          data["TopicArn"].(string),
				MessageAttributes: data["MessageAttributes"].(map[string]interface{}),
				SignatureVersion:  data["SignatureVersion"].(string),
				Timestamp:         data["Timestamp"].(time.Time),
				SigningCertURL:    data["SigningCertURL"].(string),
				Message:           data["Message"].(string),
				UnsubscribeURL:    data["UnsubscribeURL"].(string),
				Subject:           data["Subject"].(string),
			},
		}

		log.Debug("Received notification: ", record)

		if err := c.CloudEvents.Send(record); err != nil {
			log.Error(err)
		}
	}
}

//Handle health checks
func health(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("OK"))
}

//Initiate subscription request
func topicSubscribe() {

	webhookBase := "sns." + topicEnv + "." + channelEnv + "." + namespaceEnv + "." + domainEnv
	webhookURL := protocolEnv + "://" + webhookBase

	//Create client for SNS
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegionEnv),
		Credentials: credentials.NewSharedCredentials(awsCredsFile, "default"),
		MaxRetries:  aws.Int(5),
	})
	if err != nil {
		log.Fatal("Unable to create SNS client: ", err)
	}
	snsClient := sns.New(sess)

	//Attempt subscription flow until successful
	for {
		time.Sleep(10 * time.Second)
		ip, err := net.LookupIP(webhookBase)
		if err != nil {
			log.Error("Waiting for DNS entry: ", err)
			continue
		}
		log.Info("Found IP: ", ip)

		//CreateTopic is supposed to be idempotent, so if topic name already exists, just returns ARN.
		topic, err := snsClient.CreateTopic(&sns.CreateTopicInput{Name: aws.String(topicEnv)})
		if err != nil {
			log.Fatal("Unable to create/fetch SNS topic: ", err)
			continue
		}

		//Tells SNS to send the subscription confirmation payload the the endpoint provided.
		_, err = snsClient.Subscribe(&sns.SubscribeInput{
			Endpoint: &webhookURL,
			Protocol: &protocolEnv,
			TopicArn: topic.TopicArn,
		})
		if err != nil {
			log.Error("Unable to send SNS subscription request: ", err)
			continue
		}
		log.Debug("Finished subscription flow")
		break
	}
}
