// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
// Copyright (C) 2022 Fujitsu Limited
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/edgexfoundry/device-sdk-go/v2"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/startup"
	"github.com/edgexfoundry/device-lora/driver"
)

const (
	serviceName string = "device-lora"
)

func main() {
	d := driver.LoRaDriver{}
	startup.Bootstrap(serviceName, device.Version, &d)
}
