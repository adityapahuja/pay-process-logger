package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

type MockDoType func(req *http.Request) (*http.Response, error)

type MockClient struct {
	MockDo MockDoType
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

func TestGetTaskDetails(t *testing.T) {
	os.Setenv(ECS_CONTAINER_METADATA_URI_V4, "http://169.254.170.2/v4")

	jsonResponse := `{
		"Cluster": "arn:aws:ecs:eu-west-1:111111111111:cluster/development-fargate-fg",
		"TaskARN": "arn:aws:ecs:eu-west-1:111111111111:task/development-fargate-fg/fa12345678901234567890abcdefabcd",
		"Family": "development-fargate_application",
		"Revision": "14",
		"DesiredStatus": "RUNNING",
		"KnownStatus": "RUNNING",
		"Limits": {
		  "CPU": 1,
		  "Memory": 2048
		},
		"PullStartedAt": "2021-05-07T18:25:05.438653698Z",
		"PullStoppedAt": "2021-05-07T18:25:27.824805785Z",
		"AvailabilityZone": "eu-west-1c",
		"Containers": [
		  {
			"DockerId": "fa12345678901234567890abcdefabcd-4000290907",
			"Name": "telegraf",
			"DockerName": "telegraf",
			"Image": "111111111111.dkr.ecr.eu-west-1.amazonaws.com/magic/telegraf:latest-master",
			"ImageID": "sha256:32b84c28d9427c2a37bd670943f932c433194fe7dc253fdbebc0205ee2087a9e",
			"Labels": {
			  "com.amazonaws.ecs.cluster": "arn:aws:ecs:eu-west-1:111111111111:cluster/development-fargate-fg",
			  "com.amazonaws.ecs.container-name": "telegraf",
			  "com.amazonaws.ecs.task-arn": "arn:aws:ecs:eu-west-1:111111111111:task/development-fargate-fg/fa12345678901234567890abcdefabcd",
			  "com.amazonaws.ecs.task-definition-family": "development-fargate_application",
			  "com.amazonaws.ecs.task-definition-version": "14"
			},
			"DesiredStatus": "RUNNING",
			"KnownStatus": "RUNNING",
			"Limits": {
			  "CPU": 256,
			  "Memory": 512
			},
			"CreatedAt": "2021-05-07T18:25:33.305851504Z",
			"StartedAt": "2021-05-07T18:25:33.305851504Z",
			"Type": "NORMAL",
			"Networks": [
			  {
				"NetworkMode": "awsvpc",
				"IPv4Addresses": [
				  "111.11.111.21"
				],
				"AttachmentIndex": 0,
				"MACAddress": "00:00:fb:00:00:c0",
				"IPv4SubnetCIDRBlock": "111.11.111.0/24",
				"DomainNameServers": [
				  "111.11.1.2"
				],
				"DomainNameSearchList": [
				  "eu-west-1.compute.internal"
				],
				"PrivateDNSName": "ip-111-11-111-21.eu-west-1.compute.internal",
				"SubnetGatewayIpv4Address": "111.11.111.1/24"
			  }
			],
			"ContainerARN": "arn:aws:ecs:eu-west-1:111111111111:container/development-fargate-fg/fa12345678901234567890abcdefabcd/111111fa-abe1-11fc-b11c-ab1eca1d11ad",
			"LogOptions": {
			  "awslogs-group": "telegraf",
			  "awslogs-region": "eu-west-1",
			  "awslogs-stream": "development-fargate-application/telegraf/fa12345678901234567890abcdefabcd"
			},
			"LogDriver": "awslogs"
		  }
		],
		"LaunchType": "FARGATE"
	}`

	r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
	Client = &MockClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	}

	task, err := GetTaskDetails()
	if err != nil {
		t.Error("TestGetTaskDetails failed.")
		return
	}

	if task.family != "development-fargate_application" {
		t.Error("TestGetTaskDetails failed, the task.family does not contain \"development-fargate_application\".")
		return
	}

	if task.taskId != "fa12345678901234567890abcdefabcd" {
		t.Error("TestGetTaskDetails failed, the task.taskId does not contain \"fa12345678901234567890abcdefabcd\".")
		return
	}
}

func TestGetContainerDetails(t *testing.T) {
	os.Setenv(ECS_CONTAINER_METADATA_URI_V4, "http://169.254.170.2/v4")

	jsonResponse := `{
		"DockerId": "abcdef01234567890123456789012345-1234567890",
		"Name": "application",
		"DockerName": "application",
		"Image": "111111111111.dkr.ecr.eu-west-1.amazonaws.com/magic/application:latest-master",
		"ImageID": "sha256:1b111111d1ce1b111a1111db1111e11111c111a1dbfd111111ceaee111ee111f",
		"Labels": {
		  "com.amazonaws.ecs.cluster": "arn:aws:ecs:eu-west-1:111111111111:cluster/development-fargate-fg",
		  "com.amazonaws.ecs.container-name": "application",
		  "com.amazonaws.ecs.task-arn": "arn:aws:ecs:eu-west-1:111111111111:task/development-fargate-fg/abcdef01234567890123456789012345",
		  "com.amazonaws.ecs.task-definition-family": "development-fargate_application_FG",
		  "com.amazonaws.ecs.task-definition-version": "14"
		},
		"DesiredStatus": "RUNNING",
		"KnownStatus": "RUNNING",
		"Limits": {
		  "CPU": 512,
		  "Memory": 1024
		},
		"CreatedAt": "2021-05-07T18:25:33.519661446Z",
		"StartedAt": "2021-05-07T18:25:33.519661446Z",
		"Type": "NORMAL",
		"Networks": [
		  {
			"NetworkMode": "awsvpc",
			"IPv4Addresses": [
			  "111.11.111.21"
			],
			"AttachmentIndex": 0,
			"MACAddress": "11:11:11:11:11:11",
			"IPv4SubnetCIDRBlock": "111.11.111.0/24",
			"DomainNameServers": [
			  "111.11.0.2"
			],
			"DomainNameSearchList": [
			  "eu-west-1.compute.internal"
			],
			"PrivateDNSName": "ip-111-11-111-21.eu-west-1.compute.internal",
			"SubnetGatewayIpv4Address": "111.11.111.1/24"
		  }
		],
		"ContainerARN": "arn:aws:ecs:eu-west-1:111111111111:container/development-fargate-fg/abcdef01234567890123456789012345/11f1111d-1b11-1ec1-a11d-1c11b11111b1",
		"LogOptions": {
		  "awslogs-group": "application",
		  "awslogs-region": "eu-west-1",
		  "awslogs-stream": "development-fargate/application/abcdef01234567890123456789012345"
		},
		"LogDriver": "awslogs"
	  }`

	r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
	Client = &MockClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	}

	container, err := GetContainerDetails()
	if err != nil {
		t.Error("TestGetContainerDetails failed.")
		return
	}

	if container.containerName != "application" {
		t.Error("TestGetContainerDetails failed, the container.containerName does not contain \"application\".")
		return
	}

	if container.region != "eu-west-1" {
		t.Error("TestGetContainerDetails failed, the container.region does not contain \"eu-west-1\".")
		return
	}
}
