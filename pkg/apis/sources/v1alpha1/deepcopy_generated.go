//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright (c) 2020-2021 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	apis "github.com/triggermesh/aws-event-sources/pkg/apis"
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCloudWatchLogsSource) DeepCopyInto(out *AWSCloudWatchLogsSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCloudWatchLogsSource.
func (in *AWSCloudWatchLogsSource) DeepCopy() *AWSCloudWatchLogsSource {
	if in == nil {
		return nil
	}
	out := new(AWSCloudWatchLogsSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSCloudWatchLogsSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCloudWatchLogsSourceList) DeepCopyInto(out *AWSCloudWatchLogsSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSCloudWatchLogsSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCloudWatchLogsSourceList.
func (in *AWSCloudWatchLogsSourceList) DeepCopy() *AWSCloudWatchLogsSourceList {
	if in == nil {
		return nil
	}
	out := new(AWSCloudWatchLogsSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSCloudWatchLogsSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCloudWatchLogsSourceSpec) DeepCopyInto(out *AWSCloudWatchLogsSourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	out.ARN = in.ARN
	if in.PollingInterval != nil {
		in, out := &in.PollingInterval, &out.PollingInterval
		*out = new(apis.Duration)
		**out = **in
	}
	in.Credentials.DeepCopyInto(&out.Credentials)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCloudWatchLogsSourceSpec.
func (in *AWSCloudWatchLogsSourceSpec) DeepCopy() *AWSCloudWatchLogsSourceSpec {
	if in == nil {
		return nil
	}
	out := new(AWSCloudWatchLogsSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCloudWatchMetric) DeepCopyInto(out *AWSCloudWatchMetric) {
	*out = *in
	if in.Dimensions != nil {
		in, out := &in.Dimensions, &out.Dimensions
		*out = make([]AWSCloudWatchMetricDimension, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCloudWatchMetric.
func (in *AWSCloudWatchMetric) DeepCopy() *AWSCloudWatchMetric {
	if in == nil {
		return nil
	}
	out := new(AWSCloudWatchMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCloudWatchMetricDimension) DeepCopyInto(out *AWSCloudWatchMetricDimension) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCloudWatchMetricDimension.
func (in *AWSCloudWatchMetricDimension) DeepCopy() *AWSCloudWatchMetricDimension {
	if in == nil {
		return nil
	}
	out := new(AWSCloudWatchMetricDimension)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCloudWatchMetricQuery) DeepCopyInto(out *AWSCloudWatchMetricQuery) {
	*out = *in
	if in.Expression != nil {
		in, out := &in.Expression, &out.Expression
		*out = new(string)
		**out = **in
	}
	if in.Metric != nil {
		in, out := &in.Metric, &out.Metric
		*out = new(AWSCloudWatchMetricStat)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCloudWatchMetricQuery.
func (in *AWSCloudWatchMetricQuery) DeepCopy() *AWSCloudWatchMetricQuery {
	if in == nil {
		return nil
	}
	out := new(AWSCloudWatchMetricQuery)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCloudWatchMetricStat) DeepCopyInto(out *AWSCloudWatchMetricStat) {
	*out = *in
	in.Metric.DeepCopyInto(&out.Metric)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCloudWatchMetricStat.
func (in *AWSCloudWatchMetricStat) DeepCopy() *AWSCloudWatchMetricStat {
	if in == nil {
		return nil
	}
	out := new(AWSCloudWatchMetricStat)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCloudWatchSource) DeepCopyInto(out *AWSCloudWatchSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCloudWatchSource.
func (in *AWSCloudWatchSource) DeepCopy() *AWSCloudWatchSource {
	if in == nil {
		return nil
	}
	out := new(AWSCloudWatchSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSCloudWatchSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCloudWatchSourceList) DeepCopyInto(out *AWSCloudWatchSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSCloudWatchSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCloudWatchSourceList.
func (in *AWSCloudWatchSourceList) DeepCopy() *AWSCloudWatchSourceList {
	if in == nil {
		return nil
	}
	out := new(AWSCloudWatchSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSCloudWatchSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCloudWatchSourceSpec) DeepCopyInto(out *AWSCloudWatchSourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	if in.PollingInterval != nil {
		in, out := &in.PollingInterval, &out.PollingInterval
		*out = new(apis.Duration)
		**out = **in
	}
	if in.MetricQueries != nil {
		in, out := &in.MetricQueries, &out.MetricQueries
		*out = make([]AWSCloudWatchMetricQuery, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Credentials.DeepCopyInto(&out.Credentials)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCloudWatchSourceSpec.
func (in *AWSCloudWatchSourceSpec) DeepCopy() *AWSCloudWatchSourceSpec {
	if in == nil {
		return nil
	}
	out := new(AWSCloudWatchSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCodeCommitSource) DeepCopyInto(out *AWSCodeCommitSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCodeCommitSource.
func (in *AWSCodeCommitSource) DeepCopy() *AWSCodeCommitSource {
	if in == nil {
		return nil
	}
	out := new(AWSCodeCommitSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSCodeCommitSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCodeCommitSourceList) DeepCopyInto(out *AWSCodeCommitSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSCodeCommitSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCodeCommitSourceList.
func (in *AWSCodeCommitSourceList) DeepCopy() *AWSCodeCommitSourceList {
	if in == nil {
		return nil
	}
	out := new(AWSCodeCommitSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSCodeCommitSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCodeCommitSourceSpec) DeepCopyInto(out *AWSCodeCommitSourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	out.ARN = in.ARN
	if in.EventTypes != nil {
		in, out := &in.EventTypes, &out.EventTypes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.Credentials.DeepCopyInto(&out.Credentials)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCodeCommitSourceSpec.
func (in *AWSCodeCommitSourceSpec) DeepCopy() *AWSCodeCommitSourceSpec {
	if in == nil {
		return nil
	}
	out := new(AWSCodeCommitSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCognitoIdentitySource) DeepCopyInto(out *AWSCognitoIdentitySource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCognitoIdentitySource.
func (in *AWSCognitoIdentitySource) DeepCopy() *AWSCognitoIdentitySource {
	if in == nil {
		return nil
	}
	out := new(AWSCognitoIdentitySource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSCognitoIdentitySource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCognitoIdentitySourceList) DeepCopyInto(out *AWSCognitoIdentitySourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSCognitoIdentitySource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCognitoIdentitySourceList.
func (in *AWSCognitoIdentitySourceList) DeepCopy() *AWSCognitoIdentitySourceList {
	if in == nil {
		return nil
	}
	out := new(AWSCognitoIdentitySourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSCognitoIdentitySourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCognitoIdentitySourceSpec) DeepCopyInto(out *AWSCognitoIdentitySourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	out.ARN = in.ARN
	in.Credentials.DeepCopyInto(&out.Credentials)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCognitoIdentitySourceSpec.
func (in *AWSCognitoIdentitySourceSpec) DeepCopy() *AWSCognitoIdentitySourceSpec {
	if in == nil {
		return nil
	}
	out := new(AWSCognitoIdentitySourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCognitoUserPoolSource) DeepCopyInto(out *AWSCognitoUserPoolSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCognitoUserPoolSource.
func (in *AWSCognitoUserPoolSource) DeepCopy() *AWSCognitoUserPoolSource {
	if in == nil {
		return nil
	}
	out := new(AWSCognitoUserPoolSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSCognitoUserPoolSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCognitoUserPoolSourceList) DeepCopyInto(out *AWSCognitoUserPoolSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSCognitoUserPoolSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCognitoUserPoolSourceList.
func (in *AWSCognitoUserPoolSourceList) DeepCopy() *AWSCognitoUserPoolSourceList {
	if in == nil {
		return nil
	}
	out := new(AWSCognitoUserPoolSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSCognitoUserPoolSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCognitoUserPoolSourceSpec) DeepCopyInto(out *AWSCognitoUserPoolSourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	out.ARN = in.ARN
	in.Credentials.DeepCopyInto(&out.Credentials)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCognitoUserPoolSourceSpec.
func (in *AWSCognitoUserPoolSourceSpec) DeepCopy() *AWSCognitoUserPoolSourceSpec {
	if in == nil {
		return nil
	}
	out := new(AWSCognitoUserPoolSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSDynamoDBSource) DeepCopyInto(out *AWSDynamoDBSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSDynamoDBSource.
func (in *AWSDynamoDBSource) DeepCopy() *AWSDynamoDBSource {
	if in == nil {
		return nil
	}
	out := new(AWSDynamoDBSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSDynamoDBSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSDynamoDBSourceList) DeepCopyInto(out *AWSDynamoDBSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSDynamoDBSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSDynamoDBSourceList.
func (in *AWSDynamoDBSourceList) DeepCopy() *AWSDynamoDBSourceList {
	if in == nil {
		return nil
	}
	out := new(AWSDynamoDBSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSDynamoDBSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSDynamoDBSourceSpec) DeepCopyInto(out *AWSDynamoDBSourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	out.ARN = in.ARN
	in.Credentials.DeepCopyInto(&out.Credentials)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSDynamoDBSourceSpec.
func (in *AWSDynamoDBSourceSpec) DeepCopy() *AWSDynamoDBSourceSpec {
	if in == nil {
		return nil
	}
	out := new(AWSDynamoDBSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSKinesisSource) DeepCopyInto(out *AWSKinesisSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSKinesisSource.
func (in *AWSKinesisSource) DeepCopy() *AWSKinesisSource {
	if in == nil {
		return nil
	}
	out := new(AWSKinesisSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSKinesisSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSKinesisSourceList) DeepCopyInto(out *AWSKinesisSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSKinesisSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSKinesisSourceList.
func (in *AWSKinesisSourceList) DeepCopy() *AWSKinesisSourceList {
	if in == nil {
		return nil
	}
	out := new(AWSKinesisSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSKinesisSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSKinesisSourceSpec) DeepCopyInto(out *AWSKinesisSourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	out.ARN = in.ARN
	in.Credentials.DeepCopyInto(&out.Credentials)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSKinesisSourceSpec.
func (in *AWSKinesisSourceSpec) DeepCopy() *AWSKinesisSourceSpec {
	if in == nil {
		return nil
	}
	out := new(AWSKinesisSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSPerformanceInsightsSource) DeepCopyInto(out *AWSPerformanceInsightsSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSPerformanceInsightsSource.
func (in *AWSPerformanceInsightsSource) DeepCopy() *AWSPerformanceInsightsSource {
	if in == nil {
		return nil
	}
	out := new(AWSPerformanceInsightsSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSPerformanceInsightsSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSPerformanceInsightsSourceList) DeepCopyInto(out *AWSPerformanceInsightsSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSPerformanceInsightsSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSPerformanceInsightsSourceList.
func (in *AWSPerformanceInsightsSourceList) DeepCopy() *AWSPerformanceInsightsSourceList {
	if in == nil {
		return nil
	}
	out := new(AWSPerformanceInsightsSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSPerformanceInsightsSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSPerformanceInsightsSourceSpec) DeepCopyInto(out *AWSPerformanceInsightsSourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	out.ARN = in.ARN
	if in.Metrics != nil {
		in, out := &in.Metrics, &out.Metrics
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.Credentials.DeepCopyInto(&out.Credentials)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSPerformanceInsightsSourceSpec.
func (in *AWSPerformanceInsightsSourceSpec) DeepCopy() *AWSPerformanceInsightsSourceSpec {
	if in == nil {
		return nil
	}
	out := new(AWSPerformanceInsightsSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSS3Source) DeepCopyInto(out *AWSS3Source) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSS3Source.
func (in *AWSS3Source) DeepCopy() *AWSS3Source {
	if in == nil {
		return nil
	}
	out := new(AWSS3Source)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSS3Source) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSS3SourceList) DeepCopyInto(out *AWSS3SourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSS3Source, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSS3SourceList.
func (in *AWSS3SourceList) DeepCopy() *AWSS3SourceList {
	if in == nil {
		return nil
	}
	out := new(AWSS3SourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSS3SourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSS3SourceSpec) DeepCopyInto(out *AWSS3SourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	out.ARN = in.ARN
	if in.EventTypes != nil {
		in, out := &in.EventTypes, &out.EventTypes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.QueueARN != nil {
		in, out := &in.QueueARN, &out.QueueARN
		*out = new(apis.ARN)
		**out = **in
	}
	in.Credentials.DeepCopyInto(&out.Credentials)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSS3SourceSpec.
func (in *AWSS3SourceSpec) DeepCopy() *AWSS3SourceSpec {
	if in == nil {
		return nil
	}
	out := new(AWSS3SourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSS3SourceStatus) DeepCopyInto(out *AWSS3SourceStatus) {
	*out = *in
	in.EventSourceStatus.DeepCopyInto(&out.EventSourceStatus)
	if in.QueueARN != nil {
		in, out := &in.QueueARN, &out.QueueARN
		*out = new(apis.ARN)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSS3SourceStatus.
func (in *AWSS3SourceStatus) DeepCopy() *AWSS3SourceStatus {
	if in == nil {
		return nil
	}
	out := new(AWSS3SourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSNSSource) DeepCopyInto(out *AWSSNSSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSNSSource.
func (in *AWSSNSSource) DeepCopy() *AWSSNSSource {
	if in == nil {
		return nil
	}
	out := new(AWSSNSSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSSNSSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSNSSourceList) DeepCopyInto(out *AWSSNSSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSSNSSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSNSSourceList.
func (in *AWSSNSSourceList) DeepCopy() *AWSSNSSourceList {
	if in == nil {
		return nil
	}
	out := new(AWSSNSSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSSNSSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSNSSourceSpec) DeepCopyInto(out *AWSSNSSourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	out.ARN = in.ARN
	if in.SubscriptionAttributes != nil {
		in, out := &in.SubscriptionAttributes, &out.SubscriptionAttributes
		*out = make(map[string]*string, len(*in))
		for key, val := range *in {
			var outVal *string
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(string)
				**out = **in
			}
			(*out)[key] = outVal
		}
	}
	in.Credentials.DeepCopyInto(&out.Credentials)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSNSSourceSpec.
func (in *AWSSNSSourceSpec) DeepCopy() *AWSSNSSourceSpec {
	if in == nil {
		return nil
	}
	out := new(AWSSNSSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSNSSourceStatus) DeepCopyInto(out *AWSSNSSourceStatus) {
	*out = *in
	in.EventSourceStatus.DeepCopyInto(&out.EventSourceStatus)
	if in.SubscriptionARN != nil {
		in, out := &in.SubscriptionARN, &out.SubscriptionARN
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSNSSourceStatus.
func (in *AWSSNSSourceStatus) DeepCopy() *AWSSNSSourceStatus {
	if in == nil {
		return nil
	}
	out := new(AWSSNSSourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSQSSource) DeepCopyInto(out *AWSSQSSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSQSSource.
func (in *AWSSQSSource) DeepCopy() *AWSSQSSource {
	if in == nil {
		return nil
	}
	out := new(AWSSQSSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSSQSSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSQSSourceList) DeepCopyInto(out *AWSSQSSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSSQSSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSQSSourceList.
func (in *AWSSQSSourceList) DeepCopy() *AWSSQSSourceList {
	if in == nil {
		return nil
	}
	out := new(AWSSQSSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSSQSSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSQSSourceReceiveOptions) DeepCopyInto(out *AWSSQSSourceReceiveOptions) {
	*out = *in
	if in.VisibilityTimeout != nil {
		in, out := &in.VisibilityTimeout, &out.VisibilityTimeout
		*out = new(apis.Duration)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSQSSourceReceiveOptions.
func (in *AWSSQSSourceReceiveOptions) DeepCopy() *AWSSQSSourceReceiveOptions {
	if in == nil {
		return nil
	}
	out := new(AWSSQSSourceReceiveOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSQSSourceSpec) DeepCopyInto(out *AWSSQSSourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	out.ARN = in.ARN
	if in.ReceiveOptions != nil {
		in, out := &in.ReceiveOptions, &out.ReceiveOptions
		*out = new(AWSSQSSourceReceiveOptions)
		(*in).DeepCopyInto(*out)
	}
	in.Credentials.DeepCopyInto(&out.Credentials)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSQSSourceSpec.
func (in *AWSSQSSourceSpec) DeepCopy() *AWSSQSSourceSpec {
	if in == nil {
		return nil
	}
	out := new(AWSSQSSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSSecurityCredentials) DeepCopyInto(out *AWSSecurityCredentials) {
	*out = *in
	in.AccessKeyID.DeepCopyInto(&out.AccessKeyID)
	in.SecretAccessKey.DeepCopyInto(&out.SecretAccessKey)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSSecurityCredentials.
func (in *AWSSecurityCredentials) DeepCopy() *AWSSecurityCredentials {
	if in == nil {
		return nil
	}
	out := new(AWSSecurityCredentials)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventSourceStatus) DeepCopyInto(out *EventSourceStatus) {
	*out = *in
	in.SourceStatus.DeepCopyInto(&out.SourceStatus)
	in.AddressStatus.DeepCopyInto(&out.AddressStatus)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventSourceStatus.
func (in *EventSourceStatus) DeepCopy() *EventSourceStatus {
	if in == nil {
		return nil
	}
	out := new(EventSourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValueFromField) DeepCopyInto(out *ValueFromField) {
	*out = *in
	if in.ValueFromSecret != nil {
		in, out := &in.ValueFromSecret, &out.ValueFromSecret
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ValueFromField.
func (in *ValueFromField) DeepCopy() *ValueFromField {
	if in == nil {
		return nil
	}
	out := new(ValueFromField)
	in.DeepCopyInto(out)
	return out
}
