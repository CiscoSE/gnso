package main

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
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"github.com/CiscoSE/gnso/integrations"
	service "github.com/CiscoSE/gnso/pb"
)

// Variables
var (
	nsoRestconfCtl integrations.NSORestconfController
)

// nsoService implements service protobuf interface
type nsoService struct{}

// validateToken verifies that token in request is the same as the one set in the env variable
func (n *nsoService) isTokenValid(req *service.Request) bool {
	token := os.Getenv("TOKEN")
	// Only validate if token is different than null and empty
	if token != "" {
		return token == req.Token
	}
	return true
}

func (n *nsoService) GetDevices(ctx context.Context, req *service.GetDevicesRequest) (*service.GetDevicesResponse, error) {

	// Check token
	if !n.isTokenValid(req.Request) {
		return nil, status.Errorf(codes.PermissionDenied, "invalid token")
	}
	// Get devices and check for errors
	devices, err := nsoRestconfCtl.GetDevices()
	if err != nil {
		return nil, err
	}

	// Validate output is expected
	result := gjson.Get(devices, "tailf-ncs:device")
	deviceArray := make([]*service.Device, 0)
	for _, deviceJson := range result.Array() {
		device := service.Device{
			Name:      gjson.Get(deviceJson.Raw, "name").String(),
			Address:   gjson.Get(deviceJson.Raw, "address").String(),
			Authgroup: gjson.Get(deviceJson.Raw, "authgroup").String(),
		}
		// Supported (by this rpc call) types are cli, netconf or generic.
		for _, allowedType := range []string{"cli", "netconf", "generic"} {
			deviceTypeJson := gjson.Get(deviceJson.Raw, "device-type."+allowedType)
			if deviceTypeJson.Exists() {
				device.Type = &service.DeviceType{
					NedId:   gjson.Get(deviceJson.Raw, "device-type."+allowedType+".ned-id").String(),
					NedType: allowedType,
				}
				// Found the device type and ned id, moving on
				break
			}
		}
		deviceArray = append(deviceArray, &device)
	}

	getDeviceResponse := service.GetDevicesResponse{
		Devices: deviceArray,
	}

	return &getDeviceResponse, nil
}
func (n *nsoService) ExecOperation(ctx context.Context, req *service.ExecOperationRequest) (*service.ExecOperationResponse, error) {

	// Check token
	if !n.isTokenValid(req.Request) {
		return nil, status.Errorf(codes.PermissionDenied, "invalid token")
	}
	// full path represents url with query strings
	fullPath := req.GetPath()
	if req.GetOptions() != "" {
		// if options is not null, add them as query strings
		fullPath = fullPath + "?" + req.GetOptions()
	}
	result, err := nsoRestconfCtl.ExecOperations(fullPath, req.GetJsonData())
	if err != nil {
		return nil, err
	}
	// Create base response
	response := service.Response{
		Result: result,
	}
	// Build RPC response and return
	execOperationResponse := service.ExecOperationResponse{
		Response: &response,
	}
	return &execOperationResponse, nil
}

func (n *nsoService) GetConfig(ctx context.Context, req *service.GetConfigRequest) (*service.GetConfigResponse, error) {
	// Check token
	if !n.isTokenValid(req.Request) {
		return nil, status.Errorf(codes.PermissionDenied, "invalid token")
	}
	// full path represents url with query strings
	fullPath := req.GetPath()
	if req.GetOptions() != "" {
		// if options is not null, add them as query strings
		fullPath = fullPath + "?" + req.GetOptions()
	}
	config, err := nsoRestconfCtl.GetConfig(fullPath)
	if err != nil {
		return nil, err
	}
	// Create base response
	response := service.Response{
		Result: config,
	}
	// Build RPC response and return
	getConfigResponse := service.GetConfigResponse{
		Response: &response,
	}
	return &getConfigResponse, nil

}

func (n *nsoService) Query(ctx context.Context, req *service.QueryRequest) (*service.QueryResponse, error) {
	// Check token
	if !n.isTokenValid(req.Request) {
		return nil, status.Errorf(codes.PermissionDenied, "invalid token")
	}
	// full path represents url with query strings
	result, err := nsoRestconfCtl.Query(req.GetJsonQuery())
	if err != nil {
		return nil, err
	}
	// Create base response
	response := service.Response{
		Result: result,
	}
	// Build RPC response and return
	getQueryResponse := service.QueryResponse{
		Response: &response,
	}
	return &getQueryResponse, nil

}

func (n *nsoService) EditConfig(ctx context.Context, req *service.EditConfigRequest) (*service.EditConfigResponse, error) {
	// Check token
	if !n.isTokenValid(req.Request) {
		return nil, status.Errorf(codes.PermissionDenied, "invalid token")
	}
	// full path represents url with query strings
	fullPath := req.GetPath()
	if req.GetOptions() != "" {
		// if options is not null, add them as query strings
		fullPath = fullPath + "?" + req.GetOptions()
	}
	httpMethod := ""
	switch req.GetOperationType() {
	case "merge":
		httpMethod = http.MethodPatch
		break
	case "replace":
		httpMethod = http.MethodPut
		break
	case "create":
		httpMethod = http.MethodPost
		break
	case "delete":
		httpMethod = http.MethodDelete
		break
	default:
		return nil, errors.New(fmt.Sprintf("Operation type %s not supported by this server", req.GetOperationType()))

	}
	nsoReply, err := nsoRestconfCtl.EditConfig(fullPath, req.GetJsonData(), httpMethod)
	if err != nil {
		return nil, err
	}
	// Create base response
	response := service.Response{
		Result: nsoReply,
	}
	// Build RPC response and return
	editConfigResponse := service.EditConfigResponse{
		Response: &response,
	}
	return &editConfigResponse, nil
}

func main() {
	// Read the port from env variable
	port := ":" + os.Getenv("PORT")
	// Set NSO Restconf controller attributes from env variables
	nsoRestconfCtl.Username = os.Getenv("NSO_USERNAME")
	nsoRestconfCtl.Password = os.Getenv("NSO_PASSWORD")
	nsoRestconfCtl.Url = os.Getenv("NSO_URL")

	// Create needed objects for grpc server
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	creds, err := credentials.NewServerTLSFromFile("tls/cert.pem", "tls/key.pem")
	if err != nil {
		log.Fatal(err)
	}
	opts := []grpc.ServerOption{grpc.Creds(creds)}
	s := grpc.NewServer(opts...)
	// Register our nsoService struct
	service.RegisterNSOServiceServer(s, new(nsoService))

	log.Println("Starting server on port " + port)
	// Start grpc server
	s.Serve(lis)
}
