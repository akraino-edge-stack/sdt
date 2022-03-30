// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Fujitsu Limited
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

        "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

const (
	TemperatureResource = "Temperature"
	HumidityResource    = "Humidity"
	BothCommand         = "TemperatureAndHumidity"
	TemperatureField    = "temp"
	HumidityField       = "hum"
	IntervalResource    = "Interval"
	IntervalField       = "int"
)

type LoRaSensor struct {
        lc          logger.LoggingClient
	registered  bool
	stationID   int
	temperature float32
	humidity    float32
	tempOrigin  int64
	humOrigin   int64
	interval    float32
	lastReading time.Time
}

func (sensor *LoRaSensor) GetValue(resourceName string) (cv *sdkModels.CommandValue, err error) {
	switch resourceName {
	case TemperatureResource:
		if sensor.tempOrigin == 0 {
			err = fmt.Errorf("No data")
		} else {
			cv, err = sdkModels.NewCommandValueWithOrigin(resourceName, common.ValueTypeFloat32, sensor.temperature, sensor.tempOrigin)
		}
	case HumidityResource:
		if sensor.humOrigin == 0 {
			err = fmt.Errorf("No data")
		} else {
			cv, err = sdkModels.NewCommandValueWithOrigin(resourceName, common.ValueTypeFloat32, sensor.humidity, sensor.humOrigin)
		}
	case IntervalResource:
		if sensor.interval == 0.0 {
			err = fmt.Errorf("No data")
		} else {
			cv, err = sdkModels.NewCommandValueWithOrigin(resourceName, common.ValueTypeFloat32, sensor.interval, sensor.lastReading.UnixNano())
		}
	default:
		err = fmt.Errorf("Unrecognized resource %s", resourceName)
	}
	return cv, err
}

func (sensor *LoRaSensor) WriteValue(value *sdkModels.CommandValue) (string, error) {
	if value.DeviceResourceName != IntervalResource {
		return "", fmt.Errorf("%s is not a writable resource", value.DeviceResourceName)
	}
	if value.Type != common.ValueTypeFloat32 {
		return "", fmt.Errorf("%s is not the correct type for %s", value.Type, value.DeviceResourceName)
	}
	sensor.lc.Infof("Set interval to %v", value.Value.(float32))
	msgObj := map[string]interface{} {
		IntervalField: value.Value,
	}
	msgBytes, err := json.Marshal(msgObj)
	if err != nil {
		return "", err
	}
	return string(msgBytes), nil
}

func (sensor *LoRaSensor) CheckFloat32(v interface{}) error {
	if _, isFloat64 := v.(float64); !isFloat64 {
		return fmt.Errorf("Not a number")
	}
	f64 := v.(float64)
	if math.Abs(f64) > math.MaxFloat32 {
		return fmt.Errorf("Out of range")
	}
	return nil
}

func (sensor *LoRaSensor) UpdateInterval() {
	sinceLast := time.Since(sensor.lastReading)
	if sinceLast.Hours() < 1.0 {
		newInterval := float32(sinceLast.Seconds())
		if sensor.interval == 0.0 {
			sensor.interval = newInterval
		} else {
			// Running average unless change is > 25%
			if math.Abs(float64(newInterval - sensor.interval)) > float64(0.25 * sensor.interval) {
				sensor.interval = newInterval
			} else {
				sensor.interval = (newInterval + sensor.interval) * 0.5
			}
		}
	}
	sensor.lastReading = time.Now()
}

func (sensor *LoRaSensor) TranslateReceived(deviceName string, msg string) *sdkModels.AsyncValues {
	if !sensor.registered {
		sensor.lc.Infof("Sensor %s not registered yet", deviceName)
		return nil
	}
	msgData := make(map[string]interface{})
	err := json.Unmarshal([]byte(msg), &msgData)
	if err != nil {
		sensor.lc.Infof("Unable to parse message \"%s\": %s", msg, err.Error())
		return nil
	}
	sensor.UpdateInterval()
	origin := time.Now().UnixNano()
	sourceName := ""
	var cmdValues []*sdkModels.CommandValue
	for k, v := range msgData {
		switch k {
		case TemperatureField:
			err = sensor.CheckFloat32(v)
			if err != nil {
				sensor.lc.Infof("Temperature error: %s", err.Error())
			} else {
				resourceName := TemperatureResource
				sensor.temperature = float32(v.(float64))
				sensor.tempOrigin = origin
				value, verr := sdkModels.NewCommandValueWithOrigin(resourceName, common.ValueTypeFloat32, sensor.temperature, origin)
				if verr == nil {
					cmdValues = append(cmdValues, value)
					if sourceName == "" {
						sourceName = resourceName
					} else if sourceName == HumidityResource {
						sourceName = BothCommand
					}
				}
			}
		case HumidityField:
			err = sensor.CheckFloat32(v)
			if err != nil {
				sensor.lc.Infof("Humidity error: %s", err.Error())
			} else {
				resourceName := HumidityResource
				sensor.humidity = float32(v.(float64))
				sensor.humOrigin = origin
				value, verr := sdkModels.NewCommandValueWithOrigin(resourceName, common.ValueTypeFloat32, sensor.humidity, origin)
				if verr == nil {
					cmdValues = append(cmdValues, value)
					if sourceName == "" {
						sourceName = resourceName
					} else if sourceName == TemperatureResource {
						sourceName = BothCommand
					}
				}
			}
		default:
			sensor.lc.Infof("Unrecognized field %s", k)
		}
	}
	if len(cmdValues) == 0 {
		return nil
	}
	values := &sdkModels.AsyncValues{
		DeviceName:    deviceName,
		SourceName:    sourceName,
		CommandValues: cmdValues,
	}
	return values
}

func NewLoRaSensor(lc logger.LoggingClient, id int) *LoRaSensor {
	sensor := &LoRaSensor{}
	sensor.lc = lc
	sensor.registered = true
	sensor.stationID = id
	sensor.temperature = 0.0
	sensor.humidity = 0.0
	return sensor
}

