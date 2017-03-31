package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type mockedCwLogs struct {
	mock.Mock
	cloudwatchlogsiface.CloudWatchLogsAPI
}
func (m *mockedCwLogs) FilterLogEventsPages(input *cloudwatchlogs.FilterLogEventsInput, cb func(*cloudwatchlogs.FilterLogEventsOutput, bool) bool) error {
	args := m.Called(input, cb)
	return args.Error(0)
}


func TestLogsManager_ViewLogs(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedCwLogs)
	m.On("FilterLogEventsPages", mock.AnythingOfType("*cloudwatchlogs.FilterLogEventsInput"), mock.AnythingOfType("func(*cloudwatchlogs.FilterLogEventsOutput, bool) bool")).
		Return(nil).
		Run(func(args mock.Arguments) {
			cb := args.Get(1).(func(*cloudwatchlogs.FilterLogEventsOutput, bool) bool)
			cb(&cloudwatchlogs.FilterLogEventsOutput{
				Events: []*cloudwatchlogs.FilteredLogEvent {
					{
						Message: aws.String("hello world"),
					},
					{
						Message: aws.String("hello agains"),
					},
				},
			}, true)
		})

	lm := logsManager{
		logsAPI: m,
	}

	events := 0
	cb := func(loggroup string, message string, ts int64) {
		events ++
	}

	err := lm.ViewLogs("foo", false, "", cb)
	assert.Nil(err)
	assert.Equal(2, events)

	m.AssertExpectations(t)
	m.AssertNumberOfCalls(t, "FilterLogEventsPages", 1)
}

