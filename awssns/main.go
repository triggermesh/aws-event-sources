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
	log "github.com/sirupsen/logrus"
	"github.com/triggermesh/sources/tmevents"
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

type snsMsg struct{}

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

	//Start server
	m := snsMsg{}
	http.HandleFunc("/", m.ReceiveMsg)
	http.HandleFunc("/healthz", health)
	log.Info("Beginning to serve on port " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

//ReceiveMsg implements the receive interface for sns
func (snsMsg) ReceiveMsg(w http.ResponseWriter, r *http.Request) {

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

		eventTime, _ := time.Parse(time.RFC3339, data["Timestamp"].(string))
		eventInfo := tmevents.EventInfo{
			EventData:   []byte(data["Message"].(string)),
			EventID:     data["MessageId"].(string),
			EventTime:   eventTime,
			EventType:   data["Type"].(string),
			EventSource: "sns",
		}
		log.Debug("Received notification: ", eventInfo)

		err := tmevents.PushEvent(&eventInfo, sink)
		if err != nil {
			log.Error("Unable to push event: ", err)
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
