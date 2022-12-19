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
	"errors"
	"image"
	"os"
	"strings"

	"image-app/config"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/cuda"
)

const (
	serviceKey = "image-app"
)

// used to save the number of quadrangles in the image when the ResourceNname is onvifsnapshot
var selfQuadrangleCount uint16 = 0

// used to save the number of quadrangles in the image when the ResourceNname is sample-image
var otherQuadrangleCount uint16 = 0

// used to save the CrowdedThreshold in configuration.toml
var CrowdedThreshold uint16

// used to save current node's camera device name
var DeviceName string = ""

type ImageApp struct {
	service       interfaces.ApplicationService
	lc            logger.LoggingClient
	serviceConfig *config.ServiceConfig
	//	configChanged chan bool
}

type Data struct {
	ProfileName  string      `json:"profileName"`
	DeviceName   string      `json:"deviceName"`
	ResourceName string      `json:"resourceName"`
	ValueType    string      `json:"valueType"`
	Value        string      `json:"value"`
	BinaryValue  []byte      `json:"binaryValue"`
	ObjectValue  interface{} `json:"objectValue"`
}

func main() {
	app := ImageApp{}
	code := app.CreateAndRunAppService(serviceKey, pkg.NewAppService)
	os.Exit(code)
}

func (app *ImageApp) CreateAndRunAppService(serviceKey string, newServiceFactory func(string) (interfaces.ApplicationService, bool)) int {
	var ok bool
	app.service, ok = newServiceFactory(serviceKey)
	if !ok {
		return -1
	}
	app.lc = app.service.LoggingClient()

	app.serviceConfig = &config.ServiceConfig{}
	if err := app.service.LoadCustomConfig(app.serviceConfig, "Image"); err != nil {
		app.lc.Errorf("failed load custom configuration: %s", err.Error())
		return -1
	}

	if err := app.serviceConfig.Image.Validate(); err != nil {
		app.lc.Errorf("custom configuration failed validation: %s", err.Error())
		return -1
	}

	profileNames := make([]string, 0)
	for _, name := range strings.Split(app.serviceConfig.Image.ProfileNames, ",") {
		profileNames = append(profileNames, strings.TrimSpace(name))
	}
	deviceNames := make([]string, 0)
	for _, name := range strings.Split(app.serviceConfig.Image.DeviceNames, ",") {
		deviceNames = append(deviceNames, strings.TrimSpace(name))
	}
	resourceNames := make([]string, 0)
	for _, name := range strings.Split(app.serviceConfig.Image.ResourceNames, ",") {
		resourceNames = append(resourceNames, strings.TrimSpace(name))
	}

	CrowdedThreshold = app.serviceConfig.Image.CrowdedThreshold

	err := app.service.SetDefaultFunctionsPipeline(
		transforms.NewFilterFor(profileNames).FilterByProfileName,
		transforms.NewFilterFor(deviceNames).FilterByDeviceName,
		transforms.NewFilterFor(resourceNames).FilterByResourceName,
		createDataGroup,
		push2CoreData)

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

func createDataGroup(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		return false, errors.New("no data received")
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
				ProfileName:  reading.ProfileName,
				DeviceName:   reading.DeviceName,
				ResourceName: reading.ResourceName,
				ValueType:    reading.ValueType,
				BinaryValue:  reading.BinaryValue,
			}

		case "Object":
			data = Data{
				ProfileName:  reading.ProfileName,
				DeviceName:   reading.DeviceName,
				ResourceName: reading.ResourceName,
				ValueType:    reading.ValueType,
				ObjectValue:  reading.ObjectValue,
			}

		default:
			data = Data{
				ProfileName:  reading.ProfileName,
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

func push2CoreData(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := ctx.LoggingClient()

	if data == nil {
		lc.Errorf("No data received")
		return false, errors.New("no data received")
	}

	datagroup, ok := data.([]Data)
	if !ok {
		lc.Errorf("Data received is not the expected '[]Data' type")
		return false, errors.New("Data received is not the expected '[]Data' type")
	}

	for _, jsonData := range datagroup {
		// handle recived reading
		lc.Infof("Handle reading[profilename=%s devicename=%s resourcename=%s valuetype=%s]",
			jsonData.ProfileName, jsonData.DeviceName, jsonData.ResourceName, jsonData.ValueType)

		// save the camera device name
		if len(DeviceName) == 0 {
			if jsonData.DeviceName == "Camera001" {
				DeviceName = "Camera001"
			} else if jsonData.DeviceName == "Camera002" {
				DeviceName = "Camera002"
			} else {
				// do nothing
			}
		}

		switch jsonData.ValueType {

		case "Binary":
			// image analyse
			ok, quadrangleCount := imageAnalyse(ctx, jsonData.BinaryValue)
			if !ok {
				lc.Errorf("image analyse error")
				return false, errors.New("image analyse error")
			}

			// set the quadrangeles in current image (global var)
			if jsonData.ResourceName == "OnvifSnapshot" {
				// ResourceName is OnvifSnapshot
				selfQuadrangleCount = quadrangleCount
			} else {
				// ResourceName is sample-image
				otherQuadrangleCount = quadrangleCount
			}

			lc.Infof("CrowdedThreshold=%d, selfQuadrangleCount=%d, otherQuadrangleCount=%d",
				CrowdedThreshold, selfQuadrangleCount, otherQuadrangleCount)

			// check the number of quadrangles
			// If the number of quadrangles is less than the threshold, do nothing and return directly
			if selfQuadrangleCount < CrowdedThreshold || otherQuadrangleCount < CrowdedThreshold {
				return true, data
			}

			// Forward data to core-data
			lc.Infof("Send SimpleReading[profilename=camera devicename=%s resourcename=Crowded valuetype=Bool value=true] to core-data", DeviceName)

			// create simple reading and send it to core-data
			coreData := transforms.NewCoreDataSimpleReading("camera", DeviceName, "Crowded", "Bool")
			continuePipeline, _ := coreData.PushToCoreData(ctx, true)

			if !continuePipeline {
				lc.Errorf("binary type push to core-data NG")
				return false, errors.New("binary type push to core-data NG")
			}

		default:
			// Currently, only binary image processing is supported.
			// Other types are not supported.
			lc.Errorf("Unsupported ValueType")
			return false, errors.New("Unsupported ValueType")

		}

	}
	return true, data
}

func imageAnalyse(ctx interfaces.AppFunctionContext, data []byte) (bool, uint16) {
	lc := ctx.LoggingClient()

	// check parameter
	if data == nil {
		lc.Errorf("No data received")
		return false, 0
	}

	// reads an image from a buffer in memory, then convert to mat format.
	imageMat, err := gocv.IMDecode(data, gocv.IMReadColor)
	if err != nil {
		lc.Errorf("IMDecode NG. err = %s", err.Error())
		return false, 0
	}
	defer imageMat.Close()

	// There are objects that affect the analysis below and on the left side of the image
	// collected by camera, so the original image needs to be cropped before analysis.
	croppedMat := imageMat.Region(image.Rect(105, 50, imageMat.Cols()*3/4, imageMat.Rows()*2/3))
	defer croppedMat.Close()

	// change data format(Mat → GpuMat) to use cuda function
	imageGpu := cuda.NewGpuMat()
	defer imageGpu.Close()
	imageGpu.Upload(croppedMat)

	// change the color of image to gray
	gray := cuda.NewGpuMat()
	defer gray.Close()
	cuda.CvtColor(imageGpu, &gray, gocv.ColorBGRToGray)

	// use function NewGaussianFilter to blur the image
	Gaussianfilter := cuda.NewGaussianFilter(gocv.MatTypeCV8UC1, gocv.MatTypeCV8UC1, image.Pt(5, 5), 0)
	defer Gaussianfilter.Close()
	Gaussianfilter.Apply(gray, &gray)

	// applies a fixed-level threshold to each array element.
	cuda.Threshold(gray, &gray, 200, 255, gocv.ThresholdBinary)

	// change data format(GpuMat → Mat) to use gocv function
	imageMatNew := gocv.NewMat()
	defer imageMatNew.Close()
	gray.Download(&imageMatNew)

	// finds all element in a binary image.
	cnts := gocv.FindContours(imageMatNew, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	// define and initialize the quadrangle counter
	var quadrangleCount uint16 = 0
	for i := 0; i < cnts.Size(); i++ {
		// Check whether the polygon is a quadrangle.
		// If yes, add 1 to the quadrangle counter.
		peri := gocv.ArcLength(cnts.At(i), true)
		approx := gocv.ApproxPolyDP(cnts.At(i), 0.04*peri, true)
		if approx.Size() == 4 {
			quadrangleCount++
		}
	}

	return true, quadrangleCount
}
