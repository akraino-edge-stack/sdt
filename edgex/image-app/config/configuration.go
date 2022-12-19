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
	Image ImageConfig
}

// TODO: This should be dynamically learned and pipelines created as needed
type ImageConfig struct {
	ProfileNames     string
	DeviceNames      string
	ResourceNames    string
	CrowdedThreshold uint16
}

func (c *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false
	}

	*c = *configuration

	return true
}

func (ic *ImageConfig) Validate() error {
	if len(ic.ProfileNames) == 0 {
		return fmt.Errorf("ProfileNames configuration must not be blank")
	}
	if len(ic.DeviceNames) == 0 {
		return fmt.Errorf("DeviceNames configuration must not be blank")
	}
	if len(ic.ResourceNames) == 0 {
		return fmt.Errorf("ResourceNames configuration must not be blank")
	}
	if ic.CrowdedThreshold == 0 {
		return fmt.Errorf("CrowdedThreshold configuration is invalid")
	}
	return nil
}
