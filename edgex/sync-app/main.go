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

package main

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"sync-app/config"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	serviceKey = "sync-app"
)

type SyncApp struct {
	service       interfaces.ApplicationService
	lc            logger.LoggingClient
	serviceConfig *config.ServiceConfig
	configChanged chan bool
}

type Data struct {
	DeviceName   string      `json:"devicename"`
	ResourceName string      `json:"resourcename"`
	ValueType    string      `json:"valueType"`
	Value        string      `json:"value"`
	BinaryValue  []byte      `json:"binaryValue"`
	ObjectValue  interface{} `json:"objectValue"`
}

type SyncFunctions struct {
	destURL string
}

func NewSyncFunctions(dest string) SyncFunctions {
	return SyncFunctions{
		destURL: dest,
	}
}

func main() {
	app := SyncApp{}
	code := app.CreateAndRunAppService(serviceKey, pkg.NewAppService)
	os.Exit(code)
}

func (app *SyncApp) CreateAndRunAppService(serviceKey string, newServiceFactory func(string) (interfaces.ApplicationService, bool)) int {
	var ok bool
	app.service, ok = newServiceFactory(serviceKey)
	if !ok {
		return -1
	}
	app.lc = app.service.LoggingClient()

	app.serviceConfig = &config.ServiceConfig{}
	if err := app.service.LoadCustomConfig(app.serviceConfig, "Sync"); err != nil {
		app.lc.Errorf("failed load custom configuration: %s", err.Error())
		return -1
	}

	if err := app.serviceConfig.Sync.Validate(); err != nil {
		app.lc.Errorf("custom configuration failed validation: %s", err.Error())
		return -1
	}

	deviceNames := make([]string, 0)
	for _, name := range strings.Split(app.serviceConfig.Sync.DeviceNames, ",") {
		deviceNames = append(deviceNames, strings.TrimSpace(name))
	}
	resourceNames := make([]string, 0)
	for _, name := range strings.Split(app.serviceConfig.Sync.ResourceNames, ",") {
		resourceNames = append(resourceNames, strings.TrimSpace(name))
	}

	sync := NewSyncFunctions(app.serviceConfig.Sync.DestinationHost)
	err := app.service.SetFunctionsPipeline(
		transforms.NewFilterFor(deviceNames).FilterByDeviceName,
		transforms.NewFilterFor(resourceNames).FilterByResourceName,
		sync.createDataGroup,
		sync.SendHTTP)

	if err != nil {
		app.lc.Errorf("SetFunctionsPipeline returned error: %s", err.Error())
		return -1
	}

	if err := app.service.MakeItRun(); err != nil {
		app.lc.Errorf("MakeItRun returned error: %s", err.Error())
		return -1
	}

	return 0
}

func (sync *SyncFunctions) createDataGroup(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		return false, errors.New("No data received")
	}

	event, ok := data.(dtos.Event)
	if !ok {
		return false, errors.New("Data received is not the expected 'dtos.Event' type")
	}

	var dataGroup []Data

	for _, reading := range event.Readings {
		var data Data
		switch reading.ValueType {

		case "Binary":
			data = Data{
				DeviceName:   reading.DeviceName,
				ResourceName: reading.ResourceName,
				ValueType:    reading.ValueType,
				BinaryValue:  reading.BinaryValue,
			}

		case "Object":
			data = Data{
				DeviceName:   reading.DeviceName,
				ResourceName: reading.ResourceName,
				ValueType:    reading.ValueType,
				ObjectValue:  reading.ObjectValue,
			}

		default:
			data = Data{
				DeviceName:   reading.DeviceName,
				ResourceName: reading.ResourceName,
				ValueType:    reading.ValueType,
				Value:        reading.Value,
			}
		}

		dataGroup = append(dataGroup, data)
	}

	return true, dataGroup
}

func (sync *SyncFunctions) SendHTTP(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := ctx.LoggingClient()

	if data == nil {
		lc.Errorf("No data received")
		return false, errors.New("No data received")
	}

	datagroup, ok := data.([]Data)
	if !ok {
		lc.Errorf("Data received is not the expected '[]Data' type")
		return false, errors.New("Data received is not the expected '[]Data' type")
	}

	for _, jsonData := range datagroup {
		// Forward data to the specified destination
		lc.Infof("Send %v to %s", jsonData, sync.destURL)

		switch jsonData.ValueType {

		case "Binary":
			sender := transforms.NewHTTPSender("http://"+sync.destURL+":59986/api/v2/resource/sample-image/jpeg", "image/jpeg", false)
			sendData := jsonData.BinaryValue
			sender.HTTPPost(ctx, sendData)

		default:
			sender := transforms.NewHTTPSender("http://"+sync.destURL+":59986/api/v2/resource/sample-json/json", "application/json", false)
			sendData, _ := json.Marshal(jsonData)
			sender.HTTPPost(ctx, sendData)
		}

	}
	return true, data
}
