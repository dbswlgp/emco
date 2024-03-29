// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Aarna Networks, Inc.

//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricsCollector) DeepCopyInto(out *MetricsCollector) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricsCollector.
func (in *MetricsCollector) DeepCopy() *MetricsCollector {
	if in == nil {
		return nil
	}
	out := new(MetricsCollector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MetricsCollector) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricsCollectorList) DeepCopyInto(out *MetricsCollectorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MetricsCollector, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricsCollectorList.
func (in *MetricsCollectorList) DeepCopy() *MetricsCollectorList {
	if in == nil {
		return nil
	}
	out := new(MetricsCollectorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MetricsCollectorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricsCollectorSpec) DeepCopyInto(out *MetricsCollectorSpec) {
	*out = *in
	if in.MetricList != nil {
		in, out := &in.MetricList, &out.MetricList
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricsCollectorSpec.
func (in *MetricsCollectorSpec) DeepCopy() *MetricsCollectorSpec {
	if in == nil {
		return nil
	}
	out := new(MetricsCollectorSpec)
	in.DeepCopyInto(out)
	return out
}
