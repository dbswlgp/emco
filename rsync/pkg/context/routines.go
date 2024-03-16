// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package context

import (
	"bytes"
	"context"
	"fmt"
	"time"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/depend"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	. "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
	contextUtils "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"





	//"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
//	"sync"
	"strings"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	"encoding/base64"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Check status of AppContext against the event to see if it is valid
func (c *Context) checkStateChange(ctx context.Context, e RsyncEvent) (StateChange, bool, error) {
	var supported bool = false
	var err error
	var dState, cState appcontext.AppContextStatus
	var event StateChange

	event, ok := StateChanges[e]
	if !ok {
		return StateChange{}, false, pkgerrors.Errorf("Invalid Event %s:", e)
	}
	// Check Stop flag return error no processing desired
	sFlag, err := c.acRef.GetAppContextFlag(ctx, StopFlagKey)
	if err != nil {
		return event, false, pkgerrors.Errorf("AppContext Error: %s:", err)
	}
	if sFlag {
		return event, false, pkgerrors.Errorf("Stop flag set for context: %s", c.acID)
	}
	// Check PendingTerminate Flag
	tFlag, err := c.acRef.GetAppContextFlag(ctx, PendingTerminateFlagKey)
	if err != nil {
		return event, false, pkgerrors.Errorf("AppContext Error: %s:", err)
	}

	if tFlag && e != TerminateEvent {
		return event, false, pkgerrors.Errorf("Terminate Flag is set, Ignoring event: %s:", e)
	}
	// Update the desired state of the AppContext based on this event
	state, err := c.acRef.GetAppContextStatus(ctx, CurrentStateKey)
	if err != nil {
		return event, false, err
	}
	for _, s := range event.SState {
		if s == state.Status {
			supported = true
			break
		}
	}
	if !supported {
		//Exception to state machine, if event is terminate and the current state is already
		// terminated, don't change the status to TerminateFailed
		if e == TerminateEvent && state.Status == appcontext.AppContextStatusEnum.Terminated {
			return event, false, pkgerrors.Errorf("Invalid Source state %s for the Event %s:", state, e)
		}
		return event, true, pkgerrors.Errorf("Invalid Source state %s for the Event %s:", state, e)
	} else {
		dState.Status = event.DState
		cState.Status = event.CState
	}
	// Event is supported. Update Desired state and current state
	err = c.acRef.UpdateAppContextStatus(ctx, DesiredStateKey, dState)
	if err != nil {
		return event, false, err
	}
	err = c.acRef.UpdateAppContextStatus(ctx, CurrentStateKey, cState)
	if err != nil {
		return event, false, err
	}
	err = c.acRef.UpdateAppContextStatus(ctx, StatusKey, cState)
	if err != nil {
		return event, false, err
	}
	return event, true, nil
}

// UpdateQStatus updates status of an element in the queue
func (c *Context) UpdateQStatus(ctx context.Context, index int, status string) error {
	qUtils := &AppContextQueueUtils{ac: c.acRef.GetAppContextHandle()}
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if err := qUtils.UpdateStatus(ctx, index, status); err != nil {
		return err
	}
	return nil
}

// If terminate event recieved set flag to ignore other events
func (c *Context) terminateContextRoutine(ctx context.Context) {
	// Set Terminate Flag to Pending
	if err := c.acRef.UpdateAppContextFlag(ctx, PendingTerminateFlagKey, true); err != nil {
		return
	}
	// Make all waiting goroutines to stop waiting
	if c.cancel != nil {
		c.cancel()
	}
}

// Start Main Thread for handling
func (c *Context) startMainThread(ctx context.Context, a interface{}, ucid interface{}, con Connector) error {

	if ucid != nil  {

		//acID := fmt.Sprintf("%v", a)
		ucID := fmt.Sprintf("%v", ucid)

		ref, err := utils.NewAppContextReference(ctx, ucID)
		if err != nil {
			return err
		}
		// Read AppContext into CompositeApp structure
		c.ca, err = contextUtils.ReadAppContext(ctx, ucid)
		if err != nil {
			log.Error("Fatal! error reading appContext", log.Fields{"err": err})
			return err
		}
		c.acID = ucID
		c.con = con
		c.acRef = ref
		// Wait for 2 secs
		c.waitTime = 2
		c.maxRetry = getMaxRetries()
		// Check flags in AppContext to create if they don't exist and add default values
		_, err = c.acRef.GetAppContextStatus(ctx, CurrentStateKey)
		// If CurrentStateKey doesn't exist assuming this is the very first event for the appcontext
		if err != nil {
			as := appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Created}
			if err := c.acRef.UpdateAppContextStatus(ctx, CurrentStateKey, as); err != nil {
				return err
			}
		}
		_, err = c.acRef.GetAppContextFlag(ctx, StopFlagKey)
		// Assume doesn't exist and add
		if err != nil {
			if err := c.acRef.UpdateAppContextFlag(ctx, StopFlagKey, false); err != nil {
				return err
			}
		}
		_, err = c.acRef.GetAppContextFlag(ctx, PendingTerminateFlagKey)
		// Assume doesn't exist and add
		if err != nil {
			if err := c.acRef.UpdateAppContextFlag(ctx, PendingTerminateFlagKey, false); err != nil {
				return err
			}
		}
		_, err = c.acRef.GetAppContextStatus(ctx, StatusKey)
		// If CurrentStateKey doesn't exist assuming this is the very first event for the appcontext
		if err != nil {
			as := appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Created}
			if err := c.acRef.UpdateAppContextStatus(ctx, StatusKey, as); err != nil {
				return err
			}
		}
		// Read the statusAcID to use with status
		c.statusAcID, err = c.acRef.GetStatusAppContext(ctx, StatusAppContextIDKey)
		if err != nil {
			// Use appcontext as status appcontext also
			c.statusAcID = c.acID
			c.scRef = c.acRef
		} else {
			scRef, err := utils.NewAppContextReference(ctx, c.statusAcID)
			if err != nil {
				return err
			}
			c.scRef = scRef
		}
		// Intialize dependency management
		c.dm = depend.NewDependManager(c.acID)
		// Start Routine to handle AppContext

		go c.appContextRoutine(ctx)

		return nil

	} else {

		acID := fmt.Sprintf("%v", a)

		ref, err := utils.NewAppContextReference(ctx, acID)
		if err != nil {
			return err
		}
		// Read AppContext into CompositeApp structure
		c.ca, err = contextUtils.ReadAppContext(ctx, a)
		if err != nil {
			log.Error("Fatal! error reading appContext", log.Fields{"err": err})
			return err
		}
		c.acID = acID
		c.con = con
		c.acRef = ref
		// Wait for 2 secs
		c.waitTime = 2
		c.maxRetry = getMaxRetries()
		// Check flags in AppContext to create if they don't exist and add default values
		_, err = c.acRef.GetAppContextStatus(ctx, CurrentStateKey)
		// If CurrentStateKey doesn't exist assuming this is the very first event for the appcontext
		if err != nil {
			as := appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Created}
			if err := c.acRef.UpdateAppContextStatus(ctx, CurrentStateKey, as); err != nil {
				return err
			}
		}
		_, err = c.acRef.GetAppContextFlag(ctx, StopFlagKey)
		// Assume doesn't exist and add
		if err != nil {
			if err := c.acRef.UpdateAppContextFlag(ctx, StopFlagKey, false); err != nil {
				return err
			}
		}
		_, err = c.acRef.GetAppContextFlag(ctx, PendingTerminateFlagKey)
		// Assume doesn't exist and add
		if err != nil {
			if err := c.acRef.UpdateAppContextFlag(ctx, PendingTerminateFlagKey, false); err != nil {
				return err
			}
		}
		_, err = c.acRef.GetAppContextStatus(ctx, StatusKey)
		// If CurrentStateKey doesn't exist assuming this is the very first event for the appcontext
		if err != nil {
			as := appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Created}
			if err := c.acRef.UpdateAppContextStatus(ctx, StatusKey, as); err != nil {
				return err
			}
		}
		// Read the statusAcID to use with status
		c.statusAcID, err = c.acRef.GetStatusAppContext(ctx, StatusAppContextIDKey)
		if err != nil {
			// Use appcontext as status appcontext also
			c.statusAcID = c.acID
			c.scRef = c.acRef
		} else {
			scRef, err := utils.NewAppContextReference(ctx, c.statusAcID)
			if err != nil {
				return err
			}
			c.scRef = scRef
		}
		// Intialize dependency management
		c.dm = depend.NewDependManager(c.acID)
		// Start Routine to handle AppContext
		go c.appContextRoutine(ctx)
		return nil
	}
}

// Handle AppContext
func (c *Context) appContextRoutine(callerCtx context.Context) {
	var lctx context.Context
	var l context.Context
	var lGroup *errgroup.Group
	var lDone context.CancelFunc
	var op RsyncOperation

	// This function is executed asynchronously, so we must create
	// a new (not derived) context to prevent the context from
	// being cancelled when the caller completes: a cancelled
	// context will cause the below work to exit early.
	ctx := context.Background()

	// A link is used so that the traces can be associated.
	tracer := otel.Tracer("rsync")
	ctx, span := tracer.Start(ctx, "appContextRoutine",
		trace.WithLinks(trace.LinkFromContext(callerCtx)),
	)
	defer span.End()

	qUtils := &AppContextQueueUtils{ac: c.acRef.GetAppContextHandle()}
	// Create context for the running threads
	ctx, done := context.WithCancel(ctx)
	gGroup, gctx := errgroup.WithContext(ctx)
	// Stop all running goroutines
	defer done()
	// Start thread to watch for external stop flag
	gGroup.Go(func() error {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				flag, err := c.acRef.GetAppContextFlag(ctx, StopFlagKey)
				if err != nil {
					done()
				} else if flag == true {
					log.Info("Forced stop context", log.Fields{})
					// Forced Stop from outside
					done()
				}
			case <-gctx.Done():
				log.Info("Context done", log.Fields{})
				return gctx.Err()
			}
		}
	})
	// Go over all messages
	for {
		// Get first event to process
		c.Lock.Lock()
		index, ele := qUtils.FindFirstPending(ctx)


		fmt.Println("\n\n")
		fmt.Println("index: ",index)
		fmt.Println("ele: ",ele)
		fmt.Println("\n\n")



		if index >= 0 {
			c.Lock.Unlock()
			e := ele.Event
			state, skip, err := c.checkStateChange(ctx, e)
			// Event is not valid event for the current state of AppContext
			if err != nil {
				log.Error("State Change Error", log.Fields{"error": err})
				if err := c.UpdateQStatus(ctx, index, "Skip"); err != nil {
					break
				}
				if !skip {
					// Update status with error
					err = c.acRef.UpdateAppContextStatus(ctx, StatusKey, appcontext.AppContextStatus{Status: state.ErrState})
					if err != nil {
						break
					}
					// Update Current status with error
					err = c.acRef.UpdateAppContextStatus(ctx, CurrentStateKey, appcontext.AppContextStatus{Status: state.ErrState})
					if err != nil {
						break
					}
				}
				// Continue to process more events
				continue
			}
			// Create a derived context
			l, lDone = context.WithCancel(ctx)
			lGroup, lctx = errgroup.WithContext(l)
			c.Lock.Lock()
			c.cancel = lDone
			c.Lock.Unlock()
			switch e {
			case InstantiateEvent:
				op = OpApply
			case TerminateEvent:
				op = OpDelete
			case ReadEvent:
				op = OpRead
			case UpdateEvent:
				// In Instantiate Phase find out resources that need to be modified and
				// set skip to be true for those that match
				// This is done to avoid applying resources that have no differences
				if err := c.updateModifyPhase(ctx, ele); err != nil {
				//if err := c.updateModifyPhase(ctx, ele); err != nil {
					break
				}
				op = OpApply
				// Enqueue Delete Event for the AppContext that is being updated to
				//go HandleAppContext(ctx, ele.UCID, c.acID, UpdateDeleteEvent, c.con)
			case UpdateDeleteEvent:
				// Update AppContext to decide what needs deleted
/*
				fmt.Println("\n\n")
				fmt.Println("case UpdateDeleteEvent")
				fmt.Println("ele: ", ele)
				fmt.Println("\n\n")
*/
				if err := c.updateDeletePhase(ctx, ele); err != nil {
					break
				}
				op = OpDelete
			case AddChildContextEvent:
				log.Error("Not Implemented", log.Fields{"event": e})
				if err := c.UpdateQStatus(ctx, index, "Skip"); err != nil {
					break
				}
				continue
			default:
				log.Error("Unknown event", log.Fields{"event": e})
				if err := c.UpdateQStatus(ctx, index, "Skip"); err != nil {
					break
				}
				continue
			}
			lGroup.Go(func() error {
				return c.run(lctx, lGroup, op, e)
			})

			switch e {
			case UpdateEvent:
/*
				for i := 0; i < 30; i++ {

					resourcesready = "/context/" + c.statusAcID + "/app/voting-app/cluster/OpenStack+cluster1/resourcesready/"
					var value3 interface{}
					err = contextdb.Db.Get(ctx, resourcesready, &value3)
					fmt.Println("\n 1668224601331225158")
					fmt.Println("\n 1668224601331225158resourcesready", value3)

					// 2초 동안 대기
					time.Sleep(5 * time.Second)
				}
*/


				//var clusterList []string
				//var clusterList string

				time.Sleep(30 * time.Second)

				fmt.Println("\n\n\n")
				for _, app := range c.ca.AppOrder {
					// Iterate over all clusters
					for _, cluster := range c.ca.Apps[app].Clusters {

						fmt.Println("cluster: ",cluster)

						// If marked to skip then no processing needed
						if !cluster.Skip {
							//log.Info("Update Skipping Cluster::", log.Fields{"App": app, "cluster": cluster})
							// Reset bit and skip cluster
							//cluster.Skip = false
							//continue
							//clusterList = cluster.Name

							fmt.Println("\n\n\n")

							kubeconfig, err := GetKubeConfig(ctx, cluster.Name, "0", "default")

							config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
							if err != nil {
								fmt.Printf("Error building kubeconfig: %v\n", err)
							}


							clientset, err := kubernetes.NewForConfig(config)
							if err != nil {
								fmt.Printf("Error building client: %v\n", err)
							}


							// 특정 label을 갖는 Pod Informer 생성
							factory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
								options.LabelSelector = "emco/deployment-id=" + c.statusAcID + "-" + app // 특정 label 지정
							}))
							informer := factory.Core().V1().Pods().Informer()

							// Informer 시작
							go factory.Start(ctx.Done())

							// Informer가 동기화될 때까지 대기
							if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
								fmt.Println("Failed to sync")
								return
							}

							// context 생성
							ctx, cancel := context.WithCancel(context.Background())
							defer cancel()

							var allPodsRunning bool
							ticker := time.NewTicker(5 * time.Second)
							timeout := time.After(2 * time.Minute)

							go func() {
								<-timeout
								cancel()
							}()

							for {
								select {
								case <-ticker.C:
									runningPods := getRunningPods(informer.GetStore())
									fmt.Printf("Number of Running Pods of app(%s): %d\n", app, len(runningPods))

									// 모든 Pod이 Running 상태인지 확인
									if len(runningPods) == len(informer.GetStore().List()) {
										// 다음 단계를 수행하는 로직 추가
										fmt.Println("All Pods are in Running state. Proceeding to the next step...")
										allPodsRunning = true
									}
								case <-ctx.Done():
									// context가 종료되면 종료
									return
								}

								if allPodsRunning {
									break
								}
							}

						}
						fmt.Println("\n")
					}

				}

				HandleAppContext(ctx, ele.UCID, c.acID, UpdateDeleteEvent, c.con)
/*
			case UpdateDeleteEvent:
				fmt.Println("\n\n")
				fmt.Println("c: ",c)
				for _, app := range c.ca.AppOrder {
					// Iterate over all clusters
					for _, cluster := range c.ca.Apps[app].Clusters {

						fmt.Println("app: ",app)
						fmt.Println("cluster: ",cluster)
						fmt.Println("cluster.Skip: ",cluster.Skip)


						var restodelete bool

						for _, res := range cluster.Resources {
							fmt.Println("res: ",res)
							fmt.Println("res.Skip: ",res.Skip)

							if res.Skip == false {
								restodelete = true
								break
							}
						}


						if restodelete {

							kubeconfig, err := GetKubeConfig(ctx, cluster.Name, "0", "default")

							config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
							if err != nil {
								fmt.Printf("Error building kubeconfig: %v\n", err)
							}


							clientset, err := kubernetes.NewForConfig(config)
							if err != nil {
								fmt.Printf("Error building client: %v\n", err)
							}


							// 특정 label을 갖는 Pod Informer 생성
							factory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
								options.LabelSelector = "emco/deployment-id=" + c.statusAcID + "-" + app // 특정 label 지정
							}))
							informer := factory.Core().V1().Pods().Informer()

							// Informer 시작
							go factory.Start(ctx.Done())

							// Informer가 동기화될 때까지 대기
							if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
								fmt.Println("Failed to sync")
								return
							}

							// context 생성
							ctx, cancel := context.WithCancel(context.Background())
							defer cancel()

							var deployedPods int
							ticker := time.NewTicker(5 * time.Second)
							timeout := time.After(3 * time.Minute)

							go func() {
								<-timeout
								cancel()
							}()

							for {
								select {
								case <-ticker.C:
									deployedPods = len(informer.GetStore().List())
									fmt.Printf("Number of deployedPods of app(%s): %d\n", app, deployedPods)

								case <-ctx.Done():
									// context가 종료되면 종료
									return
								}

								if deployedPods == 0 {

									break
								}
							}
						}
					}
				}
				fmt.Println("\n\n")
*/



/*
				//var clusterList []string
				var clusterList string
				fmt.Println("\n\n\n")
				for _, app := range c.ca.AppOrder {
					// Iterate over all clusters
					for _, cluster := range c.ca.Apps[app].Clusters {
						// If marked to skip then no processing needed
						if !cluster.Skip {
							//log.Info("Update Skipping Cluster::", log.Fields{"App": app, "cluster": cluster})
							// Reset bit and skip cluster
							//cluster.Skip = false
							//continue
							clusterList = cluster.Name
						}
						//cluster := cluster.Name
						fmt.Println("cluster: ",cluster)
						fmt.Println("\n")
					}

				}
				fmt.Println("\n\n\n")

				kubeconfig, err := GetKubeConfig(ctx, clusterList, "0", "default")

				config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
				if err != nil {
					fmt.Printf("Error building kubeconfig: %v\n", err)
				}


				clientset, err := kubernetes.NewForConfig(config)
				if err != nil {
					fmt.Printf("Error building client: %v\n", err)
				}


				time.Sleep(30 * time.Second)

				// 특정 label을 갖는 Pod Informer 생성
				factory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
					options.LabelSelector = "emco/deployment-id=" + c.statusAcID + "-" + "voting-app" // 특정 label 지정
				}))
				informer := factory.Core().V1().Pods().Informer()

				// Informer 시작
				go factory.Start(ctx.Done())

				// Informer가 동기화될 때까지 대기
				if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
					fmt.Println("Failed to sync")
					return
				}

				// context 생성
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				var allPodsRunning bool
				ticker := time.NewTicker(5 * time.Second)
				timeout := time.After(3 * time.Minute)

				go func() {
					<-timeout
					cancel()
				}()

				for {
					select {
					case <-ticker.C:
						runningPods := getRunningPods(informer.GetStore())
						fmt.Printf("Number of Running Pods: %d\n", len(runningPods))

						// 모든 Pod이 Running 상태인지 확인
						if len(runningPods) == len(informer.GetStore().List()) {
							// 다음 단계를 수행하는 로직 추가
							fmt.Println("All Pods are in Running state. Proceeding to the next step...")
							allPodsRunning = true
							HandleAppContext(ctx, ele.UCID, c.acID, UpdateDeleteEvent, c.con)
						}
					case <-ctx.Done():
						// context가 종료되면 종료
						return
					}

					if allPodsRunning {
						break
					}
				}

*/


















/*
				var resourcesready string
				resourcesready = "/context/" + c.statusAcID + "/app/voting-app/cluster/" + clusterList + "/resourcesready/"
				var value3 bool

				time.Sleep(40 * time.Second)

				var wg sync.WaitGroup

				// 대기 그룹에 추가
				wg.Add(1)

				// 조건이 true가 될 때까지 대기
				go func() {
					defer wg.Done()

					for !value3 {
						err = contextdb.Db.Get(ctx, resourcesready, &value3)
						if value3{
							fmt.Println("\n\nresourcesready set true\n\n")
						}
						// 대기 시간 조절 (조정 가능)
						time.Sleep(5 * time.Second)
					}
				}()

				time.AfterFunc(80*time.Second, func() {
					value3 = true
				})

				// 대기 그룹 대기
				wg.Wait()

				HandleAppContext(ctx, ele.UCID, c.acID, UpdateDeleteEvent, c.con)
*/


/*
				// true 상태가 10초 동안 유지되는지 확인
				select {
				case <-time.After(10 * time.Second):
					// 10초 동안 true 상태가 유지됨
					fmt.Println("The condition remained true for 10 seconds. Proceeding to the next step...")
					go HandleAppContext(ctx, ele.UCID, c.acID, UpdateDeleteEvent, c.con)
					// 여기에 다음 단계를 수행하는 로직 추가
				default:
					// 10초 동안 true 상태가 유지되지 않음
					fmt.Println("The condition did not remain true for 10 seconds. Aborting...")
				}

*/



				/*
				for _, app := range c.ca.AppOrder {
					//label := c.acID + "-" + app
					label := "2177999313949122407" + "-" + app
					namespace, _ := c.acRef.GetNamespace(ctx)
					b, _ := status.GetStatusCR(label, "", namespace)
					fmt.Println("c.acID: ",c.acID)
					fmt.Println("ele.UCID: ",ele.UCID)
					fmt.Println(string(b))
				}*/
/*

				time.Sleep(120 * time.Second)

				for _, app := range c.ca.AppOrder {
					label := c.acID + "-" + app
					namespace, _ := c.acRef.GetNamespace(ctx)
					b, _ := status.GetStatusCR(label, "", namespace)
					fmt.Println("c.acID: ",c.acID)
					fmt.Println("ele.UCID: ",ele.UCID)
					fmt.Println(string(b))
				}

*/
				//go HandleAppContext(ctx, ele.UCID, c.acID, UpdateDeleteEvent, c.con)
			//	HandleAppContext(ctx, c.acID, ele.UCID, UpdateDeleteEvent, c.con)
			}

			// Wait for all subtasks to complete
			log.Info("Wait for all subtasks to complete", log.Fields{})
			if err := lGroup.Wait(); err != nil {
				log.Error("Failed run", log.Fields{"error": err})
				// Mark the event in Queue
				if err := c.UpdateQStatus(ctx, index, "Error"); err != nil {
					break
				}
				// Update failed status
				err = c.acRef.UpdateAppContextStatus(ctx, StatusKey, appcontext.AppContextStatus{Status: state.ErrState})
				if err != nil {
					break
				}
				// Update Current status with error
				err = c.acRef.UpdateAppContextStatus(ctx, CurrentStateKey, appcontext.AppContextStatus{Status: state.ErrState})
				if err != nil {
					break
				}
				continue
			}
			log.Info("Success all subtasks completed", log.Fields{})
			fmt.Println("\n\n")
			fmt.Println("appContextRoutine func() e: ",e)
			fmt.Println("\n\n")
			// Mark the event in Queue
			if err := c.UpdateQStatus(ctx, index, "Done"); err != nil {
				break
			}

			// 여기서 실제 배포된 리소스의 상태를 체크한 후에 CurrentState를 업데이트 해야 함. 배포된 후에 삭제되도록 하기 위해

/*
			select {
			case <-time.After(120 * time.Second):
				// Continue with the rest of the cleanup or exit logic
			case <-ctx.Done():
				// The context was canceled before the 120 seconds elapsed
				log.Info("Context canceled before waiting for 120 seconds", log.Fields{"context": c.acID})
			}
*/


			ds, _ := c.acRef.GetAppContextStatus(ctx, DesiredStateKey)

			fmt.Println("\n\n\n\n")
			//fmt.Println("acRef.acID: ",c.acID)
			//fmt.Println("c.ca: ",c.ca.AppOrder)



			//fmt.Println("ds: ",ds)
			//fmt.Println("StatusKey: ",StatusKey)
			//fmt.Println("CurrentStateKey: ",CurrentStateKey)
			fmt.Println("\n\n\n\n")

			// Success - Update Status for the AppContext to match the Desired State
			err = c.acRef.UpdateAppContextStatus(ctx, StatusKey, ds)
			err = c.acRef.UpdateAppContextStatus(ctx, CurrentStateKey, ds)






		} else {
			// Done Processing all elements in queue
			log.Info("Done Processing - no new messages", log.Fields{"context": c.acID})
			// Set the TerminatePending Flag to false before exiting
			_ = c.acRef.UpdateAppContextFlag(ctx, PendingTerminateFlagKey, false)
			// release the active contextIDs
			ok, err := DeleteActiveContextRecord(ctx, c.acID)
			if !ok {
				log.Info("Deleting activeContextID failed", log.Fields{"context": c.acID, "error": err})
			}
			c.Running = false
			c.Lock.Unlock()
			return
		}
	}
	// Any error in reading/updating appContext is considered
	// fatal and all processing stopped for the AppContext
	// Set running flag to false before exiting
	c.Lock.Lock()
	// release the active contextIDs
	ok, err := DeleteActiveContextRecord(ctx, c.acID)
	if !ok {
		log.Info("Deleting activeContextID failed", log.Fields{"context": c.acID, "error": err})
	}
	c.Running = false
	c.Lock.Unlock()
}





// GetKubeConfig uses the connectivity client to get the kubeconfig based on the name
// of the clustername.
var GetKubeConfig = func(ctx context.Context, clustername string, level string, namespace string) ([]byte, error) {
        if !strings.Contains(clustername, "+") {
                return nil, pkgerrors.New("Not a valid cluster name")
        }
        strs := strings.Split(clustername, "+")
        if len(strs) != 2 {
                return nil, pkgerrors.New("Not a valid cluster name")
        }

        ccc := db.NewCloudConfigClient()

        log.Info("Querying CloudConfig", log.Fields{"strs": strs, "level": level, "namespace": namespace})
        cconfig, err := ccc.GetCloudConfig(ctx, strs[0], strs[1], level, namespace)
        if err != nil {
                return nil, pkgerrors.Wrap(err, "Get kubeconfig failed")
        }
        log.Info("Successfully looked up CloudConfig", log.Fields{".Provider": cconfig.Provider, ".Cluster": cconfig.Cluster, ".Level": cconfig.Level, ".Namespace": cconfig.Namespace})

        dec, err := base64.StdEncoding.DecodeString(cconfig.Config)
        if err != nil {
                return nil, err
        }
        return dec, nil
}


// getRunningPods returns a list of Pods that are in the Running state
func getRunningPods(store cache.Store) []*v1.Pod {
	var runningPods []*v1.Pod

	for _, obj := range store.List() {
		pod, ok := obj.(*v1.Pod)
		if !ok {
			continue
		}

		// Check if the Pod is in the Running state
		if pod.Status.Phase == v1.PodRunning {
			runningPods = append(runningPods, pod)
		}
	}

	return runningPods
}








// Iterate over the appcontext to mark apps/cluster/resources that doesn't need to be deleted
func (c *Context) updateDeletePhase(ctx context.Context, e AppContextQueueElement) error {

	// Read Update AppContext into CompositeApp structure
	uca, err := contextUtils.ReadAppContext(ctx, e.UCID)
	if err != nil {
		log.Error("Fatal! error reading appContext", log.Fields{"err": err})
		return err
	}
	// Iterate over all the subapps and mark all apps, clusters and resources
	// that shouldn't be deleted
	for _, app := range c.ca.Apps {
		foundApp := contextUtils.FindApp(uca, app.Name)
		// If app not found that will be deleted (skip false)
		if foundApp {
			// Check if any clusters are deleted
			for _, cluster := range app.Clusters {
				foundCluster := contextUtils.FindCluster(uca, app.Name, cluster.Name)
				if foundCluster {
					// Check if any resources are deleted
					var resCnt int = 0
					for _, res := range cluster.Resources {
						foundRes := contextUtils.FindResource(uca, app.Name, cluster.Name, res.Name)
						if foundRes {
							// If resource found in both appContext don't delete it
							res.Skip = true
						} else {
							// Resource found to be deleted
							resCnt++
						}
					}
					// No resources marked for deletion, mark this cluster for not deleting
					if resCnt == 0 {
						cluster.Skip = true
					}
				}
			}
		}
	}
	return nil
}

// Iterate over the appcontext to mark apps/cluster/resources that doesn't need to be Modified
func (c *Context) updateModifyPhase(ctx context.Context, e AppContextQueueElement) error {
	// Read Update from AppContext into CompositeApp structure
	uca, err := contextUtils.ReadAppContext(ctx, e.UCID)
	if err != nil {
		log.Error("Fatal! error reading appContext", log.Fields{"err": err})
		return err
	}
	//acUtils := &utils.AppContextUtils{Ac: c.ac}
	// Load update appcontext also
	uRef, err := utils.NewAppContextReference(ctx, e.UCID)
	if err != nil {
		return err
	}
	//updateUtils := &utils.AppContextUtils{Ac: uac}
	// Iterate over all the subapps and mark all apps, clusters and resources
	// that match exactly and shouldn't be changed
	for _, app := range c.ca.Apps {
		foundApp := contextUtils.FindApp(uca, app.Name)
		if foundApp {
			// Check if any clusters are modified
			for _, cluster := range app.Clusters {
				foundCluster := contextUtils.FindCluster(uca, app.Name, cluster.Name)
				if foundCluster {
					diffRes := false
					// Check if any resources are added or modified
					for _, res := range cluster.Resources {
						foundRes := contextUtils.FindResource(uca, app.Name, cluster.Name, res.Name)
						if foundRes {
							// Read the resource from both AppContext and Compare
							cRes, _, err1 := c.acRef.GetRes(ctx, res.Name, app.Name, cluster.Name)
							uRes, _, err2 := uRef.GetRes(ctx, res.Name, app.Name, cluster.Name)
							if err1 != nil || err2 != nil {
								log.Error("Fatal Error: reading resources", log.Fields{"err1": err1, "err2": err2})
								return err1
							}
							if bytes.Equal(cRes, uRes) {
								res.Skip = true
							} else {
								log.Info("Update Resource Diff found::", log.Fields{"resource": res.Name, "cluster": cluster})
								diffRes = true
							}
						} else {
							// Found a new resource that is added to the cluster
							diffRes = true
						}
					}
					// If no resources diff, skip cluster
					if !diffRes {
						cluster.Skip = true
					}
				}
			}
		}
	}
	return nil
}

// Iterate over the appcontext to apply/delete/read resources
func (c *Context) run(ctx context.Context, g *errgroup.Group, op RsyncOperation, e RsyncEvent) error {

/*
	fmt.Println("\n\n")
	fmt.Println("c: ",c)
	fmt.Println("\n\n")

	fmt.Println("\n\n")
	fmt.Println("run func() e: ", e)
	fmt.Println("\n\n")
*/

	if op == OpApply {
		// Setup dependency before starting app and cluster threads
		// Only for Apply for now
		for _, a := range c.ca.AppOrder {
			app := a
			if len(c.ca.Apps[app].Dependency) > 0 {
				c.dm.AddDependency(app, c.ca.Apps[app].Dependency)
			}
		}
	}
	// Iterate over all the subapps and start go Routines per app
	for _, a := range c.ca.AppOrder {
		app := a
		// If marked to skip then no processing needed
		if c.ca.Apps[app].Skip {
			log.Info("Update Skipping App::", log.Fields{"App": app})
			// Reset bit and skip app
			t := c.ca.Apps[app]
			t.Skip = false
			continue
		}
		g.Go(func() error {
			err := c.runApp(ctx, g, op, app, e)
			if op == OpApply {
				// Notify dependency that app got deployed
				c.dm.NotifyAppliedStatus(app)
			}
			return err
		})
	}
	return nil
}

func (c *Context) runApp(ctx context.Context, g *errgroup.Group, op RsyncOperation, app string, e RsyncEvent) error {

	if op == OpApply {
		// Check if any dependency and wait for dependencies to be met
		if err := c.dm.WaitForDependency(ctx, app); err != nil {
			return err
		}
	}
	// Iterate over all clusters
	for _, cluster := range c.ca.Apps[app].Clusters {
		// If marked to skip then no processing needed
		if cluster.Skip {
			log.Info("Update Skipping Cluster::", log.Fields{"App": app, "cluster": cluster})
			// Reset bit and skip cluster
			cluster.Skip = false
			continue
		}
		cluster := cluster.Name
		g.Go(func() error {
			return c.runCluster(ctx, op, e, app, cluster)
		})
	}
	return nil
}

func (c *Context) runCluster(ctx context.Context, op RsyncOperation, e RsyncEvent, app, cluster string) error {

	fmt.Println("\n\n")
	fmt.Println("e: ",e)
	fmt.Println("c: ",c)
	fmt.Println("c.statusAcID: ",c.statusAcID)
	fmt.Println("\n\n")

	log.Info(" runCluster::", log.Fields{"app": app, "cluster": cluster})
	namespace, level := c.acRef.GetNamespace(ctx)
	cl, err := c.con.GetClientProviders(ctx, app, cluster, level, namespace)
	if err != nil {
		log.Error("Error in creating client", log.Fields{"error": err, "cluster": cluster, "app": app})
		return err
	}
	defer cl.CleanClientProvider()
	// Start cluster watcher if there are resources to be watched
	// case like admin cloud has no resources
	if len(c.ca.Apps[app].Clusters[cluster].ResOrder) > 0 {
		err = cl.StartClusterWatcher(ctx)
		if err != nil {
			log.Error("Error starting Cluster Watcher", log.Fields{
				"error":   err,
				"cluster": cluster,
			})
			return err
		}
		log.Trace("Started Cluster Watcher", log.Fields{
			"error":   err,
			"cluster": cluster,
		})
	}
	r := resProvd{app: app, cluster: cluster, cl: cl, context: *c}
	// Timer key
	key := app + depend.SEPARATOR + cluster
	switch e {
	case InstantiateEvent:
		// Apply config for the cluster if there are any resources to be applied
		if len(c.ca.Apps[app].Clusters[cluster].ResOrder) > 0 {
			err = cl.ApplyConfig(ctx, nil)
			if err != nil {
				return err
			}
		}
		// Check if delete of status tracker is scheduled, if so stop and delete the timer
		c.StopDeleteStatusCRTimer(key)
		// Based on the discussions in Helm handling of CRD's
		// https://helm.sh/docs/chart_best_practices/custom_resource_definitions/
		if len(c.ca.Apps[app].Clusters[cluster].Dependency["crd-install"]) > 0 {
			// Create CRD Resources if they don't exist
			log.Info("Creating CRD Resources if they don't exist", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["crd-install"]})
			_, err := r.handleResources(ctx, OpCreate, c.ca.Apps[app].Clusters[cluster].Dependency["crd-install"])
			if err != nil {
				return err
			}
		}
		if len(c.ca.Apps[app].Clusters[cluster].Dependency["pre-install"]) > 0 {
			log.Info("Installing preinstall hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["pre-install"]})
			// Add Status tracking
			if err := r.addStatusTracker(ctx, status.PreInstallHookLabel, namespace); err != nil {
				return err
			}
			// Install Preinstall hooks with wait
			_, err := r.handleResourcesWithWait(ctx, op, c.ca.Apps[app].Clusters[cluster].Dependency["pre-install"])
			if err != nil {
				r.deleteStatusTracker(ctx, status.PreInstallHookLabel, namespace)
				return err
			}
			// Delete Status tracking, will be added after the main resources are added
			r.deleteStatusTracker(ctx, status.PreInstallHookLabel, namespace)
			log.Info("Done Installing preinstall hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["pre-install"]})
		}
		// Install main resources without wait
		log.Info("Installing main resources", log.Fields{"App": app, "cluster": cluster, "resources": c.ca.Apps[app].Clusters[cluster].ResOrder})
		i, err := r.handleResources(ctx, op, c.ca.Apps[app].Clusters[cluster].ResOrder)
		// handle status tracking before exiting if at least one resource got handled
		if i > 0 {
			// Add Status tracking
			r.addStatusTracker(ctx, "", namespace)
		}
		if err != nil {
			log.Error("Error installing resources for app", log.Fields{"App": app, "cluster": cluster, "resources": c.ca.Apps[app].Clusters[cluster].ResOrder})
			return err
		}
		log.Info("Done Installing main resources", log.Fields{"App": app, "cluster": cluster, "resources": c.ca.Apps[app].Clusters[cluster].ResOrder})
		if len(c.ca.Apps[app].Clusters[cluster].Dependency["post-install"]) > 0 {
			log.Info("Installing Post-install Hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["post-install"]})
			// Install Postinstall hooks with wait
			_, err = r.handleResourcesWithWait(ctx, op, c.ca.Apps[app].Clusters[cluster].Dependency["post-install"])
			if err != nil {
				return err
			}
			log.Info("Done Installing Post-install Hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["post-install"]})
		}
	case TerminateEvent:
		// Apply Predelete hooks with wait
		if len(c.ca.Apps[app].Clusters[cluster].Dependency["pre-delete"]) > 0 {
			log.Info("Deleting pre-delete Hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["pre-delete"]})
			_, err = r.handleResourcesWithWait(ctx, OpApply, c.ca.Apps[app].Clusters[cluster].Dependency["pre-delete"])
			if err != nil {
				return err
			}
			log.Info("Done Deleting pre-delete Hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["pre-delete"]})
		}
		// Delete main resources without wait
		_, err = r.handleResources(ctx, op, c.ca.Apps[app].Clusters[cluster].ResOrder)
		if err != nil {
			return err
		}
		// Apply Postdelete hooks with wait
		if len(c.ca.Apps[app].Clusters[cluster].Dependency["post-delete"]) > 0 {
			log.Info("Deleting post-delete Hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["post-delete"]})
			_, err = r.handleResourcesWithWait(ctx, OpApply, c.ca.Apps[app].Clusters[cluster].Dependency["post-delete"])
			if err != nil {
				return err
			}
			log.Info("Done Deleting post-delete Hooks", log.Fields{"App": app, "cluster": cluster, "hooks": c.ca.Apps[app].Clusters[cluster].Dependency["post-delete"]})
		}
		var rl []string
		// Delete all hook resources also
		// Ignore errors - There can be errors if the hook resources are not applied
		// like rollback hooks and test hooks
		for _, d := range c.ca.Apps[app].Clusters[cluster].Dependency {
			rl = append(rl, d...)
		}
		// Ignore errors
		_, _ = r.handleResources(ctx, op, rl)

		// Delete config for the cluster if applied
		if len(c.ca.Apps[app].Clusters[cluster].ResOrder) > 0 {
			err = cl.DeleteConfig(ctx, nil)
			if err != nil {
				return err
			}
		}
		// Check if delete of status tracker is scheduled, if so stop and delete the timer
		// before scheduling a new one
		c.StopDeleteStatusCRTimer(key)
		timer := ScheduleDeleteStatusTracker(ctx, c.statusAcID, app, cluster, level, namespace, c.con)
		c.UpdateDeleteStatusCRTimer(key, timer)
	case UpdateEvent, UpdateDeleteEvent:
		// Update and Rollback hooks are not supported at this time
		var rl []string
		// Find resources to handle based on skip bit
		resOrder := c.ca.Apps[app].Clusters[cluster].ResOrder
		for _, res := range resOrder {
			// If marked to skip then no processing needed
			if !c.ca.Apps[app].Clusters[cluster].Resources[res].Skip {
				rl = append(rl, res)
			}
		}
		// Handle main resources without wait
		_, err = r.handleResources(ctx, op, rl)
		if err != nil {
			return err
		}
		// Add Status tracking if not already applied for the cluster
		if op == OpApply {
			r.addStatusTracker(ctx, "", namespace)
		}
	}
	return nil
}

// Schedule delete status tracker to run after 2 mins
// This gives time for delete status to be recorded in the monitor CR
func ScheduleDeleteStatusTracker(ctx context.Context, acID, app, cluster, level, namespace string, con Connector) *time.Timer {

	DurationOfTime := time.Duration(120) * time.Second
	label := acID + "-" + app
	b, err := status.GetStatusCR(label, "", namespace)
	if err != nil {
		log.Error("Failed to get status CR for deleting", log.Fields{"error": err, "label": label})
		return &time.Timer{}
	}

	tracer := otel.Tracer("rsync")
	ctx, span := tracer.Start(context.Background(), "DeleteStatusCR",
		trace.WithLinks(trace.LinkFromContext(ctx)),
	)

	f := func() {
		cl, err := con.GetClientProviders(ctx, app, cluster, level, namespace)
		if err != nil {
			log.Error("Error in creating client", log.Fields{"error": err, "cluster": cluster, "app": app})
			span.End()
			return
		}
		defer cl.CleanClientProvider()
		if err = cl.DeleteStatusCR(ctx, label, b); err != nil {
			log.Info("Failed to delete res", log.Fields{"error": err, "app": app, "label": label})
			span.End()
			return
		}
		span.End()
	}
	// Schedule for running at a later time
	handle := time.AfterFunc(DurationOfTime, f)
	return handle
}
