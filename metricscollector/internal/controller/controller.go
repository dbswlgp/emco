// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

// Package controller EdgeMetricsCollector is a kubernetes custom controller, which watches the Custom Resource, "MetricsCollector".
// MetricsCollector CR will contain the list of metrics, that this controller will periodically fetch from custom.metrics.k8s.io.
// This controller will watch only resources that are managed by EMCO. It uses the label "emco/deployment-id" as the selector to filter the resounces
// managed by EMCO.
package controller

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	clientSet "edgemetricscollector/pkg/generated/clientset/versioned"
	metricsScheme "edgemetricscollector/pkg/generated/clientset/versioned/scheme"
	informers "edgemetricscollector/pkg/generated/informers/externalversions/edgemetricscollector/v1alpha1"
)

const (
	controllerAgentName = "edgemetricscollector"
	emcoSelector        = "emco/deployment-id"
)

// NewController initializes the EMCO EdgeMetricsCollector controller.
func NewController(
	kubeClientSet kubernetes.Interface,
	collectorClientSet clientSet.Interface,
	podInformer coreinformers.PodInformer,
	metricsCollectorInformer informers.MetricsCollectorInformer,
	metricsStream chan<- Measurements) *Controller {

	// Standard steps for initializing a k8s controller
	utilruntime.Must(metricsScheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeClientSet:      kubeClientSet,
		collectorClientSet: collectorClientSet,
		collectorLister:    metricsCollectorInformer.Lister(),
		collectorSynced:    metricsCollectorInformer.Informer().HasSynced,
		podLister:          podInformer.Lister(),
		podsSynced:         podInformer.Informer().HasSynced,
		workQueue:          workqueue.NewNamedDelayingQueue("EdgeMetricsCollectors"),
		recorder:           recorder,
		metricsStream:      metricsStream,
	}

	metricsCollectorInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.enqueue,
		UpdateFunc: controller.requeue,
		DeleteFunc: controller.dequeue,
	})

	return controller
}

// Run will set up the event handlers, as well as syncing informer caches and starting workers.
// It will block until stopCh is closed, at which point it will shutdown the workQueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()
	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting EdgeMetircsCollector controller")

	// Wait for the caches to be synced before starting workers
	// We need both pod and metricCollector cache to be sync-ed
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.podsSynced, c.collectorSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch workers
	// More than one  worker is not tested.
	// Different MetricCollector CR may allow to run multiple workers, but currently this is not tested
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workQueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workQueue and
// attempt to process it, by calling the metricsHandler.
// Once the process is done, it put back the item to queue, which creates the rate-limited loop
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workQueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			// As the item in the workQueue is actually invalid,
			// we are not adding back to the queue
			// Delete and create of CR might be required in this case
			utilruntime.HandleError(fmt.Errorf("expected string in workQueue but got %#v", obj))
			return nil
		}
		err := c.metricsHandler(key)
		// We are creating a loop here by adding back. The MetricsCollector CR is always added back to workqueue by the
		// controller. This allows metrics to be consumed in a rate limited way
		// We will put the item back on the workQueue even in case of error.
		// Basic flow is as follows:
		// Enqueue CR to WorkQueue -> Dequeue from Queue -> Process the Item (Fetch metrics listed in CR and send it EMCO Policy controller)
		//   -> Enqueue CR to WorkQueue -> .. (loop continues)
		c.workQueue.AddAfter(key, 30*time.Second)
		if err != nil {
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		klog.Infof("Successfully processed metrics for '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// metricsHandler traverse through all the pods that are managed by EMCO.
// It gets the custom metrics for these pods, for all the metrics that need to be watched as per
// MetricsCollector CR's definition.
func (c *Controller) metricsHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the MetricCollector resource
	// This should get the updated value. Hence, we may not need to handle the updates (see requeue function)
	metricObject, err := c.collectorLister.MetricsCollectors(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("MetricCollector '%s' in work queue no longer exists.Ignoring", key))

			return nil
		}
		return err
	}
	metrics := metricObject.Spec.MetricList
	emcoManaged, err := labels.Parse(emcoSelector)
	// Get list of pods managed by EMCO
	pods, err := c.podLister.List(emcoManaged)
	if err != nil {
		klog.Errorf("Error while reading pod list from cache %s", err.Error())
		return err
	}

	for _, pod := range pods {
		podName := pod.ObjectMeta.Name
		podNameSpace := pod.ObjectMeta.Namespace
		deploymentIdSplit := strings.Split(pod.Labels[emcoSelector], "-")
		contextId := deploymentIdSplit[0]
		appName := deploymentIdSplit[1]
		for _, metric := range metrics {
			value, err := c.GetCustomMetrics(podNameSpace, podName, metric)
			if err != nil {
				klog.Errorf("Error while reading custom metrics. Ignoring. Err =  %s", err.Error())
				continue
			}
			measurement := Measurements{
				Pod:       podName,
				Namespace: podNameSpace,
				ContextId: contextId,
				AppName:   appName,
				Metric:    metric,
				Value:     value,
			}
			fmt.Println("Metrics from custom.metrics.k8s.io, with enriched data:", measurement)
			// TODO: Need more clarity on side-effects of this channel getting blocked (when emco controller is not connected)
			// and the progress of the workqueue.
			c.metricsStream <- measurement
			// We should move to sending more measurement in one go.
			// This require refactoring of the code, both at the controller and the agent
			// measurements = append(measurements, measurement)
		}
	}

	//c.metricsStream <- measurements
	return nil
}

// GetCustomMetrics reads metrics from custom.metrics.k8s.io API
// HPA (Horizontal Pod AutoScaler) uses an SDK for this. We may need to move it in future
// Currently we will have minimal support by directly using k8s APIs
func (c *Controller) GetCustomMetrics(namespace string, pod string, metric string) ([]byte, error) {
	apiPrefix := "apis/custom.metrics.k8s.io/v1beta1/namespaces/" + namespace + "/pods/" + pod
	api := apiPrefix + "/" + metric
	return c.kubeClientSet.CoreV1().RESTClient().Get().AbsPath(api).DoRaw(context.Background())
}

// enqueue method for pushing item to workqueue. This gets called when CR is created.
// This is where we add CR to queue first. Later processNextWorkItem add this to queue after each
// processing to create a loop.
func (c *Controller) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workQueue.Add(key)
}

func (c *Controller) dequeue(_ interface{}) {
	// TODO: Delayed Queue does't support a delete of object like  forget int ratelimiting queue
	// We need to find some other mechanism to delete  (like a flag)
	// Currently the object will get added back even if we delete it
	// An error will appear when trying to access the CR since it is not available any more.
	// This need to be fixed
}

// TODO need to revisit
func (c *Controller) requeue(_, _ interface{}) {

	// TODO Need to confirm whether we need to handle updates
	// - If we get updated values cache always in sync handler, we can ignore the updates
	// - Else we need to enqueue new object and do necessary to avoid older object getting queued
	// Possible solution  is as follows
	//
	//	c.workQueue.Forget(old)
	//	var key string
	//	var err error
	//	if key, err = cache.MetaNamespaceKeyFunc(new); err != nil {
	//		utilruntime.HandleError(err)
	//		return
	//	}
	//	c.workQueue.Add(key)
	//
}
