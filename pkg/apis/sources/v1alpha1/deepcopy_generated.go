// +build !ignore_autogenerated

/*
Copyright (c) 2020 TriggerMesh Inc.

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
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

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
func (in *AWSCodeCommitSourceStatus) DeepCopyInto(out *AWSCodeCommitSourceStatus) {
	*out = *in
	in.SourceStatus.DeepCopyInto(&out.SourceStatus)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCodeCommitSourceStatus.
func (in *AWSCodeCommitSourceStatus) DeepCopy() *AWSCodeCommitSourceStatus {
	if in == nil {
		return nil
	}
	out := new(AWSCodeCommitSourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCognitoSource) DeepCopyInto(out *AWSCognitoSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCognitoSource.
func (in *AWSCognitoSource) DeepCopy() *AWSCognitoSource {
	if in == nil {
		return nil
	}
	out := new(AWSCognitoSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSCognitoSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCognitoSourceList) DeepCopyInto(out *AWSCognitoSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AWSCognitoSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCognitoSourceList.
func (in *AWSCognitoSourceList) DeepCopy() *AWSCognitoSourceList {
	if in == nil {
		return nil
	}
	out := new(AWSCognitoSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AWSCognitoSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCognitoSourceSpec) DeepCopyInto(out *AWSCognitoSourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	in.Credentials.DeepCopyInto(&out.Credentials)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCognitoSourceSpec.
func (in *AWSCognitoSourceSpec) DeepCopy() *AWSCognitoSourceSpec {
	if in == nil {
		return nil
	}
	out := new(AWSCognitoSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCognitoSourceStatus) DeepCopyInto(out *AWSCognitoSourceStatus) {
	*out = *in
	in.SourceStatus.DeepCopyInto(&out.SourceStatus)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCognitoSourceStatus.
func (in *AWSCognitoSourceStatus) DeepCopy() *AWSCognitoSourceStatus {
	if in == nil {
		return nil
	}
	out := new(AWSCognitoSourceStatus)
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
func (in *AWSDynamoDBSourceStatus) DeepCopyInto(out *AWSDynamoDBSourceStatus) {
	*out = *in
	in.SourceStatus.DeepCopyInto(&out.SourceStatus)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSDynamoDBSourceStatus.
func (in *AWSDynamoDBSourceStatus) DeepCopy() *AWSDynamoDBSourceStatus {
	if in == nil {
		return nil
	}
	out := new(AWSDynamoDBSourceStatus)
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