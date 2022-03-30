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

type ServiceConfig struct {
	LoRaDevice LoRaDeviceServiceConfig
}

type LoRaDeviceServiceConfig struct {
	StationID int
	Channel   int
	GroupID   int
}

func (sc *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	c, ok := rawConfig.(*ServiceConfig)
	if !ok {
		// log.Error("unable to cast raw config to type 'ServiceConfig'")
		return false
	}

	*sc = *c
	return true
}

func (c *LoRaDeviceServiceConfig) Validate() error {
	if c.StationID <= 0 || c.StationID >= 65535  {
		return fmt.Errorf("Station ID configuration %d out of valid ID range", c.StationID)
	}
	if c.Channel < 24 || c.Channel > 61  {
		return fmt.Errorf("Channel configuration %d out of valid range (24-61)", c.Channel)
	}
	if c.GroupID < 0 || c.GroupID > 65535  {
		return fmt.Errorf("Group ID configuration %d out of valid range (24-61)", c.GroupID)
	}
	return nil
}
