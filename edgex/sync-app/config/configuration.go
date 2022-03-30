//
// Copyright (c) 2022 Fujitsu Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package config

import (
	"fmt"
)

// TODO: Define your structured custom configuration types. Must be wrapped with an outer struct with
//       single element that matches the top level custom configuration element in your configuration.toml file,
//       'AppCustom' in this example. Replace this example with your configuration structure or
//       remove this file if not using structured custom configuration.
type ServiceConfig struct {
	Sync SyncConfig
}

// TODO: This should be dynamically learned and pipelines created as needed
type SyncConfig struct {
	DeviceNames     string
	ResourceNames   string
	DestinationHost string
}

func (c *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false
	}

	*c = *configuration

	return true
}

func (sc *SyncConfig) Validate() error {
	if len(sc.DeviceNames) == 0 {
		return fmt.Errorf("SourceNames configuration must not be blank")
	}
	if len(sc.ResourceNames) == 0 {
		return fmt.Errorf("ResourceNames configuration must not be blank")
	}
	if len(sc.DestinationHost) == 0 {
		return fmt.Errorf("DestinationHost configuration must not be blank")
	}
	return nil
}
