// Copyright 2017 The Oracle Kubernetes Cloud Controller Manager Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bmcs

import (
	"reflect"
	"testing"

	baremetal "github.com/oracle/bmcs-go-sdk"
)

var getBackendModificationsTestCases = []struct {
	desired   baremetal.BackendSet
	actual    baremetal.BackendSet
	additions []baremetal.Backend
	removals  []baremetal.Backend
}{
	{
		desired: baremetal.BackendSet{
			Backends: []baremetal.Backend{
				{IPAddress: "0.0.0.0", Port: 80},
				{IPAddress: "0.0.0.1", Port: 80},
			},
		},
		actual: baremetal.BackendSet{
			Backends: []baremetal.Backend{
				{IPAddress: "0.0.0.0", Port: 80},
			},
		},
		additions: []baremetal.Backend{{IPAddress: "0.0.0.1", Port: 80}},
		removals:  []baremetal.Backend{},
	}, {
		desired: baremetal.BackendSet{
			Backends: []baremetal.Backend{
				{IPAddress: "0.0.0.0", Port: 80},
			},
		},
		actual: baremetal.BackendSet{
			Backends: []baremetal.Backend{
				{IPAddress: "0.0.0.0", Port: 80},
				{IPAddress: "0.0.0.1", Port: 80},
			},
		},
		additions: []baremetal.Backend{},
		removals:  []baremetal.Backend{{IPAddress: "0.0.0.1", Port: 80}},
	}, {
		desired: baremetal.BackendSet{
			Backends: []baremetal.Backend{
				{IPAddress: "0.0.0.0", Port: 80},
			},
		},
		actual: baremetal.BackendSet{
			Backends: []baremetal.Backend{
				{IPAddress: "0.0.0.1", Port: 80},
			},
		},
		additions: []baremetal.Backend{{IPAddress: "0.0.0.0", Port: 80}},
		removals:  []baremetal.Backend{{IPAddress: "0.0.0.1", Port: 80}},
	}, {
		desired: baremetal.BackendSet{
			Backends: []baremetal.Backend{
				{IPAddress: "0.0.0.0", Port: 443},
				{IPAddress: "0.0.0.1", Port: 443},
			},
		},
		actual: baremetal.BackendSet{
			Backends: []baremetal.Backend{
				{IPAddress: "0.0.0.0", Port: 80},
				{IPAddress: "0.0.0.1", Port: 80},
			},
		},
		additions: []baremetal.Backend{
			{IPAddress: "0.0.0.0", Port: 443},
			{IPAddress: "0.0.0.1", Port: 443},
		},
		removals: []baremetal.Backend{
			{IPAddress: "0.0.0.0", Port: 80},
			{IPAddress: "0.0.0.1", Port: 80},
		},
	}, {
		desired: baremetal.BackendSet{
			Backends: []baremetal.Backend{
				{IPAddress: "0.0.0.0", Port: 80},
			},
		},
		actual: baremetal.BackendSet{
			Backends: []baremetal.Backend{
				{IPAddress: "0.0.0.0", Port: 80},
			},
		},
		additions: nil,
		removals:  nil,
	},
}

func TestGetBackendModifications(t *testing.T) {
	for _, tt := range getBackendModificationsTestCases {
		additions, removals := getBackendModifications(tt.desired, tt.actual)
		if !reflect.DeepEqual(additions, tt.additions) && !reflect.DeepEqual(removals, tt.removals) {
			t.Errorf("getBackendModifications(%v, %v) => (%v, %v), want (%v, %v)",
				tt.desired, tt.actual, additions, removals, tt.additions, tt.removals)
		}
	}
}

var getListenerModificationsTestCases = []struct {
	desired   map[string]baremetal.Listener
	actual    map[string]baremetal.Listener
	additions []baremetal.Listener
	removals  []baremetal.Listener
}{
	{
		desired: map[string]baremetal.Listener{"TCP-443": baremetal.Listener{
			Name: "TCP-443",
			DefaultBackendSetName: "TCP-443",
			Protocol:              "TCP",
			Port:                  443,
		}},
		actual: map[string]baremetal.Listener{"TCP-80": baremetal.Listener{
			Name: "TCP-80",
			DefaultBackendSetName: "TCP-80",
			Protocol:              "TCP",
			Port:                  80,
		}},
		additions: []baremetal.Listener{baremetal.Listener{
			Name: "TCP-443",
			DefaultBackendSetName: "TCP-443",
			Protocol:              "TCP",
			Port:                  443,
		}},
		removals: []baremetal.Listener{baremetal.Listener{
			Name: "TCP-80",
			DefaultBackendSetName: "TCP-80",
			Protocol:              "TCP",
			Port:                  80,
		}},
	}, {
		desired: map[string]baremetal.Listener{"TCP-80": baremetal.Listener{
			Name: "TCP-80",
			DefaultBackendSetName: "TCP-80",
			Protocol:              "TCP",
			Port:                  80,
		}},
		actual: map[string]baremetal.Listener{"TCP-80": baremetal.Listener{
			Name: "TCP-80",
			DefaultBackendSetName: "TCP-80",
			Protocol:              "TCP",
			Port:                  80,
		}},
		additions: nil,
		removals:  nil,
	}, {
		desired: map[string]baremetal.Listener{
			"TCP-80": baremetal.Listener{
				Name: "TCP-80",
				DefaultBackendSetName: "TCP-80",
				Protocol:              "TCP",
				Port:                  80,
			},
			"TCP-443": baremetal.Listener{
				Name: "TCP-443",
				DefaultBackendSetName: "TCP-443",
				Protocol:              "TCP",
				Port:                  443,
			},
		},
		actual: map[string]baremetal.Listener{"TCP-80": baremetal.Listener{
			Name: "TCP-80",
			DefaultBackendSetName: "TCP-80",
			Protocol:              "TCP",
			Port:                  80,
		}},
		additions: []baremetal.Listener{baremetal.Listener{
			Name: "TCP-443",
			DefaultBackendSetName: "TCP-443",
			Protocol:              "TCP",
			Port:                  443,
		}},
		removals: nil,
	}, {
		desired: map[string]baremetal.Listener{"TCP-80": baremetal.Listener{
			Name: "TCP-80",
			DefaultBackendSetName: "TCP-80",
			Protocol:              "TCP",
			Port:                  80,
		}},
		actual: map[string]baremetal.Listener{
			"TCP-80": baremetal.Listener{
				Name: "TCP-80",
				DefaultBackendSetName: "TCP-80",
				Protocol:              "TCP",
				Port:                  80,
			},
			"TCP-443": baremetal.Listener{
				Name: "TCP-443",
				DefaultBackendSetName: "TCP-443",
				Protocol:              "TCP",
				Port:                  443,
			},
		},
		additions: nil,
		removals: []baremetal.Listener{baremetal.Listener{
			Name: "TCP-443",
			DefaultBackendSetName: "TCP-443",
			Protocol:              "TCP",
			Port:                  443,
		}},
	},
}

func TestGetListenerModifications(t *testing.T) {
	for _, tt := range getListenerModificationsTestCases {
		additions, removals, err := getListenerModifications(tt.desired, tt.actual)
		if err != nil {
			t.Errorf("getListenerModifications(%v, %v) => error: %v", tt.desired, tt.actual, err)
			continue
		}
		if !reflect.DeepEqual(additions, tt.additions) && !reflect.DeepEqual(removals, tt.removals) {
			t.Errorf("getListenerModifications(%v, %v) => (%v, %v), want (%v, %v)",
				tt.desired, tt.actual, additions, removals, tt.additions, tt.removals)
		}
	}
}