package integrations

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
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tidwall/gjson"
)

// NSORestconfController is used as single interface for all RESTCONF calls to NSO
type NSORestconfController struct {
	Url      string
	Username string
	Password string
}

// PrepareRequest set the defaults headers/authentications for all api calls
func (r *NSORestconfController) prepareRequest(req *http.Request) *http.Request {
	// Add basic auth
	req.SetBasicAuth(r.Username, r.Password)
	// Add json headers
	req.Header.Add("Accept", "application/yang-data+json")
	req.Header.Add("Content-Type", "application/yang-data+json")
	return req
}

func (r *NSORestconfController) makeRequest(req *http.Request) (string, error) {
	client := &http.Client{}
	req = r.prepareRequest(req)

	// Make request and check for errors
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	// Read response
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", errors.New(fmt.Sprintf("Status %s returned from NSO: %s", resp.Status, string(data)))
	}

	// Unmarshall - this is only done to check if NSO returned error message
	result := gjson.Get(string(data), "errors")

	// If error key is found, raise error with message
	if result.Exists() {
		return "", errors.New(fmt.Sprintf("Error returned from NSO: %s", result))
	}
	// All good, just return the string
	return string(data), nil
}

// GetDevices returns a json formatted string including list of devices.
func (r *NSORestconfController) GetDevices() (string, error) {
	// Set restconf endpoint
	restconfEndpoint := "/data/tailf-ncs:devices/device?fields=address;name;device-type;authgroup&depth=2"

	req, err := http.NewRequest("GET", r.Url+restconfEndpoint, nil)
	if err != nil {
		return "", err
	}
	// Execute request
	data, err := r.makeRequest(req)
	if err != nil {
		return "", err
	}
	// All good, just return the string
	return data, nil
}

// GetConfig returns a json formatted string for the given URL.
func (r *NSORestconfController) GetConfig(purl string) (string, error) {
	// Set restconf endpoint
	restconfEndpoint := "/data" + purl

	req, err := http.NewRequest("GET", r.Url+restconfEndpoint, nil)
	if err != nil {
		return "", err
	}
	// Execute request
	data, err := r.makeRequest(req)
	if err != nil {
		return "", err
	}
	// All good, just return the string
	return data, nil
}

// EditConfig makes a config change in NSO according to the requested URL, payload and method.
func (r *NSORestconfController) EditConfig(purl string, payload string, httpMethod string) (string, error) {
	// Set restconf endpoint
	restconfEndpoint := "/data" + purl

	req, err := http.NewRequest(httpMethod, r.Url+restconfEndpoint, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", err
	}
	// Execute request
	data, err := r.makeRequest(req)
	if err != nil {
		return "", err
	}
	// All good, just return the string
	return data, nil
}

// Query retrieves data requested in the given payload
func (r *NSORestconfController) Query(payload string) (string, error) {
	// Set restconf endpoint
	restconfEndpoint := "/tailf/query"

	req, err := http.NewRequest(http.MethodPost, r.Url+restconfEndpoint, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", err
	}
	// Execute request
	data, err := r.makeRequest(req)
	if err != nil {
		return "", err
	}
	// All good, just return the string
	return data, nil
}

// Query retrieves data requested in the given payload
func (r *NSORestconfController) ExecOperations(purl string, payload string) (string, error) {
	// Set restconf endpoint
	restconfEndpoint := "/operations" + purl

	req, err := http.NewRequest(http.MethodPost, r.Url+restconfEndpoint, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", err
	}
	// Execute request
	data, err := r.makeRequest(req)
	if err != nil {
		return "", err
	}
	// All good, just return the string
	return data, nil
}
