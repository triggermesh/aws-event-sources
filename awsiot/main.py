'''
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
'''

from AWSIoTPythonSDK.MQTTLib import AWSIoTMQTTClient

from cloudevents.sdk import converters
from cloudevents.sdk import marshaller
from cloudevents.sdk.converters import structured
from cloudevents.sdk.event import v01

import logging
import requests
import time
import argparse
import json
import os
import datetime

topic = os.environ['THING_TOPIC']
host = os.environ['THING_SHADOW_ENDPOINT']
rootCAPath = os.environ['ROOT_CA_PATH']
certificatePath = os.environ['CERTIFICATE_PATH']
privateKeyPath = os.environ['PRIVATE_KEY_PATH']

rootCA = os.environ['ROOT_CA']
certificate = os.environ['CERTIFICATE']
privateKey = os.environ['PRIVATE_KEY']
sink_url = os.environ['SINK']

def write_credentials():
  rca=open(rootCAPath, 'w')
  cert=open(certificatePath, 'w')
  pk=open(privateKeyPath, 'w')

  rca.write(str(rootCA))
  cert.write(str(certificate))
  pk.write(str(privateKey))

  rca.close()
  cert.close()
  pk.close()

logger = logging.getLogger("AWSIoTPythonSDK.core")
logger.setLevel(logging.DEBUG)



def customCallback(client, userdata, message):
    logger.info("Received a new message: ")
    logger.info(message.payload)
    logger.info("from topic: ")
    logger.info(message.topic)
    logger.info("--------------\n\n")

    print(message.data.decode())
    print(sink_url)

    event = (
        v01.Event().
        SetContentType("application/json").
        SetData(message.payload).
        SetEventID("my-id").
        SetSource("AWS IoT").
        SetEventTime(datetime.datetime.now()).
        SetEventType("cloudevent.greet.you")
    )
    m = marshaller.NewHTTPMarshaller(
        [
            structured.NewJSONHTTPCloudEventConverter()
        ]
    )

    headers, body = m.ToRequest(event, converters.TypeStructured, lambda x: x)

    requests.post(sink_url, data=body, headers=headers)
    message.ack()

if __name__ == "__main__":
    # Configure logging
    streamHandler = logging.StreamHandler()
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    streamHandler.setFormatter(formatter)
    logger.addHandler(streamHandler)


    myAWSIoTMQTTClient = AWSIoTMQTTClient("testIoTPySDK")
    myAWSIoTMQTTClient.configureEndpoint(host, 443)
    myAWSIoTMQTTClient.configureCredentials(rootCAPath, privateKeyPath, certificatePath)

    # AWSIoTMQTTClient connection configuration
    myAWSIoTMQTTClient.configureAutoReconnectBackoffTime(1, 32, 20)
    myAWSIoTMQTTClient.configureOfflinePublishQueueing(-1)  # Infinite offline Publish queueing
    myAWSIoTMQTTClient.configureDrainingFrequency(2)  # Draining: 2 Hz
    myAWSIoTMQTTClient.configureConnectDisconnectTimeout(10)  # 10 sec
    myAWSIoTMQTTClient.configureMQTTOperationTimeout(5)  # 5 sec
    
    # Connect and subscribe to AWS IoT
    myAWSIoTMQTTClient.connect()
    myAWSIoTMQTTClient.subscribe(topic, 1, customCallback)
   

    while True:
        pass
