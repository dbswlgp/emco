// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	register "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/metrics"
	rpc "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/tracing"
	mtypes "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/types"
)

// Controller contains the parameters needed for Controllers
// It implements the interface for managing the Controllers
type Controller struct {
	Metadata mtypes.Metadata `json:"metadata"`
	Spec     ControllerSpec  `json:"spec"`
}

type ControllerSpec struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Type     string `json:"type"`
	Priority int    `json:"priority"`
}

const (
	MinControllerPriority            = 1
	MaxControllerPriority            = 1000000
	CONTROLLER_TYPE_ACTION    string = "action"
	CONTROLLER_TYPE_PLACEMENT string = "placement"
	CONTROLLER_TYPE_MIGRATION string = "migration"
)

var CONTROLLER_TYPES = [...]string{CONTROLLER_TYPE_ACTION, CONTROLLER_TYPE_PLACEMENT, CONTROLLER_TYPE_MIGRATION}

// ControllerKey is the key structure that is used in the database
type ControllerKey struct {
	ControllerName  string `json:"controller"`
	ControllerGroup string `json:"controllerGroup"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (mk ControllerKey) String() string {
	out, err := json.Marshal(mk)
	if err != nil {
		return ""
	}

	return string(out)
}

// ControllerManager is an interface exposes the Controller functionality
type ControllerManager interface {
	CreateController(ctx context.Context, ms Controller, mayExist bool) (Controller, error)
	GetController(ctx context.Context, name string) (Controller, error)
	GetControllers(ctx context.Context) ([]Controller, error)
	InitControllers(ctx context.Context)
	DeleteController(ctx context.Context, name string) error
}

// ControllerClient implements the Manager
// It will also be used to maintain some localized state
type ControllerClient struct {
	collectionName string
	tagMeta        string
	tagGroup       string
}

// ControllerServer implements a controller/manager service
type ControllerServer struct {
	ListenAndServe func() error
	Shutdown       func(context.Context) error
}

// NewControllerClient returns an instance of the ControllerClient
// which implements the Manager
func NewControllerClient(name, tag, group string) *ControllerClient {
	return &ControllerClient{
		collectionName: name,
		tagMeta:        tag,
		tagGroup:       group,
	}
}

// CreateController a new collection based on the Controller
func (mc *ControllerClient) CreateController(ctx context.Context, m Controller, mayExist bool) (Controller, error) {
	log.Info("CreateController .. start", log.Fields{"Controller": m, "exists": mayExist})

	// Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName:  m.Metadata.Name,
		ControllerGroup: mc.tagGroup,
	}

	// Check if this Controller already exists
	_, err := mc.GetController(ctx, m.Metadata.Name)
	if err == nil && !mayExist {
		return Controller{}, pkgerrors.New("Controller already exists")
	}

	err = db.DBconn.Insert(ctx, mc.collectionName, key, nil, mc.tagMeta, m)
	if err != nil {
		return Controller{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	// send message to create/update the  rpc connection
	rpc.UpdateRpcConn(m.Metadata.Name, m.Spec.Host, m.Spec.Port)

	log.Info("CreateController .. end", log.Fields{"Controller": m, "exists": mayExist})
	return m, nil
}

// GetController returns the Controller for corresponding name
func (mc *ControllerClient) GetController(ctx context.Context, name string) (Controller, error) {
	// Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName: name,
	}
	value, err := db.DBconn.Find(ctx, mc.collectionName, key, mc.tagMeta)
	if err != nil {
		return Controller{}, err
	} else if len(value) == 0 {
		return Controller{}, pkgerrors.New("Controller not found")
	}

	if value != nil {
		microserv := Controller{}
		err = db.DBconn.Unmarshal(value[0], &microserv)
		if err != nil {
			return Controller{}, err
		}
		return microserv, nil
	}

	return Controller{}, pkgerrors.New("Unknown Error")
}

// GetControllers returns all the  Controllers that are registered
func (mc *ControllerClient) GetControllers(ctx context.Context) ([]Controller, error) {
	// Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName:  "",
		ControllerGroup: mc.tagGroup,
	}

	var resp []Controller
	values, err := db.DBconn.Find(ctx, mc.collectionName, key, mc.tagMeta)
	if err != nil {
		return []Controller{}, err
	}

	for _, value := range values {
		microserv := Controller{}
		err = db.DBconn.Unmarshal(value, &microserv)
		if err != nil {
			return []Controller{}, err
		}

		resp = append(resp, microserv)
	}

	return resp, nil
}

// DeleteController the  Controller from database
func (mc *ControllerClient) DeleteController(ctx context.Context, name string) error {
	// Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName:  name,
		ControllerGroup: mc.tagGroup,
	}
	err := db.DBconn.Remove(ctx, mc.collectionName, key)
	if err != nil {
		return err
	}

	// send message to close rpc connection
	rpc.RemoveRpcConn(name)

	return nil
}

// InitControllers initializes connctions for controllers in the DB
func (mc *ControllerClient) InitControllers(ctx context.Context) {
	vals, _ := mc.GetControllers(ctx)
	for _, v := range vals {
		log.Info("Initializing RPC connection for controller", log.Fields{
			"Controller": v.Metadata.Name,
		})
		rpc.UpdateRpcConn(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
	}
}

func NewControllerServer(name string, httpRouter *mux.Router, grpcServer *register.GrpcServer) (*ControllerServer, error) {
	if httpRouter == nil && grpcServer == nil {
		return nil, errors.New("NewControllerServer: must provide non-nil httpRouter or grpcServer")
	}

	err := tracing.InitializeTracer()
	if err != nil {
		return nil, errors.New("Unable to initialize tracing")
	}

	prometheus.MustRegister(metrics.NewBuildInfoCollector(name))

	httpServerPort := config.GetConfiguration().ServicePort
	fmt.Println(config.GetConfiguration())
	fmt.Println(httpServerPort)
	if httpServerPort == "" {
		return nil, errors.New("NewControllerServer: must configure a \"service-port\"")
	}
	if httpRouter == nil {
		httpRouter = mux.NewRouter()
	}
	httpRouter.Use(tracing.Middleware)
	httpServer, err := newHttpServer(httpServerPort, httpRouter)
	if err != nil {
		log.Error("Unable to create HTTP server", log.Fields{"Error": err})
		return nil, err
	}

	return &ControllerServer{
		ListenAndServe: func() error {
			log.Info("Starting HTTP server", log.Fields{"Port": httpServerPort})
			httpLis, err := net.Listen("tcp", ":"+httpServerPort)
			if err != nil {
				log.Error("Could not listen on HTTP port", log.Fields{"Error": err, "Port": httpServerPort})
				return err
			}

			var grpcLis net.Listener
			if grpcServer != nil {
				log.Info("Starting gRPC server", log.Fields{"Port": grpcServer.Port})
				grpcLis, err = net.Listen("tcp", fmt.Sprintf(":%d", grpcServer.Port))
				if err != nil {
					log.Error("Could not listen on gRPC port", log.Fields{"Error": err, "Port": grpcServer.Port})
					return err
				}
			}

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				err = httpServer.Serve(httpLis)
				if err != nil {
					log.Error("HTTP server stopped", log.Fields{"Error": err})
				}
				wg.Done()
			}()
			if grpcServer != nil {
				wg.Add(1)
				go func() {
					err = grpcServer.Serve(grpcLis)
					if err != nil {
						log.Error("gRPC server stopped", log.Fields{"Error": err})
					}
					wg.Done()
				}()
			}
			wg.Wait()
			return nil
		},
		Shutdown: func(ctx context.Context) error {
			if grpcServer != nil {
				err = grpcServer.Shutdown(ctx)
				if err != nil {
					log.Error("gRPC server shutdown failed", log.Fields{"Error": err})
				}
			}
			err = httpServer.Shutdown(ctx)
			if err != nil {
				log.Error("HTTP server shutdown failed", log.Fields{"Error": err})
			}
			return err
		},
	}, nil
}

func newHttpServer(port string, httpRouter *mux.Router) (*http.Server, error) {
	httpRouter.Handle("/metrics", promhttp.Handler())
	loggedRouter := handlers.LoggingHandler(os.Stdout, httpRouter)

	return &http.Server{
		Handler: loggedRouter,
		Addr:    ":" + port,
	}, nil
}
