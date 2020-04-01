package test

/**
 * @license
 * Copyright (c) 2020 Cisco and/or its affiliates.
 *
 * This software is licensed to you under the terms of the Cisco Sample
 * Code License, Version 1.0 (the "License"). You may obtain a copy of the
 * License at
 *
 *                https://developer.cisco.com/docs/licenses
 *
 * All use of the material herein must be in accordance with the terms of
 * the License. All rights not expressly granted by the License are
 * reserved. Unless required by applicable law or agreed to separately in
 * writing, software distributed under the License is distributed on an "AS
 * IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied.
 */

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/tidwall/gjson"
	"github.com/CiscoSE/gnso/integrations"
)

var (
	nsoRestconfCtl integrations.NSORestconfController
)

// Setup is the function that should be called at the start of each test
func setup() {
	nsoRestconfCtl.Username = os.Getenv("NSO_USERNAME")
	nsoRestconfCtl.Password = os.Getenv("NSO_PASSWORD")
	nsoRestconfCtl.Url = os.Getenv("NSO_URL")
}

// TestGetDevices test the retrieval of devices from NSO restconf controller
func TestGetDevices(t *testing.T) {
	// Prepare test
	setup()

	// Get devices and check for errors
	devices, err := nsoRestconfCtl.GetDevices()
	if err != nil {
		t.Errorf("Error while retrieving devices: %s", err)
		return
	}

	// Validate output is expected
	result := gjson.Get(devices, "tailf-ncs:device")

	// Per each device, print name, address, authentication group, device type and NED ID
	for _, device := range result.Array() {

		name := gjson.Get(device.Raw, "name")
		address := gjson.Get(device.Raw, "address")
		authgroup := gjson.Get(device.Raw, "authgroup")
		// deviceType and nedID are set in the next loop
		deviceType := ""
		nedId := gjson.Get(device.Raw, "")
		// Supported (by this client) types are cli, netconf or generic.
		for _, allowedType := range []string{"cli", "netconf", "generic"} {
			deviceTypeJson := gjson.Get(device.Raw, "device-type."+allowedType)
			if deviceTypeJson.Exists() {
				deviceType = allowedType
				nedId = gjson.Get(device.Raw, "device-type."+allowedType+".ned-id")
				// Found the device type and ned id, moving on
				break
			}
		}
		// Print results
		log.Println(fmt.Sprintf("Device: %s - Address: %s - Device Type: %s %s - Authgroup: %s", name, address, deviceType, nedId, authgroup))
	}
}

func TestGetConfig(t *testing.T) {
	// Prepare test
	setup()

	// Get config and check for errors
	config, err := nsoRestconfCtl.GetConfig("/tailf-ncs:devices/device=vwlc1/config")
	if err != nil {
		t.Errorf("Error while retrieving config: %s", err)
		return
	}

	// Print results
	log.Println(fmt.Sprintf("Config: %s", config))

}

func TestEditConfig(t *testing.T) {
	// Prepare test
	setup()

	// Set payload
	payload := `{
		"data": {
		  "tailf-ncs:devices": {
			"device": [
			  {
				"name": "vwlc1",
				"config": {
				  "tailf-ned-cisco-aireos:wlan": {
					"create": [
					  {
						"wlan-id": 2,
						"profile": "test-wireless-2",
						"ssid": "test-wireless-2"
					  }
					]
				  }
				}
			  }
			]
		  }
		}
	  }`
	// Edit config and check for errors
	_, err := nsoRestconfCtl.EditConfig("/", payload, http.MethodPatch)
	if err != nil {
		t.Errorf("Error while editing config: %s", err)
		return
	}

	// Print results
	log.Println("Configuration edited correctly")

}

func TestQuery(t *testing.T) {
	// Prepare test
	setup()

	// Set payload
	payload := `
	{
		"immediate-query": {
	   "foreach": "/devices/device/device-type/cli[starts-with(ned-id,'cisco-')]",
	   "select": [
		   { 
			   "label":"nedId",
		   "expression": "ned-id", "result-type": ["string"]
	   },
	   { 
		"label":"deviceName",
		   "expression": "../../name", "result-type": ["string"]
	   },
	   { 
		"label":"deviceAddress",
		   "expression": "../../address", "result-type": ["string"]
	   }
		   ],
		   "sort-by": ["../../name"]
		} 
		
	}`
	// Edit config and check for errors
	result, err := nsoRestconfCtl.Query(payload)
	if err != nil {
		t.Errorf("Error while editing config: %s", err)
		return
	}

	// Print results
	log.Println(fmt.Sprintf("Query result: %s", result))

}

func TestExecOperations(t *testing.T) {
	// Prepare test
	setup()

	// Set payload
	payload := ""
	// Edit config and check for errors
	data, err := nsoRestconfCtl.ExecOperations("/devices/fetch-ssh-host-keys", payload)
	if err != nil {
		t.Errorf("Error while executing operation : %s", err)
		return
	}

	// Print results
	log.Println(fmt.Sprintf("Operation executed correctly: %s", data))

}
