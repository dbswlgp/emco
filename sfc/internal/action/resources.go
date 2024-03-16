// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package action

import (
	"fmt"
	"strings"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	v1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1core "k8s.io/api/core/v1"
)

func updatePodTemplateLabels(pt *v1core.PodTemplateSpec, label string) {
	kv := strings.Split(label, "=")
	if len(kv) != 2 {
		log.Warn("SFC link label has invalid format", log.Fields{"label": label})
		return
	}
	pt.Labels[kv[0]] = kv[1]
}

// AddLabelsToPodTemplates adds the labels in matchLabels to the labels in the pod template
// of the resource r.
func addLabelToPodTemplates(r interface{}, label string) {

	switch o := r.(type) {
	case *batch.Job:
		updatePodTemplateLabels(&o.Spec.Template, label)
	case *batchv1beta1.CronJob:
		updatePodTemplateLabels(&o.Spec.JobTemplate.Spec.Template, label)
	case *v1.DaemonSet:
		updatePodTemplateLabels(&o.Spec.Template, label)
		return
	case *v1.Deployment:
		updatePodTemplateLabels(&o.Spec.Template, label)
		return
	case *v1.ReplicaSet:
		updatePodTemplateLabels(&o.Spec.Template, label)
	case *v1.StatefulSet:
		updatePodTemplateLabels(&o.Spec.Template, label)
	case *v1core.Pod:
		kv := strings.Split(label, "=")
		if len(kv) != 2 {
			log.Warn("SFC link label has invalid format", log.Fields{"label": label})
			return
		}
		o.Labels[kv[0]] = kv[1]
		return
	case *v1core.ReplicationController:
		updatePodTemplateLabels(o.Spec.Template, label)
		return
	default:
		typeStr := fmt.Sprintf("%T", o)
		log.Warn("Resource type does not have pod template", log.Fields{
			"resource type": typeStr,
		})
	}
}
