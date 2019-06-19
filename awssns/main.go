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
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	cloudevents "github.com/cloudevents/sdk-go"
	log "github.com/sirupsen/logrus"
)

const port = ":8081"

var (
	topicEnv               string
	accountRegion          string
	accountAccessKeyID     string
	accountSecretAccessKey string
	sink                   string
	protocol               string
	host                   string
)

//Clients struct represent all clients
type Clients struct {
	SNS         snsiface.SNSAPI
	CloudEvents cloudevents.Client
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

	//Set logging output levels
	_, varPresent := os.LookupEnv("DEBUG")
	if varPresent {
		log.SetLevel(log.DebugLevel)
	}

	topicEnv = os.Getenv("TOPIC")
	accountRegion = os.Getenv("AWS_REGION")
	accountAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	accountSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
}

func main() {

	flag.Parse()

	protocol = strings.Split(sink, "://")[0]
	host = strings.Split(sink, "://")[1]

	//Create client for SNS
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(accountRegion),
		Credentials: credentials.NewStaticCredentials(accountAccessKeyID, accountSecretAccessKey, ""),
	})
	if err != nil {
		log.Fatal("Unable to create SNS client: ", err)
	}

	t, err := cloudevents.NewHTTPTransport()
	if err != nil {
		log.Fatal(err)
	}

	c, err := cloudevents.NewClient(t, cloudevents.WithTimeNow())
	if err != nil {
		log.Fatal(err)
	}

	clients := Clients{
		SNS:         sns.New(sess),
		CloudEvents: c,
	}

	//Setup subscription in the background. Will keep us from having chicken/egg between server
	//being ready to respond and us having the info we need for the subscription request
	go func() {
		for {
			err := clients.attempSubscription()
			if err == nil {
				break
			}
			log.Error(err)
		}
	}()

	//Start server
	http.HandleFunc("/", clients.HandleNotification)
	http.HandleFunc("/health", healthCheckHandler)
	log.Info("Beginning to serve on port " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func (clients Clients) attempSubscription() error {
	time.Sleep(10 * time.Second)

	topic, err := clients.SNS.CreateTopic(&sns.CreateTopicInput{Name: aws.String(topicEnv)})
	if err != nil {
		return err
	}

	_, err = clients.SNS.Subscribe(&sns.SubscribeInput{
		Endpoint: &sink,
		Protocol: &protocol,
		TopicArn: topic.TopicArn,
	})
	if err != nil {
		return err
	}
	log.Debug("Finished subscription flow")
	return nil
}

//HandleNotification implements the receive interface for sns
func (clients Clients) HandleNotification(w http.ResponseWriter, r *http.Request) {

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
	log.Info(string(body))
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

		eventTime, _ := time.Parse(time.RFC3339, data["Timestamp"].(string))

		record := &SNSEventRecord{
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
				Timestamp:         eventTime,
				SigningCertURL:    data["SigningCertURL"].(string),
				Message:           data["Message"].(string),
				UnsubscribeURL:    data["UnsubscribeURL"].(string),
				Subject:           data["Subject"].(string),
			},
		}

		event := cloudevents.Event{
			Context: cloudevents.EventContextV03{
				Type:            "AWS SNS Record",
				Subject:         aws.String("AWS SNS"),
				ID:              data["MessageId"].(string),
				SpecVersion:     "1.0",
				DataContentType: aws.String("application/json"),
			}.AsV03(),
			Data: record,
		}

		_, err := clients.CloudEvents.Send(context.Background(), event)
		if err != nil {
			log.Error("Failed to send cloud events: ", err)
		}
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("OK"))
}
