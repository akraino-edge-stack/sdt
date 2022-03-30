// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Fujitsu Limited
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"sync"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/device-lora/config"
)

const (
	ProtocolName = "lora"
	StationIDKey = "ID"
)

type LoRaDriver struct {
	lc                logger.LoggingClient
	serviceConfig     *config.ServiceConfig
	asyncValues       chan<- *sdkModels.AsyncValues
	discoveredDevice  chan<- []sdkModels.DiscoveredDevice
	// serviceConfig     *config.ServiceConfig
	sensorMutex       sync.Mutex
	sensors           map[string]*LoRaSensor
	txch              chan LoRaMessage
	rxch              chan LoRaMessage
	transport         *LoRaTransport
	watchersAdded     bool
}

func IDFromProtocols(protocols map[string]models.ProtocolProperties) (int, bool) {
	var id int
	props, ok := protocols[ProtocolName]
	if ok {
		idStr, idok := props[StationIDKey]
		if idok {
			convid, err := strconv.Atoi(idStr)
			if err == nil {
				id = convid
			} else {
				ok = false
			}
		} else {
			ok = false
		}
	}
	return id, ok
}

// Handle messages from the LoRa device
func (driver *LoRaDriver) HandleLoRa() error {
	for msg := range driver.rxch {
		driver.lc.Infof("Received from %d: %s", msg.id, msg.msg)
		deviceName := "Lora-" + strconv.Itoa(msg.id)
		driver.sensorMutex.Lock()
		sensor, ok := driver.sensors[deviceName]
		if ok && sensor != nil {
			driver.sensorMutex.Unlock()
			// Convert received data to values using the
			// appropriate sensor
			values := sensor.TranslateReceived(deviceName, msg.msg)
			if values != nil {
				driver.asyncValues <- values
			}
		} else {
			driver.lc.Infof("Device %s discovered", deviceName)
			sensor = NewLoRaSensor(driver.lc, msg.id)
			sensor.registered = false
			driver.sensors[deviceName] = sensor
			driver.sensorMutex.Unlock()
		}
	}
	driver.lc.Info("LoRa transport closed")
	close(driver.asyncValues)
	return nil
}

func (driver *LoRaDriver) AddProvisionWatchers() {
	watcherName := "LoRa-Provision-Watcher"
	devService := service.RunningService()
	if _, err := devService.GetProvisionWatcherByName(watcherName); err == nil {
		// Watcher already exists (process restarted?)
		driver.watchersAdded = true
		return
	}

	identifiers := map[string]string{StationIDKey: "[0-9]+"}
	watcher := models.ProvisionWatcher{
		Name:        watcherName,
		Identifiers: identifiers,
		ProfileName: "LoRa-Device",
		ServiceName: "device-lora",
		AdminState:  models.Unlocked,
	}
	_, err := devService.AddProvisionWatcher(watcher)
	if err != nil {
		driver.lc.Errorf("Error adding provision watcher: %s", err.Error())
	} else {
		driver.watchersAdded = true
	}
}

func (driver *LoRaDriver) AddExistingDevices() {
	devService := service.RunningService()
	sensorDevices := devService.Devices()
	for _, sensorDevice := range sensorDevices {
		deviceName := sensorDevice.Name
		driver.lc.Infof("Adding registered device %s", deviceName)
		// Get the station ID
		id, ok := IDFromProtocols(sensorDevice.Protocols)
		if ok {
			sensor := NewLoRaSensor(driver.lc, id)
			driver.sensors[deviceName] = sensor
		} else {
			driver.lc.Errorf("Device %s has no station ID", deviceName)
		}
	}
}

// Perform protocol-specific initialization for the device service
func (driver *LoRaDriver) Initialize(lc logger.LoggingClient, asyncValues chan<- *sdkModels.AsyncValues, discoveredDevice chan<- []sdkModels.DiscoveredDevice) error {
	driver.lc = lc
	driver.lc.Info("Initializing LoRaDriver")
	driver.serviceConfig = &config.ServiceConfig{}
	driver.asyncValues = asyncValues
	driver.discoveredDevice = discoveredDevice
	driver.rxch = make(chan LoRaMessage)
	driver.txch = make(chan LoRaMessage)
	driver.transport = NewLoRaTransport(lc, driver.rxch, driver.txch)
	driver.sensors = make(map[string]*LoRaSensor, 1)

	// Provision watchers are not added here because the device profile
	// is added after the driver is initialized.
	driver.watchersAdded = false

	// Add all devices already present in the metadata
	driver.AddExistingDevices()

	// Start handling asynchronous readings
	go driver.HandleLoRa()

	s := service.RunningService()
	if err := s.LoadCustomConfig(driver.serviceConfig, "LoRaDevice"); err != nil {
		return fmt.Errorf("Error loading configuration: %s", err.Error())
	}
	if err := driver.serviceConfig.LoRaDevice.Validate(); err != nil {
		return fmt.Errorf("Config validation failed: %s", err.Error())
	}
	// Reflect configuration to transport
	driver.transport.config.ownID = driver.serviceConfig.LoRaDevice.StationID
	driver.transport.config.channel = driver.serviceConfig.LoRaDevice.Channel
	driver.transport.config.groupID = driver.serviceConfig.LoRaDevice.GroupID

	// Finally, start the LoRa transport layer
	return driver.transport.Start()
}

func (driver *LoRaDriver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModels.CommandRequest) (res []*sdkModels.CommandValue, err error) {
	res = make([]*sdkModels.CommandValue, 0)
	for _, req := range reqs {
		driver.lc.Debugf("Read from deviceName: %s protocols: %v resource: %v attributes: %v", deviceName, protocols, req.DeviceResourceName, req.Attributes)
		sensor, ok := driver.sensors[deviceName]
		if ok {
			cv, err := sensor.GetValue(req.DeviceResourceName)
			if err == nil {
				res = append(res, cv)
			} else {
				driver.lc.Errorf("Error getting value: %s", err.Error())
			}
		} else {
			driver.lc.Errorf("Device %s not recognized", deviceName)
		}
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("No data")
	}
	return res, nil
}

func (driver *LoRaDriver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModels.CommandRequest, params []*sdkModels.CommandValue) error {
	driver.lc.Infof("Write to deviceName: %s protocols: %v", deviceName, protocols)
	sensor, ok := driver.sensors[deviceName]
	if !ok {
		driver.lc.Errorf("Device %s not recognized", deviceName)
		return fmt.Errorf("Device %s not recognized", deviceName)
	}
	for _, param := range params {
		driver.lc.Infof("Write value resource: %s Type: %s", param.DeviceResourceName, param.Type)
		msgString, err := sensor.WriteValue(param)
		if err == nil {
			if len(msgString) > 0 {
				driver.txch <- LoRaMessage{
					id: sensor.stationID,
					msg: msgString,
				}
			}
		} else {
			return err
		}
	}
	return nil
}

func (driver *LoRaDriver) Stop(force bool) error {
	// Then Logging Client might not be initialized
	if driver.lc != nil {
		driver.lc.Infof("LoRaDriver.Stop called: force=%v", force)
	}
	// Closing the TX channel causes the transport to stop and
	// close the RX channel, which in turn causes HandleLoRa to close
	// the asyncValues channel
	close(driver.txch)
	return nil
}

func (driver *LoRaDriver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	driver.lc.Infof("Device %s added", deviceName)
	id, ok := IDFromProtocols(protocols)
	if ok {
		driver.lc.Infof("Device %s station ID %d", deviceName, id)
		sensor := NewLoRaSensor(driver.lc, id)
		driver.sensorMutex.Lock()
		// May actually be replacing an unregistered version
		driver.sensors[deviceName] = sensor
		driver.sensorMutex.Unlock()
	} else {
		return fmt.Errorf("Device %s could not be added: missing station ID", deviceName)
	}
	return nil
}

func (driver *LoRaDriver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	driver.lc.Infof("Device %s updated", deviceName)
	return nil
}

func (driver *LoRaDriver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	driver.lc.Infof("Device %s removed", deviceName)
	driver.sensorMutex.Lock()
	_, ok := driver.sensors[deviceName]
	if ok {
		delete(driver.sensors, deviceName)
	}
	driver.sensorMutex.Unlock()
	return nil
}

func (driver *LoRaDriver) Discover() {
	driver.lc.Debug("Discovery")
	if !driver.watchersAdded {
		// Add provision watchers to enable device discovery
		driver.AddProvisionWatchers()
	}

	if driver.watchersAdded {
		driver.sensorMutex.Lock()
		for deviceName, sensor := range driver.sensors {
			if !sensor.registered {
				proto := make(map[string]models.ProtocolProperties)
				proto[ProtocolName] = map[string]string{StationIDKey: strconv.Itoa(sensor.stationID)}
				device := sdkModels.DiscoveredDevice{
					Name:        deviceName,
					Protocols:   proto,
					Description: "found by discovery",
					Labels:      []string{"auto-discovery"},
				}
				devices := []sdkModels.DiscoveredDevice{device}
				driver.lc.Infof("Register device %s", deviceName)
				driver.discoveredDevice <- devices
				// DeviceAdd will be called if the device
				// passes the provision watcher filter
				// (see the device-sdk-go/example/README.md)
			}
		}
		driver.sensorMutex.Unlock()
	}
}
