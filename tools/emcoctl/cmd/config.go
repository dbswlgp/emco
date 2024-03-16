// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package cmd

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

// Configurations exported
type EmcoConfigurations struct {
	Ingress      ControllerConfigurations
	Orchestrator ControllerConfigurations
	Clm          ControllerConfigurations
	Ncm          ControllerConfigurations
	Dcm          ControllerConfigurations
	Rsync        ControllerConfigurations
	OvnAction    ControllerConfigurations
	Gac          ControllerConfigurations
	Dtc          ControllerConfigurations
	HpaPlacement ControllerConfigurations
	Sfc          ControllerConfigurations
	SfcClient    ControllerConfigurations
	WorkflowMgr  ControllerConfigurations
	Tac          ControllerConfigurations
	CaCert       ControllerConfigurations
	Policy       ControllerConfigurations
	Migration    ControllerConfigurations
}

// ControllerConfigurations exported
type ControllerConfigurations struct {
	Port       int
	StatusPort int
	Host       string
}

const urlVersion string = "v2"
const urlPrefix string = "http://"

var Configurations EmcoConfigurations

// SetDefaultConfiguration default configuration if t
func SetDefaultConfiguration() {
	Configurations.Orchestrator.Host = "localhost"
	Configurations.Orchestrator.Port = 30415
	Configurations.Orchestrator.StatusPort = 30416
	Configurations.Clm.Host = "localhost"
	Configurations.Clm.Port = 30461
	Configurations.Ncm.Host = "localhost"
	Configurations.Ncm.Port = 9081
	Configurations.Ncm.StatusPort = 9082
	Configurations.Dcm.Host = "localhost"
	Configurations.Dcm.Port = 30477
	Configurations.Dcm.StatusPort = 30478
	Configurations.OvnAction.Host = "localhost"
	Configurations.OvnAction.Port = 9051
	Configurations.Dtc.Host = "localhost"
	Configurations.Dtc.Port = 30492
	Configurations.Gac.Host = "localhost"
	Configurations.Gac.Port = 30491
	Configurations.HpaPlacement.Host = "localhost"
	Configurations.HpaPlacement.Port = 30481
	Configurations.Sfc.Host = "localhost"
	Configurations.Sfc.Port = 9055
	Configurations.SfcClient.Host = "localhost"
	Configurations.SfcClient.Port = 9057
	Configurations.WorkflowMgr.Host = "localhost"
	Configurations.WorkflowMgr.Port = 9095
	Configurations.Tac.Host = "localhost"
	Configurations.Tac.Port = 9064
	Configurations.CaCert.Host = "localhost"
	Configurations.CaCert.Port = 9036
	Configurations.Policy.Host = "localhost"
	Configurations.Policy.Port = 9060
	Configurations.Migration.Host = "localhost"
	Configurations.Migration.Port = 9900
}

// GetIngressURL Url for Ingress
func GetIngressURL() string {
	if Configurations.Ingress.Host == "" || Configurations.Ingress.Port == 0 {
		return ""
	}
	return urlPrefix + net.JoinHostPort(Configurations.Ingress.Host, strconv.Itoa(Configurations.Ingress.Port)) + "/" + urlVersion
}

// GetOrchestratorURL Url for Orchestrator
func GetOrchestratorURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Orchestrator.Host == "" || Configurations.Orchestrator.Port == 0 {
		fmt.Println("Fatal: No Orchestrator Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.Orchestrator.Host, strconv.Itoa(Configurations.Orchestrator.Port)) + "/" + urlVersion
}

// GetClmURL Url for Clm
func GetClmURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Clm.Host == "" || Configurations.Clm.Port == 0 {
		fmt.Println("Fatal: No Clm Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.Clm.Host, strconv.Itoa(Configurations.Clm.Port)) + "/" + urlVersion
}

// GetNcmURL Url for Ncm
func GetNcmURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Ncm.Host == "" || Configurations.Ncm.Port == 0 {
		fmt.Println("Fatal: No Ncm Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.Ncm.Host, strconv.Itoa(Configurations.Ncm.Port)) + "/" + urlVersion
}

// GetDcmURL Url for Dcm
func GetDcmURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Dcm.Host == "" || Configurations.Dcm.Port == 0 {
		fmt.Println("Fatal: No Dcm Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.Dcm.Host, strconv.Itoa(Configurations.Dcm.Port)) + "/" + urlVersion
}

// GetOvnactionURL Url for Ovnaction
func GetOvnactionURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.OvnAction.Host == "" || Configurations.OvnAction.Port == 0 {
		fmt.Println("Fatal: No Ovn Action Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.OvnAction.Host, strconv.Itoa(Configurations.OvnAction.Port)) + "/" + urlVersion
}

// GetDtcURL Url for Dtc
func GetDtcURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Dtc.Host == "" || Configurations.Dtc.Port == 0 {
		fmt.Println("Fatal: No DTC Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.Dtc.Host, strconv.Itoa(Configurations.Dtc.Port)) + "/" + urlVersion
}

// GetGacURL Url for gac
func GetGacURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Gac.Host == "" || Configurations.Gac.Port == 0 {
		fmt.Println("Fatal: No GAC Action Information in Config File!")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.Gac.Host, strconv.Itoa(Configurations.Gac.Port)) + "/" + urlVersion
}

// GetHpaPlacementURL Url for Hpa Placement controller
func GetHpaPlacementURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.HpaPlacement.Host == "" || Configurations.HpaPlacement.Port == 0 {
		fmt.Println("Fatal: No HPA Placement Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.HpaPlacement.Host, strconv.Itoa(Configurations.HpaPlacement.Port)) + "/" + urlVersion
}

// GetSfcURL Url for sfc
func GetSfcURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Sfc.Host == "" || Configurations.Sfc.Port == 0 {
		fmt.Println("Fatal: No SFC Action Information in Config File!")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.Sfc.Host, strconv.Itoa(Configurations.Sfc.Port)) + "/" + urlVersion
}

// GetSfcClientURL Url for sfcclient
func GetSfcClientURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.SfcClient.Host == "" || Configurations.SfcClient.Port == 0 {
		fmt.Println("Fatal: No SFC Client Action Information in Config File!")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.SfcClient.Host, strconv.Itoa(Configurations.SfcClient.Port)) + "/" + urlVersion
}

func GetWorkflowMgrURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.WorkflowMgr.Host == "" || Configurations.WorkflowMgr.Port == 0 {
		fmt.Println("Fatal: No WorkflowMgr Information in Config File!")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.WorkflowMgr.Host, strconv.Itoa(Configurations.WorkflowMgr.Port)) + "/" + urlVersion
}

// get the URL for the temporal action controller
func GetTacURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Tac.Host == "" || Configurations.Tac.Port == 0 {
		fmt.Println("Fatal: No TAC Information in Config File!")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.Tac.Host, strconv.Itoa(Configurations.Tac.Port)) + "/" + urlVersion
}

// GetOrchestratorGrpcEndpoint gRPC endpoint for Orchestrator
func GetOrchestratorGrpcEndpoint() string {
	if Configurations.Orchestrator.Host == "" || Configurations.Orchestrator.StatusPort == 0 {
		fmt.Println("Fatal: No Orchestrator StatusPort Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return net.JoinHostPort(Configurations.Orchestrator.Host, strconv.Itoa(Configurations.Orchestrator.StatusPort))
}

// GetNcmGrpcEndpoint gRPC endpoint for ncm
func GetNcmGrpcEndpoint() string {
	if Configurations.Ncm.Host == "" || Configurations.Ncm.StatusPort == 0 {
		fmt.Println("Fatal: No NCM StatusPort Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return net.JoinHostPort(Configurations.Ncm.Host, strconv.Itoa(Configurations.Ncm.StatusPort))
}

// GetDcmGrpcEndpoint gRPC endpoint for dcm
func GetDcmGrpcEndpoint() string {
	if Configurations.Dcm.Host == "" || Configurations.Dcm.StatusPort == 0 {
		fmt.Println("Fatal: No DCM StatusPort Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return net.JoinHostPort(Configurations.Dcm.Host, strconv.Itoa(Configurations.Dcm.StatusPort))
}

// GetCaCertUrl construct the baseUrl for CaCert controller
func GetCaCertUrl() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}

	if Configurations.CaCert.Host == "" ||
		Configurations.CaCert.Port == 0 {
		fmt.Println("Fatal: No CA Cert Information in Config File!")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + Configurations.CaCert.Host + ":" + strconv.Itoa(Configurations.CaCert.Port) + "/" + urlVersion
}


// GetPolicyUrl construct the baseUrl for policy controller
func GetPolicyURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Policy.Host == "" || Configurations.Policy.Port == 0 {
		fmt.Println("Fatal: No TAC Information in Config File!")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.Policy.Host, strconv.Itoa(Configurations.Policy.Port)) + "/" + urlVersion
}


// GetMigrationURL Url for Hpa Placement controller
func GetMigrationURL() string {
	// If Ingress is available use that url
	if s := GetIngressURL(); s != "" {
		return s
	}
	if Configurations.Migration.Host == "" || Configurations.Migration.Port == 0 {
		fmt.Println("Fatal: No Migration Information in Config File")
		// Exit executing
		os.Exit(1)
	}
	return urlPrefix + net.JoinHostPort(Configurations.Migration.Host, strconv.Itoa(Configurations.Migration.Port)) + "/" + urlVersion
}
