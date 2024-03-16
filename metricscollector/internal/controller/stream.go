// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

package controller

import (
	"encoding/json"
	events "gitlab.com/project-emco/core/emco-base/src/policy/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/klog/v2"
	"net"
)

const (
	InAddrAny = "0.0.0.0"
)

// MetricsServer initializes the grpc server.
// Each agent instance act as a grpc server, and emco policy controller connects to them
// when the agent is registered.
func MetricsServer(metricsStream <-chan Measurements, agentId string, port string) {
	lis, err := net.Listen("tcp", InAddrAny+":"+port)
	if err != nil {
		klog.Fatalf("Failed to listen %s", err.Error())
	}
	s := grpc.NewServer()
	server := &StreamServer{
		metricsStream: metricsStream,
		agentId:       agentId,
	}
	events.RegisterEventsServer(s, server)
	err = s.Serve(lis)
	if err != nil {
		klog.Errorf("metricsServer Failed to serve: %s", err.Error())
		return
	}
}

// EventUpdate is method that can be invoked over rpc to get the metrics.
// Metrics updates are streamed to the central controller (caller)
func (s StreamServer) EventUpdate(_ *events.ServerSpec, updateServer events.Events_EventUpdateServer) error {
	agentSpec := &events.AgentSpec{AgentId: s.agentId}
	spec, err := anypb.New(agentSpec)
	if err != nil {
		klog.Errorf("Converting agent spec failed: %s", err.Error())
		return err
	}
	// Stream every metrics fetched to central controller.
	// Metrics are enriched with agent and deployment information.
	for metricList := range s.metricsStream {
		metricListJson, err := json.Marshal(metricList)
		if err != nil {
			klog.Errorf("Encoding metrics failed (Ignored measurement): %s", err.Error())
			continue
		}
		err = updateServer.Send(&events.Event{
			AgentId:    s.agentId,
			EventId:    metricList.Metric,
			Spec:       spec,
			ContextId:  metricList.ContextId,
			AppName:    metricList.AppName,
			MetricList: metricListJson,
		})

		if err != nil {
			klog.Errorf("Stream send failed: %s", err.Error())
			return err
		}
	}
	return nil
}
