/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package util

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"

	"github.com/kubernetes/deployment-manager/pkg/common"
)

var serviceInput = `
  kind: "Service"
  apiVersion: "v1"
  metadata:
    name: "mock"
    labels:
      app: "mock"
  spec:
    ports:
      -
        protocol: "TCP"
        port: 99
        targetPort: 9949
    selector:
      app: "mock"
`

var serviceExpected = `
name: mock
type: kubernetes.v1.service
properties:
    kind: "Service"
    apiVersion: "v1"
    metadata:
      name: "mock"
      labels:
        app: "mock"
    spec:
      ports:
        -
          protocol: "TCP"
          port: 99
          targetPort: 9949
      selector:
        app: "mock"
`

var rcInput = `
  kind: "ReplicationController"
  apiVersion: "v1"
  metadata:
    name: "mockname"
    labels:
      app: "mockapp"
      foo: "bar"
  spec:
    replicas: 1
    selector:
      app: "mockapp"
    template:
      metadata:
        labels:
          app: "mocklabel"
      spec:
        containers:
          -
            name: "mock-container"
            image: "kubernetes/pause"
            ports:
              -
                containerPort: 9949
                protocol: "TCP"
              -
                containerPort: 9949
                protocol: "TCP"
`

var rcExpected = `
name: mockname
type: kubernetes.v1.replicationcontroller
properties:
    kind: "ReplicationController"
    apiVersion: "v1"
    metadata:
      name: "mockname"
      labels:
        app: "mockapp"
        foo: "bar"
    spec:
      replicas: 1
      selector:
        app: "mockapp"
      template:
        metadata:
          labels:
            app: "mocklabel"
        spec:
          containers:
            -
              name: "mock-container"
              image: "kubernetes/pause"
              ports:
                -
                  containerPort: 9949
                  protocol: "TCP"
                -
                  containerPort: 9949
                  protocol: "TCP"
`

func unmarshalResource(t *testing.T, object string) *common.Resource {
	r := &common.Resource{}
	if err := yaml.Unmarshal([]byte(object), &r); err != nil {
		t.Fatalf("cannot unmarshal test object: %s", err)
	}

	return r
}

func testConversion(t *testing.T, input, expected string) {
	e := unmarshalResource(t, expected)
	result, err := ParseKubernetesObject([]byte(input))
	if err != nil {
		t.Fatalf("ParseKubernetesObject failed: %s", err)
	}
	// Since the object name gets created on the fly, we have to rejigger the returned object
	// slightly to make sure the DeepEqual works as expected.
	// First validate the name matches the expected format.
	var i int
	format := e.Name + "-%d"
	count, err := fmt.Sscanf(result.Name, format, &i)
	if err != nil || count != 1 {
		t.Errorf("unexpected name format, want:%s have:%s", format, result.Name)
	}
	e.Name = result.Name
	if !reflect.DeepEqual(result, e) {
		t.Errorf("want:\n%s\nhave:\n%s\n", ToYAMLOrError(e), ToYAMLOrError(result))
	}

}

func TestSimple(t *testing.T) {
	testConversion(t, rcInput, rcExpected)
	testConversion(t, serviceInput, serviceExpected)
}

func TestVersionAndKindToTypeName(t *testing.T) {
	inputConfig := &common.Configuration{Resources: []*common.Resource{
		parseKubernetesObjectOrDie(t, []byte(serviceInput)),
		parseKubernetesObjectOrDie(t, []byte(rcInput)),
	}}

	expectedConfig := &common.Configuration{Resources: []*common.Resource{
		unmarshalResource(t, serviceExpected),
		unmarshalResource(t, rcExpected),
	}}

	ConvertKubernetesResourceTypes(inputConfig)
	if !reflect.DeepEqual(inputConfig, expectedConfig) {
		t.Errorf("want:\n%s\nhave:\n%s\n", ToYAMLOrError(expectedConfig), ToYAMLOrError(inputConfig))
	}
}

func parseKubernetesObjectOrDie(t *testing.T, object []byte) *common.Resource {
	o := &common.KubernetesObject{}
	if err := yaml.Unmarshal(object, &o); err != nil {
		dieWithParseError(t, err)
	}

	r := &common.Resource{}
	var name string
	if rawN, ok := o.Metadata["name"]; ok {
		if name, ok = rawN.(string); !ok {
			dieWithParseError(t, fmt.Errorf("name is not a string: %#v", rawN))
		}
	}

	r.Name = name
	r.Type = o.Kind
	r.Properties = make(map[string]interface{})
	if err := yaml.Unmarshal(object, &r.Properties); err != nil {
		dieWithParseError(t, err)
	}

	return r
}

func dieWithParseError(t *testing.T, err error) {
	t.Fatalf("cannot unmarshal native kubernetes object: %#v", err)
}
