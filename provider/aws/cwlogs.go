package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/stelligent/mu/common"
	"time"
)

type logsManager struct {
	logsAPI cloudwatchlogsiface.CloudWatchLogsAPI
}

func newLogsManager(sess *session.Session) (common.LogsManager, error) {
	log.Debug("Connecting to CloudWatch Logs service")
	logsAPI := cloudwatchlogs.New(sess)

	return &logsManager{
		logsAPI: logsAPI,
	}, nil
}

// ViewLogs view the logs in CW
func (logsMgr *logsManager) ViewLogs(logGroup string, searchDuration time.Duration, follow bool, filter string, callback func(string, string, int64)) error {
	logsAPI := logsMgr.logsAPI

	startTime := time.Now().Add(-searchDuration).Unix() * 1000

	for {
		log.Debugf("Searching for logs in log_group '%s' after time '%d' and filter '%s'", logGroup, startTime, filter)

		params := &cloudwatchlogs.FilterLogEventsInput{
			StartTime:     aws.Int64(startTime + 1),
			Interleaved:   aws.Bool(true),
			LogGroupName:  aws.String(logGroup),
			FilterPattern: aws.String(filter),
		}

		err := logsAPI.FilterLogEventsPages(params,
			func(page *cloudwatchlogs.FilterLogEventsOutput, lastPage bool) bool {
				for _, event := range page.Events {
					ts := aws.Int64Value(event.Timestamp)
					if ts > startTime {
						startTime = ts
					}
					callback(aws.StringValue(event.LogStreamName), aws.StringValue(event.Message), ts)
				}
				return true
			})
		if err != nil {
			return err
		}

		if !follow {
			break
		}
		time.Sleep(5 * time.Second)
	}

	return nil
}
