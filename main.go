package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type InstanceAction string

const (
	InstanceActionStart InstanceAction = "start"
	InstanceActionStop  InstanceAction = "stop"
)

func handleRequest(ctx context.Context, event events.CloudWatchEvent) error {
	log.Printf("Running maestro version: %s\n", "v1.0.0")
	m, err := NewMaestro(ctx, event)
	if err != nil {
		return err
	}

	if err := m.Validate(); err != nil {
		return err
	}

	switch m.action {
	case InstanceActionStart:
		return m.StartInstances(ctx)
	case InstanceActionStop:
		return m.StopInstances(ctx)
	}

	return nil
}

func main() {
	lambda.Start(handleRequest)
}

// ###### Maestro ######

type Maestro struct {
	ec2Client *ec2.Client
	action    InstanceAction
}

func NewMaestro(ctx context.Context, event events.CloudWatchEvent) (*Maestro, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK: %v", err)
	}

	return &Maestro{
		ec2Client: ec2.NewFromConfig(cfg),
		action:    InstanceAction(strings.ToLower(event.DetailType)),
	}, nil
}

func (m *Maestro) Validate() error {
	switch m.action {
	case InstanceActionStart, InstanceActionStop:
		return nil
	default:
		return fmt.Errorf("invalid action: %s", m.action)
	}
}

func (m *Maestro) getInstanceIDs(ctx context.Context) ([]string, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:ec2maestro"),
				Values: []string{"yes"},
			},
		},
	}

	result, err := m.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error describing instances: %v", err)
	}

	instanceIDs := make([]string, 0, len(result.Reservations))

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instanceIDs = append(instanceIDs, *instance.InstanceId)
		}
	}

	return instanceIDs, nil
}

func (m *Maestro) StartInstances(ctx context.Context) error {
	return m.handleInstanceAction(ctx, InstanceActionStart)
}

func (m *Maestro) StopInstances(ctx context.Context) error {
	return m.handleInstanceAction(ctx, InstanceActionStop)
}

func (m *Maestro) handleInstanceAction(ctx context.Context, action InstanceAction) error {
	instanceIDs, err := m.getInstanceIDs(ctx)
	if err != nil {
		return err
	}

	if len(instanceIDs) == 0 {
		log.Printf("No instances found for %s action", action)
		return nil
	}

	switch action {
	case InstanceActionStart:
		startInput := &ec2.StartInstancesInput{
			InstanceIds: instanceIDs,
		}
		_, err = m.ec2Client.StartInstances(ctx, startInput)
	case InstanceActionStop:
		stopInput := &ec2.StopInstancesInput{
			InstanceIds: instanceIDs,
		}
		_, err = m.ec2Client.StopInstances(ctx, stopInput)
	}

	if err != nil {
		return fmt.Errorf("error running %s action: %v", action, err)
	}

	log.Printf("Successfully executed %s action on %d instances", action, len(instanceIDs))

	return nil
}
