package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type Task struct {
	family string
	taskId string
}

type Container struct {
	containerName string
	region        string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client HTTPClient
)

func init() {
	Client = &http.Client{}
}

const ECS_CONTAINER_METADATA_URI_V4 = "ECS_CONTAINER_METADATA_URI_V4"

func GetTaskDetails() (Task, error) {
	var metadata_uri = os.Getenv(ECS_CONTAINER_METADATA_URI_V4) + "/task"

	request, err := http.NewRequest(http.MethodGet, metadata_uri, nil)
	if err != nil {
		return Task{}, err
	}

	response, err := Client.Do(request)
	if err != nil {
		return Task{}, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return Task{}, err
		}

		var task map[string]interface{}

		err = json.Unmarshal(bodyBytes, &task)
		if err != nil {
			panic(err)
		}
		var t Task
		t.family = task["Family"].(string)
		t.taskId = strings.Split(task["TaskARN"].(string), "/")[2]

		return t, nil
	}
	return Task{}, errors.New("Failed to get task details from the task metadata endpoint " + metadata_uri)
}

func GetContainerDetails() (Container, error) {
	var metadata_uri = os.Getenv(ECS_CONTAINER_METADATA_URI_V4)

	request, err := http.NewRequest(http.MethodGet, metadata_uri, nil)
	if err != nil {
		return Container{}, err
	}

	response, err := Client.Do(request)
	if err != nil {
		return Container{}, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return Container{}, err
		}

		var task map[string]interface{}

		err = json.Unmarshal(bodyBytes, &task)
		if err != nil {
			panic(err)
		}
		var c Container
		c.containerName = task["Name"].(string)
		c.region = strings.Split(task["ContainerARN"].(string), ":")[3]

		return c, nil
	}
	return Container{}, errors.New("Failed to get container details from the task metadata endpoint " + metadata_uri)
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func main() {
	if os.Getenv("ENVIRONMENT") == "" {
		log.Fatal("The environment variable ENVIRONMENT cannot be empty.")
		os.Exit(1)
	}

	sleepTime, err := strconv.Atoi(os.Getenv("PROCESS_LOGGER_SLEEP_TIME"))
	if err != nil {
		log.Fatal("The environment variable PROCESS_LOGGER_SLEEP_TIME should contain number only.")
		os.Exit(2)
	}

	task, err := GetTaskDetails()
	if err != nil {
		log.Fatal("Failed to get task details. Error: " + err.Error())
		os.Exit(3)
	}

	container, err := GetContainerDetails()
	if err != nil {
		log.Fatal("Failed to get container details. Error: " + err.Error())
		os.Exit(4)
	}

	processLogGroup := os.Getenv("ENVIRONMENT") + "__process_logger"
	logStream := task.family + "/" + task.taskId + "/" + container.containerName

	mySession := session.Must(session.NewSession())
	service := cloudwatchlogs.New(mySession, aws.NewConfig().WithRegion(container.region))

	response, err := service.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(processLogGroup),
		LogStreamNamePrefix: aws.String(logStream),
	})
	if err != nil {
		log.Fatal("Failed to describe log streams. Error: " + err.Error())
		os.Exit(5)
	}

	sequenceToken := ""
	if len(response.LogStreams) > 0 {
		sequenceToken = *response.LogStreams[0].UploadSequenceToken
	}

	for {
		psOutput, err := exec.Command("ps", "aux").Output()
		if err != nil {
			log.Fatal("Failed to execute 'ps aux' command. Error: " + err.Error())
			os.Exit(6)
		}

		event := &cloudwatchlogs.InputLogEvent{
			Timestamp: aws.Int64(makeTimestamp()),
			Message:   aws.String(string(psOutput)),
		}

		if len(sequenceToken) == 0 {
			_, err := service.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
				LogGroupName:  aws.String(processLogGroup),
				LogStreamName: aws.String(logStream),
			})
			if err != nil {
				log.Fatal("Failed to create log stream. Error: " + err.Error())
				os.Exit(7)
			}
			putLogEventsResp, err := service.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
				LogEvents:     []*cloudwatchlogs.InputLogEvent{event},
				LogGroupName:  aws.String(processLogGroup),
				LogStreamName: aws.String(logStream),
			})
			if err != nil {
				log.Fatal("Failed to put log events first time. Error: " + err.Error())
				os.Exit(8)
			}
			sequenceToken = *putLogEventsResp.NextSequenceToken
		} else {
			putLogEventsResp, err := service.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
				LogEvents:     []*cloudwatchlogs.InputLogEvent{event},
				LogGroupName:  aws.String(processLogGroup),
				LogStreamName: aws.String(logStream),
				SequenceToken: aws.String(sequenceToken),
			})
			if err != nil {
				log.Fatal("Failed to put log events. Error: " + err.Error())
				os.Exit(9)
			}
			sequenceToken = *putLogEventsResp.NextSequenceToken
		}
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
}
