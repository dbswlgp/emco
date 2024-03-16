// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package controllers

import (
	"context"
	"fmt"
	"log"

	k8spluginv1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type deletefn func(cr *k8spluginv1alpha1.ResourceBundleState, name, namespace string, gvk schema.GroupVersionKind) (bool, error)
type updatefn func(cr *k8spluginv1alpha1.ResourceBundleState, item *unstructured.Unstructured, name, namespace string) (bool, error)

type UpdateStatusClient struct {
	del        deletefn
	update     updatefn
	c          client.Client
	item       *unstructured.Unstructured
	name       string
	namespace  string
	gvk        schema.GroupVersionKind
	defaultRes bool
}

func (u *UpdateStatusClient) Update() error {
	var err error
	var found bool
	var rbStatus *k8spluginv1alpha1.ResourceBundleState
	if u.defaultRes {
		u.del = DeleteObj
		u.update = UpdateStatus
	} else {
		u.del = DeleteResourceStatusCR
		u.update = UpdateResourceStatusCR
	}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		rbStatus, err = GetCRForResource(u.c, u.item, u.namespace)
		if err != nil {
			return err
		}
		if u.item.GetDeletionTimestamp() == nil {
			_, err = u.update(rbStatus, u.item, u.name, u.namespace)
			found = true
		} else {
			found, err = u.del(rbStatus, u.item.GetName(), u.item.GetNamespace(), u.item.GroupVersionKind())
		}
		if err != nil {
			return err
		}
		if found {
			err = u.c.Status().Update(context.TODO(), rbStatus)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update resource status: %w", err)
	}
	// Check if GIT Info is provided if so store the information in the Git Repo also

	if GitClient != (GitAccessClient{}) {
		err = GitClient.CommitCRToGit(rbStatus, rbStatus.GetLabels())
		if err != nil {
			log.Println("Error commiting status to Git", err)
		}
	}
	return nil
}

type DeleteStatusClient struct {
	del        deletefn
	c          client.Client
	gvk        schema.GroupVersionKind
	name       string
	namespace  string
	defaultRes bool
}

func (u *DeleteStatusClient) Delete() error {
	var err error
	if u.defaultRes {
		u.del = DeleteObj
	} else {
		u.del = DeleteResourceStatusCR
	}
	rbStatusList := &k8spluginv1alpha1.ResourceBundleStateList{}
	// Find all CR's tracking this item
	err = listResources(u.c, u.namespace, nil, rbStatusList)
	if err != nil || len(rbStatusList.Items) == 0 {
		log.Printf("Did not find any CRs tracking this resource\n")
		return nil
	}
	for _, cr := range rbStatusList.Items {
		// Delete from each
		err = u.DeleteOne(&cr)
	}
	return nil
}

func (u *DeleteStatusClient) DeleteOne(rbState *k8spluginv1alpha1.ResourceBundleState) error {
	var err error
	var found bool

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var namespaced types.NamespacedName
		namespaced.Name = rbState.GetName()
		namespaced.Namespace = rbState.GetNamespace()
		var cr k8spluginv1alpha1.ResourceBundleState
		err := u.c.Get(context.TODO(), namespaced, &cr)
		if err != nil {
			return err
		}
		found, err = u.del(&cr, u.name, u.namespace, u.gvk)
		if err != nil {
			return err
		}
		if found {
			err = u.c.Status().Update(context.TODO(), rbState)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update resource status: %w", err)
	}
	// Check if GIT Info is provided if so store the information in the Git Repo also

	if GitClient != (GitAccessClient{}) {
		err = GitClient.CommitCRToGit(rbState, rbState.GetLabels())
		if err != nil {
			log.Println("Error commiting status to Git", err)
		}
	}
	return nil
}
