// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2022 Intel Corporation

package statusnotifyserver

import (
	"context"
	"io"
	"sync"
	"time"

	pkgerrors "github.com/pkg/errors"
	inc "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/installappclient"
	pb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/statusnotify"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/status"
	readynotifypb "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/readynotify"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	proto "google.golang.org/protobuf/proto"
)

// StatusNotifyServerHelpers is an interface supported by the specific microservices that need to provide status notification
type StatusNotifyServerHelpers interface {
	GetAppContextId(ctx context.Context, reg *pb.StatusRegistration) (string, error)
	StatusQuery(ctx context.Context, reg *pb.StatusRegistration, qInstance, qType, qOutput string, qApps, qClusters, qResources []string) status.StatusResult
	PrepareStatusNotification(reg *pb.StatusRegistration, statusResult status.StatusResult) *pb.StatusNotification
}

// streamInfo contains information about a given status notification stream, including:
// the stream server, and information about the type of notifications desired and
// filter and output details.
type streamInfo struct {
	stream       pb.StatusNotify_StatusRegisterServer
	reg          *pb.StatusRegistration
	appContextID string
	lastNotif    *pb.StatusNotification
}

type filters struct {
	qOutputSummary bool
	apps           map[string]struct{}
	clusters       map[string]struct{}
	resources      map[string]struct{}
}

type appContextInfo struct {
	readyNotifyStream readynotifypb.ReadyNotify_AlertClient
	statusClientIDs   map[string]struct{}
	queryFilters      map[string]filters // up to two entries:  "ready", "deployed"
}

// StatusNotifyServer will be initialized by NewStatusNotifyServer() and
// its lifecycle is valid until all the clients unsubscribed the stream notification channel
type StatusNotifyServer struct {
	name string
	// clientId is expected to be unique.  Registering with a clientId that is already in the
	// map will be rejected - i.e. the new stream will close immediately with an error
	appContexts       map[string]appContextInfo // map[appcontextid]appContextInfo
	statusClients     map[string]streamInfo     // map[clientId]streamInfo
	streamChannels    map[pb.StatusNotify_StatusRegisterServer]chan int
	sh                StatusNotifyServerHelpers
	readyNotifyClient readynotifypb.ReadyNotifyClient
	mutex             sync.Mutex
}

var notifServer *StatusNotifyServer

func updateQueryFilters(clientId string, deleteFlag bool) {
	appContextID := notifServer.statusClients[clientId].appContextID

	deployedOutputSummary := true
	deployedApps := make(map[string]struct{})
	deployedClusters := make(map[string]struct{})
	deployedResources := make(map[string]struct{})
	readyOutputSummary := true
	readyApps := make(map[string]struct{})
	readyClusters := make(map[string]struct{})
	readyResources := make(map[string]struct{})

	// run through all registrations for the appContextID and compile current set of query filters
	readyCnt := 0
	deployedCnt := 0
	for client, si := range notifServer.statusClients {
		if deleteFlag && client == clientId {
			continue
		}
		if _, ok := notifServer.appContexts[appContextID].statusClientIDs[client]; !ok {
			continue
		}
		ready := true
		if si.reg.StatusType == pb.StatusValue_DEPLOYED {
			ready = false
		}
		if si.reg.Output == pb.OutputType_ALL {
			if ready {
				readyOutputSummary = false
			} else {
				deployedOutputSummary = false
			}
		}

		if ready {
			if readyCnt == 0 || (len(si.reg.Apps) > 0 && len(readyApps) > 0) {
				for _, r := range si.reg.Apps {
					readyApps[r] = struct{}{}
				}
			} else {
				if readyCnt > 0 && len(readyApps) > 0 {
					readyApps = make(map[string]struct{})
				}
			}
			if readyCnt == 0 || (len(si.reg.Clusters) > 0 && len(readyClusters) > 0) {
				for _, r := range si.reg.Clusters {
					readyClusters[r] = struct{}{}
				}
			} else {
				if readyCnt > 0 && len(readyClusters) > 0 {
					readyClusters = make(map[string]struct{})
				}
			}
			if readyCnt == 0 || (len(si.reg.Resources) > 0 && len(readyResources) > 0) {
				for _, r := range si.reg.Resources {
					readyResources[r] = struct{}{}
				}
			} else {
				if readyCnt > 0 && len(readyResources) > 0 {
					readyResources = make(map[string]struct{})
				}
			}
			readyCnt++
		} else {
			if deployedCnt == 0 || (len(si.reg.Apps) > 0 && len(deployedApps) > 0) {
				for _, r := range si.reg.Apps {
					deployedApps[r] = struct{}{}
				}
			} else {
				if deployedCnt > 0 && len(deployedApps) > 0 {
					deployedApps = make(map[string]struct{})
				}
			}
			if deployedCnt == 0 || (len(si.reg.Clusters) > 0 && len(deployedClusters) > 0) {
				for _, r := range si.reg.Clusters {
					deployedClusters[r] = struct{}{}
				}
			} else {
				if deployedCnt > 0 && len(deployedClusters) > 0 {
					deployedClusters = make(map[string]struct{})
				}
			}
			if deployedCnt == 0 || (len(si.reg.Resources) > 0 && len(deployedResources) > 0) {
				for _, r := range si.reg.Resources {
					deployedResources[r] = struct{}{}
				}
			} else {
				if deployedCnt > 0 && len(deployedResources) > 0 {
					deployedResources = make(map[string]struct{})
				}
			}
			deployedCnt++
		}
	}
	updatedFilters := make(map[string]filters)
	updatedFilters["ready"] = filters{
		qOutputSummary: readyOutputSummary,
		apps:           readyApps,
		clusters:       readyClusters,
		resources:      readyResources,
	}
	updatedFilters["deployed"] = filters{
		qOutputSummary: deployedOutputSummary,
		apps:           deployedApps,
		clusters:       deployedClusters,
		resources:      deployedResources,
	}
	acInfo := notifServer.appContexts[appContextID]
	acInfo.queryFilters = updatedFilters
	notifServer.appContexts[appContextID] = acInfo
}

func queryNeeded(qType string, apps, clusters map[string]struct{}, acInfo appContextInfo) (bool, string, []string, []string, []string) {
	doQuery := false
	var qOutput string
	qApps := make([]string, 0)
	qClusters := make([]string, 0)
	qResources := make([]string, 0)

	if len(apps) == 0 && len(clusters) == 0 {
		return false, "", qApps, qClusters, qResources
	}

	filters := acInfo.queryFilters[qType]
	if len(filters.apps) == 0 {
		doQuery = true
	} else {
		for app, _ := range apps {
			if _, ok := filters.apps[app]; ok {
				for a, _ := range filters.apps {
					qApps = append(qApps, a)
				}
				doQuery = true
				break
			}
		}
	}
	if !doQuery {
		return false, "", qApps, qClusters, qResources
	}

	if len(filters.clusters) == 0 {
		doQuery = true
	} else {
		for cluster, _ := range clusters {
			if _, ok := filters.clusters[cluster]; ok {
				for c, _ := range filters.clusters {
					qClusters = append(qClusters, c)
				}
				doQuery = true
				break
			}
		}
	}

	for r, _ := range filters.resources {
		qResources = append(qResources, r)
	}
	if filters.qOutputSummary {
		qOutput = "summary"
	} else {
		qOutput = "all"
	}
	return doQuery, qOutput, qApps, qClusters, qResources
}

// StatusRegister gets notified when a client registers for a status notification stream for a given resource
func (s *StatusNotifyServer) StatusRegister(reg *pb.StatusRegistration, stream pb.StatusNotify_StatusRegisterServer) error {
	ctx := stream.Context()

	// Check if the clientId is already in use, return error if yes
	clientId := reg.GetClientId()
	if len(clientId) == 0 {
		log.Info("[StatusNotify gRPC] Recieved a status notification registration with invalid client ID", log.Fields{})
		return pkgerrors.New("Invalid client ID")
	}
	if _, ok := s.statusClients[clientId]; ok {
		log.Info("[StatusNotify gRPC] Recieved a duplicate status notification registration",
			log.Fields{"client": clientId})
		return pkgerrors.New("Duplicate client ID: " + clientId)
	}
	appContextID, err := s.sh.GetAppContextId(ctx, reg)
	if err != nil {
		log.Info("[StatusNotify gRPC] Could not get appContextID for status notification registration",
			log.Fields{"client": clientId, "AppContextID": appContextID})
		return err
	}

	log.Info("[StatusNotify gRPC] Recieved a status notification registration",
		log.Fields{"client": clientId, "appContextID": appContextID})

	// Add the client info to the statusnotify server maps
	needReadyNotifyStream := false
	s.mutex.Lock()

	// update appContexts
	if _, ok := s.appContexts[appContextID]; !ok {
		s.appContexts[appContextID] = appContextInfo{
			readyNotifyStream: nil,
			statusClientIDs:   make(map[string]struct{}),
			queryFilters:      make(map[string]filters),
		}
		needReadyNotifyStream = true
		log.Trace("[StatusNotify gRPC] Adding appContextInfo, need Ready Notify Stream",
			log.Fields{"appContextID": appContextID, "client": clientId})
	}
	s.appContexts[appContextID].statusClientIDs[clientId] = struct{}{}

	// update statusClients
	s.statusClients[clientId] = streamInfo{
		stream:       stream,
		reg:          reg,
		appContextID: appContextID,
	}

	updateQueryFilters(clientId, false)

	// update streamChannels
	s.streamChannels[stream] = make(chan int)
	c := s.streamChannels[stream]

	var wg sync.WaitGroup

	if needReadyNotifyStream {
		s.readyNotifyClient = newReadyNotifyClient(ctx) // always call this - rsync controller could have been deleted/re-added
		if s.readyNotifyClient == nil {
			s.mutex.Unlock()
			log.Error("[StatusNotify gRPC] Could not get ReadyNotify Client",
				log.Fields{"appContextID": appContextID, "client": clientId})
			return pkgerrors.Errorf("Unable to get ReadyNotifyClient for StatusNotifyServer: %v, %v", appContextID, clientId)
		}

		// The readyNotifyClient ctx used below belongs to the
		// stream, so we must create a new (not derived)
		// context to prevent the context from being cancelled
		// when the caller completes: a cancelled context will
		// cause the below work to exit early.  A link is used
		// so that the traces can be associated.
		tracer := otel.Tracer("orchestrator")
		readyNotifyStreamCtx, span := tracer.Start(context.Background(), "sendStatusNotifications",
			trace.WithLinks(trace.LinkFromContext(ctx)),
		)

		readyNotifyStream, err := s.readyNotifyClient.Alert(readyNotifyStreamCtx,
			&readynotifypb.Topic{ClientName: s.name, AppContext: appContextID})
		if err != nil {
			s.mutex.Unlock()
			log.Error("[StatusNotify gRPC] Could not get ReadyNotify Stream",
				log.Fields{"appContextID": appContextID, "client": clientId, "error": err})
			span.End()
			return err
		}

		// set the readyNotifyStream for this appContextID - check for a race - no need to set if it's already been set
		acInfo := s.appContexts[appContextID]
		acInfo.readyNotifyStream = readyNotifyStream
		s.mutex.Unlock()

		log.Info("[StatusNotify gRPC] ready to start sending status notifications",
			log.Fields{"appContextID": appContextID, "client": clientId})
		wg.Add(1)
		go sendStatusNotifications(readyNotifyStreamCtx, readyNotifyStream, &wg, appContextID)
	} else {
		s.mutex.Unlock()
	}

	// Keep stream open
	for {
		select {
		case <-ctx.Done():
			log.Info("[StatusNotify gRPC] Client has disconnected", log.Fields{"client": clientId})
			cleanup(ctx, clientId)
			wg.Wait()
			return nil
		case <-c:
			log.Info("[StatusNotify gRPC] Stop channel has been triggered for client", log.Fields{"client": clientId})
			wg.Wait()
			return nil
		default:
		}
	}
}

// SendStatusNotification sends a status notification message to the subscriber
func sendStatusNotifications(ctx context.Context, stream readynotifypb.ReadyNotify_AlertClient, wg *sync.WaitGroup, appContextID string) error {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	type recvEvent struct {
		app     string
		cluster string
		err     error
	}
	rChan := make(chan recvEvent)
	tChan := make(chan bool)

	// start go thread to receive from the stream
	go func() {
		for true {
			resp, err := stream.Recv()
			if err != nil {
				rChan <- recvEvent{err: err}
				return
			}
			rChan <- recvEvent{app: resp.App, cluster: resp.Cluster, err: nil}
		}
	}()

	tLast := time.Now()
	timerRunning := false
	var apps map[string]struct{}
	var clusters map[string]struct{}

	// each pass through the loop handles one set of events, or exits
	for true {
		if !timerRunning {
			apps = make(map[string]struct{})
			clusters = make(map[string]struct{})
		}

		select {
		case event := <-rChan:
			if event.err != nil {
				if event.err == io.EOF {
					log.Error("[StatusNotify gRPC] ReadyNotify stream closed due to EOF", log.Fields{"err": event.err})
				} else {
					// some other error - figure out how to reconnect ?
					log.Error("[StatusNotify gRPC] Failed to receive notification", log.Fields{"err": event.err})
				}
				wg.Done()
				return event.err
			}
			apps[event.app] = struct{}{}
			clusters[event.cluster] = struct{}{}
			log.Trace("[StatusNotify gRPC] Accumulating monitor events", log.Fields{"app": event.app, "cluster": event.cluster})

			tNow := time.Now()
			if tNow.Sub(tLast) > 3*time.Second {
				tLast = tNow
			} else if !timerRunning {
				go func() {
					time.Sleep(tNow.Sub(tLast))
					tChan <- true
				}()
				log.Trace("[StatusNotify gRPC] setting timer", log.Fields{"time": tNow.Sub(tLast)})
				timerRunning = true
			}
		case <-tChan:
			log.Trace("[StatusNotify gRPC] timeout done", log.Fields{"apps": apps, "clusters": clusters})
			timerRunning = false
			tLast = time.Now()
		}

		if timerRunning {
			continue
		}
		log.Trace("[StatusNotify gRPC] handling status events", log.Fields{"apps": apps, "clusters": clusters})

		notifServer.mutex.Lock()
		acInfo, ok := notifServer.appContexts[appContextID]
		if !ok {
			notifServer.mutex.Unlock()
			log.Warn("[StatusNotify gRPC] Received a ReadyNotify alert from rsync for missing appContext", log.Fields{"appContextID": appContextID, "apps": apps, "clusters": clusters})
			continue
		}

		// loop through each type of status query
		// do one query for each type that will satisfy all registrations as well as the set of events which
		// have occurred
		// prepare and send notifications for all registrations
		for _, qType := range []string{"ready", "deployed"} {
			doQuery, qOutput, qApps, qClusters, qResources := queryNeeded(qType, apps, clusters, acInfo)

			if !doQuery {
				continue
			}

			// get any client registration for this appContextID - this will just be used to get the key elements
			// the key elements are expected to be identical for all regsitrations that resolve to this appContextID
			var reg *pb.StatusRegistration
			gotReg := false
			for clientId, _ := range acInfo.statusClientIDs {
				si := notifServer.statusClients[clientId]
				reg = si.reg
				gotReg = true
				break
			}
			if !gotReg {
				log.Error("[StatusNotify gRPC] Status registration not found", log.Fields{"appContextID": appContextID})
				continue
			}

			statusResult := notifServer.sh.StatusQuery(ctx, reg, appContextID, qType, qOutput, qApps, qClusters, qResources)

			// For a given alert, send a status notification to each status client watching the appcontextId
			for clientId, _ := range acInfo.statusClientIDs {
				si := notifServer.statusClients[clientId]
				if (si.reg.StatusType == pb.StatusValue_READY && qType != "ready") ||
					(si.reg.StatusType == pb.StatusValue_DEPLOYED && qType != "deployed") {
					continue
				}

				notification := notifServer.sh.PrepareStatusNotification(si.reg, statusResult)
				if si.lastNotif != nil && proto.Equal(si.lastNotif, notification) {
					log.Trace("[StatusNotify gRPC] Status notification equal to last notification - skipping",
						log.Fields{"clientId": clientId, "appContextID": appContextID})
					continue
				}
				si.lastNotif = notification
				notifServer.statusClients[clientId] = si
				err := si.stream.Send(notification)
				if err != nil {
					log.Error("[StatusNotify gRPC] Status notification failed to be sent", log.Fields{"clientId": clientId, "appContextID": appContextID, "err": err})
				}
				log.Trace("[StatusNotify gRPC] Status notification sent",
					log.Fields{"clientId": clientId, "appContextID": appContextID})
			}
		}
		notifServer.mutex.Unlock()

	}
	return nil
}

// cleanup will be called when the subscriber wants to terminate the stream
func cleanup(ctx context.Context, clientId string) {
	notifServer.mutex.Lock()
	defer notifServer.mutex.Unlock()

	// get the clientId entry, stop the stream
	si, ok := notifServer.statusClients[clientId]
	if !ok {
		// client already deregistered, or never existed
		return
	}

	updateQueryFilters(clientId, true)

	// remove the clientId from the appContext Info
	acInfo, ok := notifServer.appContexts[si.appContextID]
	if !ok {
		// this should not occur, but appcontext is clear already
		delete(notifServer.statusClients, clientId)
		delete(notifServer.streamChannels, si.stream)
		return
	}
	delete(acInfo.statusClientIDs, clientId)
	if len(acInfo.statusClientIDs) == 0 {
		// if no clients are left for the app context - unsubscribe from rsync readyNotify service
		_, err := notifServer.readyNotifyClient.Unsubscribe(ctx,
			&readynotifypb.Topic{ClientName: notifServer.name, AppContext: si.appContextID})
		if err != nil {
			log.Error("[StatusNotify gRPC] Error unsubscribing from rsync readyNotify", log.Fields{"rsync readyNotify clientId": notifServer.name, "appContextID": si.appContextID, "Error": err})
		}

		delete(notifServer.appContexts, si.appContextID)
		log.Trace("[StatusNotify gRPC] Cleaned up appContextId after last client removed", log.Fields{"clientId": clientId, "appContextID": si.appContextID})
	}
	delete(notifServer.statusClients, clientId)
	delete(notifServer.streamChannels, si.stream)

	log.Trace("[StatusNotify gRPC] Cleaned up clientId", log.Fields{"clientId": clientId, "appContextID": si.appContextID})
	return
}

// StatusDeregister will be called when the subscriber wants to terminate the stream
func (s *StatusNotifyServer) StatusDeregister(ctx context.Context, dereg *pb.StatusDeregistration) (*pb.StatusDeregistrationResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// get the clientId entry, stop the stream
	si, ok := s.statusClients[dereg.ClientId]
	if !ok {
		// client already deregistered, or never existed
		return &pb.StatusDeregistrationResponse{}, nil
	}
	s.streamChannels[si.stream] <- 1
	delete(s.statusClients, dereg.ClientId)
	delete(s.streamChannels, si.stream)

	// remove the clientId from the appContext Info
	acInfo, ok := s.appContexts[si.appContextID]
	if !ok {
		// this should not occur, but appcontext is clear already
		return &pb.StatusDeregistrationResponse{}, nil
	}
	delete(acInfo.statusClientIDs, dereg.ClientId)
	if len(acInfo.statusClientIDs) == 0 {
		// if no clients are left for the app context
		// TODO - do same as in cleanup above - consolidate code
		acInfo.readyNotifyStream.CloseSend()
		delete(s.appContexts, si.appContextID)
	}

	return &pb.StatusDeregistrationResponse{}, nil
}

// NewStatusNotifyServer will create a new StatusNotifyServer and destroys the previous one
func NewStatusNotifyServer(readyNotifyClientID string, sh StatusNotifyServerHelpers) *StatusNotifyServer {

	s := &StatusNotifyServer{
		name:              readyNotifyClientID,
		appContexts:       make(map[string]appContextInfo),
		statusClients:     make(map[string]streamInfo),
		streamChannels:    make(map[pb.StatusNotify_StatusRegisterServer]chan int),
		sh:                sh,
		readyNotifyClient: nil, // initialize this later on Registration call
	}
	notifServer = s
	return s
}

const rsyncName = "rsync"

func queryDBAndInitRsync(ctx context.Context) error {
	client := controller.NewControllerClient("resources", "data", "orchestrator")
	vals, _ := client.GetControllers(ctx)
	for _, v := range vals {
		if v.Metadata.Name == rsyncName {
			log.Info("Initializing RPC connection to resource synchronizer", log.Fields{
				"Controller": v.Metadata.Name,
			})
			inc.NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			inc.InitRsyncClient()
			return nil
		}
	}
	return pkgerrors.Errorf("queryRsyncInfoInMCODB Failed - Could not get find rsync by name : %v", rsyncName)
}

func newReadyNotifyClient(ctx context.Context) readynotifypb.ReadyNotifyClient {
	conn := rpc.GetRpcConn(ctx, rsyncName)
	if conn == nil {
		if !inc.InitRsyncClient() {
			err := queryDBAndInitRsync(ctx)
			if err != nil {
				log.Warn("[StatusNotify gRPC] Failed to initialize get ReadyNotifyClient", log.Fields{})
				return nil
			}
		}
		conn = rpc.GetRpcConn(ctx, rsyncName)
	}

	if conn != nil {
		return readynotifypb.NewReadyNotifyClient(conn)
	} else {
		log.Warn("[StatusNotify gRPC] Failed to get ReadyNotifyClient", log.Fields{})
		return nil
	}
}

// GetStatusParameters retrieves the status query parameters from the StatusRegistration
func GetStatusParameters(reg *pb.StatusRegistration) (string, string, []string, []string, []string) {
	var output, statusType string

	switch reg.StatusType {
	case pb.StatusValue_READY:
		statusType = "ready"
	case pb.StatusValue_DEPLOYED:
		statusType = "deployed"
	}

	switch reg.Output {
	case pb.OutputType_SUMMARY:
		output = "summary"
	case pb.OutputType_ALL:
		output = "all"
	}

	return statusType, output, reg.Apps, reg.Clusters, reg.Resources
}
