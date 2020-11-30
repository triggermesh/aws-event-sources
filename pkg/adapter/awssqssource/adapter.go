/*
Copyright (c) 2019-2020 TriggerMesh Inc.

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

package awssqssource

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/triggermesh/aws-event-sources/pkg/adapter/common"
)

const (
	logfieldMsgID  = "msgID"
	logfieldMsgIDs = "msgIDs"
)

// envConfig is a set parameters sourced from the environment for the source's
// adapter.
type envConfig struct {
	pkgadapter.EnvConfig

	ARN string `envconfig:"ARN" required:"true"`
}

// adapter implements the source's adapter.
type adapter struct {
	logger *zap.SugaredLogger

	sqsClient sqsiface.SQSAPI
	ceClient  cloudevents.Client

	arn arn.ARN

	processQueue chan *sqs.Message
	deleteQueue  chan *sqs.Message

	deletePeriod time.Duration
}

// NewEnvConfig returns an accessor for the source's adapter envConfig.
func NewEnvConfig() pkgadapter.EnvConfigAccessor {
	return &envConfig{}
}

// NewAdapter returns a constructor for the source's adapter.
func NewAdapter(ctx context.Context, envAcc pkgadapter.EnvConfigAccessor, ceClient cloudevents.Client) pkgadapter.Adapter {
	logger := logging.FromContext(ctx)

	env := envAcc.(*envConfig)

	arn := common.MustParseARN(env.ARN)

	cfg := session.Must(session.NewSession(aws.NewConfig().
		WithRegion(arn.Region),
	))

	// allocate generous buffer sizes to avoid blocking often
	const batchSizePerProcMultiplier = 2
	queueBufferSize := runtime.GOMAXPROCS(-1) * maxReceiveMsgBatchSize * batchSizePerProcMultiplier

	return &adapter{
		logger: logger,

		sqsClient: sqs.New(cfg),
		ceClient:  ceClient,

		arn: arn,

		processQueue: make(chan *sqs.Message, queueBufferSize),
		deleteQueue:  make(chan *sqs.Message, queueBufferSize),

		deletePeriod: maxDeleteMsgPeriod,
	}
}

// Start implements adapter.Adapter.
func (a *adapter) Start(ctx context.Context) error {
	url, err := a.queueLookup(a.arn.Resource)
	if err != nil {
		a.logger.Errorw("Unable to find URL of SQS queue "+a.arn.Resource, zap.Error(err))
		return err
	}

	queueURL := *url.QueueUrl
	a.logger.Infof("Listening to SQS queue at URL: %s", queueURL)

	msgCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	for i := 0; i < runtime.GOMAXPROCS(-1); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.runMessagesReceiver(msgCtx, queueURL)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			a.runMessagesProcessor(msgCtx)
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.runMessagesDeleter(msgCtx, queueURL)
	}()

	<-ctx.Done()
	cancel()

	a.logger.Info("Waiting for message handlers to terminate")
	wg.Wait()

	return nil
}

// queueLookup finds the URL for a given queue name in the user's env.
// Needs to be an exact match to queue name and queue must be unique name in the AWS account.
func (a *adapter) queueLookup(queueName string) (*sqs.GetQueueUrlOutput, error) {
	return a.sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
}

// prettifyBatchResultErrors returns a pretty string representing a list of
// batch failures.
func prettifyBatchResultErrors(errs []*sqs.BatchResultErrorEntry) string {
	if len(errs) == 0 {
		return ""
	}

	var errStr strings.Builder

	errStr.WriteByte('[')

	for i, f := range errs {
		errStr.WriteString(f.String())
		if i+1 < len(errs) {
			errStr.WriteByte(',')
		}
	}

	errStr.WriteByte(']')

	return errStr.String()
}
