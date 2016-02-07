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
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/kubernetes/deployment-manager/pkg/common"
)

// ParseKubernetesObject parses a Kubernetes API object in YAML format.
func ParseKubernetesObject(object []byte) (*common.Resource, error) {
	o := &common.KubernetesObject{}
	if err := yaml.Unmarshal(object, &o); err != nil {
		return nil, getParseError(err)
	}

	// Ok, it appears to be a valid object, create a Resource out of it.
	r := &common.Resource{}
	var name string
	if rawN, ok := o.Metadata["name"]; ok {
		if name, ok = rawN.(string); !ok {
			return nil, getParseError(fmt.Errorf("name is not a string: %#v", rawN))
		}
	}

	r.Name = getRandomName(name)
	r.Type = GetTypeForKubernetesVersionAndKind(o.APIVersion, o.Kind)

	r.Properties = make(map[string]interface{})
	if err := yaml.Unmarshal(object, &r.Properties); err != nil {
		return nil, getParseError(err)
	}

	return r, nil
}

func getParseError(err error) error {
	return fmt.Errorf("cannot unmarshal native kubernetes object: %#v", err)
}

func getRandomName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}

// ConvertKubernetesResourceTypes converts the Kubernetes API version and kind
// to a qualified resource type name for all resources in a configuration whose
// type names are known Kubernetes kinds.
func ConvertKubernetesResourceTypes(config *common.Configuration) {
	for _, r := range config.Resources {
		if IsKubernetesKind(strings.ToLower(r.Type)) {
			if rawV, ok := r.Properties["apiVersion"]; ok {
				if version, ok := rawV.(string); ok {
					if rawK, ok := r.Properties["kind"]; ok {
						if kind, ok := rawK.(string); ok {
							newType := GetTypeForKubernetesVersionAndKind(version, kind)
							r.Type = newType
						}
					}
				}
			}
		}
	}
}

// GetTypeForKubernetesVersionAndKind converts a Kubernetes API version and
// kind to a qualified resource type name. Using qualified resource type names
// instead of raw Kubernetes kinds lets us distinguish between resource types
// processed by kubectl and other classes of resource types.
func GetTypeForKubernetesVersionAndKind(version, kind string) string {
	return fmt.Sprintf("kubernetes.%s.%s", strings.ToLower(version), strings.ToLower(kind))
}

// IsKubernetesKind returns true if a type name is a Kubernetes kind.
func IsKubernetesKind(typeName string) bool {
	return kubernetesKinds[typeName]
}

// kubernetesKinds identifies primitive API object kinds as of Kubernetes 1.1.
var kubernetesKinds = map[string]bool{
	"binding":                   true,
	"componentstatus":           true,
	"componentstatuslist":       true,
	"deleteoptions":             true,
	"endpoints":                 true,
	"endpointslist":             true,
	"event":                     true,
	"eventlist":                 true,
	"limitrange":                true,
	"limitrangelist":            true,
	"namespace":                 true,
	"namespacelist":             true,
	"node":                      true,
	"nodelist":                  true,
	"persistentvolume":          true,
	"persistentvolumeclaim":     true,
	"persistentvolumeclaimlist": true,
	"persistentvolumelist":      true,
	"pod":                       true,
	"podlist":                   true,
	"podtemplate":               true,
	"podtemplatelist":           true,
	"replicationcontroller":     true,
	"replicationcontrollerlist": true,
	"resourcequota":             true,
	"resourcequotalist":         true,
	"secret":                    true,
	"secretlist":                true,
	"service":                   true,
	"serviceaccount":            true,
	"serviceaccountlist":        true,
	"servicelist":               true,
}
