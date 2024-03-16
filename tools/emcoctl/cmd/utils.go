// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"os"
	"strings"
	"time"

	"text/template"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	pkgerrors "github.com/pkg/errors"
	statusnotifypb "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/statusnotify"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

var inputFiles []string
var valuesFiles []string
var token []string
var stopFlag bool
var acceptWaitTime int

type ResourceContext struct {
	Anchor string `json:"anchor" yaml:"anchor"`
}

type Metadata struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	UserData1   string `yaml:"userData1,omitempty" json:"userData1,omitempty"`
	UserData2   string `yaml:"userData2,omitempty" json:"userData2,omitempty"`
}

type emcoRes struct {
	Version string                 `yaml:"version" json:"version"`
	Context ResourceContext        `yaml:"resourceContext" json:"resourceContext"`
	Meta    Metadata               `yaml:"metadata" json:"metadata"`
	Spec    map[string]interface{} `yaml:"spec,omitempty" json:"spec,omitempty"`
	File    string                 `yaml:"file,omitempty" json:"file,omitempty"`
	Files   []string               `yaml:"files,omitempty" json:"files,omitempty"`
	Label   string                 `yaml:"clusterLabel,omitempty" json:"clusterLabel,omitempty"`


	Status  map[string]interface{} `yaml:"status,omitempty" json:"status,omitempty"`
}

type emcoBody struct {
	Meta   Metadata               `json:"metadata,omitempty"`
	Label  string                 `json:"clusterLabel,omitempty"`
	Spec   map[string]interface{} `json:"spec,omitempty"`


	Status map[string]interface{} `json:"status,omitempty"`
}

type emcoCompositeAppSpec struct {
	CompositeAppVersion string `json: "compositeAppVersion"`
}

type Resources struct {
	anchor string
	body   []byte
	file   string
	files  []string
}

// GrpcClient to use with CLI
type GrpcClient struct {
	client *grpc.ClientConn
}

// CreateGrpcClient creates the gRPC Client Connection
func newGrpcClient(endpoint string) (*grpc.ClientConn, error) {
	var err error
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		fmt.Printf("Grpc Client Initialization failed with error: %v\n", err)
	}

	return conn, err
}

// RestyClient to use with CLI
type RestyClient struct {
	client *resty.Client
}

var Client RestyClient

// NewRestClient returns a rest client
func NewRestClient() RestyClient {
	// Create a Resty Client
	Client.client = resty.New()
	// Registering global Error object structure for JSON/XML request
	//Client.client.SetError(&Error{})
	return Client
}

// NewRestClientToken returns a rest client with token
func NewRestClientToken(token string) RestyClient {
	// Create a Resty Client
	Client.client = resty.New()
	// Bearer Auth Token for all request
	Client.client.SetAuthToken(token)
	// Registering global Error object structure for JSON/XML request
	//Client.client.SetError(&Error{})
	return Client
}

// readResources reads all the resources in the file provided
func readResources() []Resources {
	// TODO: Remove Assumption only one file
	// Open file and Parse to get all resources
	var resources []Resources
	var buf bytes.Buffer

	if len(valuesFiles) > 0 {
		//Apply template
		v, err := os.Open(valuesFiles[0])
		defer v.Close()
		if err != nil {
			fmt.Println("Error reading file", "error", err, "filename", valuesFiles[0])
			return []Resources{}
		}
		valDec := yaml.NewDecoder(v)
		var mapDoc interface{}
		if valDec.Decode(&mapDoc) != nil {
			fmt.Println("Values file format incorrect:", "error", err, "filename", valuesFiles[0])
			return []Resources{}
		}
		// Templatize
		t, err := template.ParseFiles(inputFiles[0])
		if err != nil {
			fmt.Println("Error reading file", "error", err, "filename", inputFiles[0])
			return []Resources{}
		}
		err = t.Execute(&buf, mapDoc)
		if err != nil {
			fmt.Println("execute: ", err)
			return []Resources{}
		}
	} else {
		f, err := os.Open(inputFiles[0])
		defer f.Close()
		if err != nil {
			fmt.Println("Error reading file", "error", err, "filename", inputFiles[0])
			return []Resources{}
		}
		io.Copy(&buf, f)
	}

	dec := yaml.NewDecoder(&buf)
	// Iterate through all resources in the file
	for {
		var doc emcoRes
		if err := dec.Decode(&doc); err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Invalid input Yaml! Exiting..", err)
				// Exit executing
				os.Exit(1)
			}
			break
		}
		body := &emcoBody{Meta: doc.Meta, Spec: doc.Spec, Status: doc.Status, Label: doc.Label}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			fmt.Println("Invalid input Yaml! Exiting..", err)
			// Exit executing
			os.Exit(1)
		}
		var res Resources
		if doc.File != "" {
			res = Resources{anchor: doc.Context.Anchor, body: jsonBody, file: doc.File}
		} else if len(doc.Files) > 0 {
			res = Resources{anchor: doc.Context.Anchor, body: jsonBody, files: doc.Files}
		} else {
			res = Resources{anchor: doc.Context.Anchor, body: jsonBody}
		}
		resources = append(resources, res)
	}
	return resources
}

//RestClientApply to post to server no multipart
func (r RestyClient) RestClientApply(anchor string, body []byte, put bool) error {
	var resp *resty.Response
	var err error
	var url string

	if put {
		if anchor, err = getUpdateUrl(anchor, body); err != nil {
			return err
		}
		if url, err = GetURL(anchor); err != nil {
			return err
		}
		// Put JSON string
		resp, err = r.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			Put(url)
	} else {
		if url, err = GetURL(anchor); err != nil {
			return err
		}
		// Post JSON string
		resp, err = r.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			Post(url)
	}
	if err != nil {
		fmt.Println(err)
		return err
	}
	if put {
		printOutput(url, "PUT", resp)
	} else {
		printOutput(url, "POST", resp)
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() <= 299 {
		// Wait for some time for the accepted call to finish
		if resp.StatusCode() == http.StatusAccepted {
			fmt.Println("API Response code 202. Waiting...")
			time.Sleep(time.Duration(acceptWaitTime) * time.Second)
		}
		return nil
	}
	return pkgerrors.Errorf("API Error")
}

//RestClientPut to post to server no multipart
func (r RestyClient) RestClientPut(anchor string, body []byte) error {
	if anchor == "" {
		return pkgerrors.Errorf("Anchor can't be empty")
	}
	s := strings.Split(anchor, "/")
	a := s[len(s)-1]
	if a == "instantiate" || a == "apply" || a == "approve" || a == "terminate" || a == "migrate" || a == "update" || a == "rollback" {
		// No put for these
		return nil
	}
	return r.RestClientApply(anchor, body, true)
}

//RestClientPost to post to server no multipart
func (r RestyClient) RestClientPost(anchor string, body []byte) error {
	return r.RestClientApply(anchor, body, false)
}

//RestClientMultipartApply to post to server with multipart
func (r RestyClient) RestClientMultipartApply(anchor string, body []byte, file string, put bool) error {
	var resp *resty.Response
	var err error
	var url string

	// Read file for multipart
	f, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading file", "error", err, "filename", file)
		return err
	}

	// Multipart Post
	formParams := neturl.Values{}
	formParams.Add("metadata", string(body))
	if put {
		if anchor, err = getUpdateUrl(anchor, body); err != nil {
			return err
		}
		if url, err = GetURL(anchor); err != nil {
			return err
		}
		resp, err = r.client.R().
			SetFileReader("file", "filename", bytes.NewReader(f)).
			SetFormDataFromValues(formParams).
			Put(url)
	} else {
		if url, err = GetURL(anchor); err != nil {
			return err
		}
		resp, err = r.client.R().
			SetFileReader("file", "filename", bytes.NewReader(f)).
			SetFormDataFromValues(formParams).
			Post(url)
	}
	if err != nil {
		fmt.Println(err)
		return err
	}
	if put {
		printOutput(url, "PUT", resp)
	} else {
		printOutput(url, "POST", resp)
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() <= 299 {
		if resp.StatusCode() == http.StatusAccepted {
			// Wait some time for the accepted call to finish
			fmt.Println("API Response code 202. Waiting...")
			time.Sleep(time.Duration(acceptWaitTime) * time.Second)

		}
		return nil
	}
	return pkgerrors.Errorf("API Error")
}

//RestClientMultipartPut to post to server with multipart
func (r RestyClient) RestClientMultipartPut(anchor string, body []byte, file string) error {
	return r.RestClientMultipartApply(anchor, body, file, true)
}

//RestClientMultipartPost to post to server with multipart
func (r RestyClient) RestClientMultipartPost(anchor string, body []byte, file string) error {
	return r.RestClientMultipartApply(anchor, body, file, false)
}

func getFile(file string) ([]byte, string, error) {
	// Read file for multipart
	f, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading file", "error", err, "filename", file)
		return []byte{}, "", err
	}
	// Extract filename
	s := strings.TrimSuffix(file, "/")
	s1 := strings.Split(s, "/")
	name := s1[len(s1)-1]
	return f, name, nil
}

//RestClientMultipartApplyMultipleFiles to post to server with multipart
func (r RestyClient) RestClientMultipartApplyMultipleFiles(anchor string, body []byte, files []string, put bool) error {
	var f []byte
	var name string
	var err error
	var url string
	var resp *resty.Response

	req := r.client.R()
	// Multipart Post
	formParams := neturl.Values{}
	formParams.Add("metadata", string(body))
	// Add all files in the list
	for _, file := range files {
		f, name, err = getFile(file)
		if err != nil {
			return err
		}
		req = req.
			SetFileReader("files", name, bytes.NewReader(f))
	}
	if put {
		if anchor, err = getUpdateUrl(anchor, body); err != nil {
			return err
		}
		if url, err = GetURL(anchor); err != nil {
			return err
		}
		resp, err = req.
			SetFormDataFromValues(formParams).
			Put(url)
	} else {
		if url, err = GetURL(anchor); err != nil {
			return err
		}
		resp, err = req.
			SetFormDataFromValues(formParams).
			Post(url)
	}
	if err != nil {
		fmt.Println(err)
		return err
	}
	if put {
		printOutput(url, "PUT", resp)
	} else {
		printOutput(url, "POST", resp)
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() <= 299 {
		if resp.StatusCode() == http.StatusAccepted {
			// Wait some time for the accepted call to finish
			fmt.Println("API Response code 202. Waiting...")
			time.Sleep(time.Duration(acceptWaitTime) * time.Second)

		}
		return nil
	}
	return pkgerrors.Errorf("API Error")
}

//RestClientMultipartPutMultipleFiles to post to server with multipart
func (r RestyClient) RestClientMultipartPutMultipleFiles(anchor string, body []byte, files []string) error {
	return r.RestClientMultipartApplyMultipleFiles(anchor, body, files, true)
}

//RestClientMultipartPostMultipleFiles to post to server with multipart
func (r RestyClient) RestClientMultipartPostMultipleFiles(anchor string, body []byte, files []string) error {
	return r.RestClientMultipartApplyMultipleFiles(anchor, body, files, false)
}

// RestClientGetAnchor returns get data from anchor
func (r RestyClient) RestClientGetAnchor(anchor string) error {
	url, err := GetURL(anchor)
	if err != nil {
		return err
	}
	s := strings.Split(anchor, "/")
	if len(s) >= 3 {
		a := s[len(s)-2]
		// Determine if multipart
		if a == "apps" || a == "profiles" || a == "clusters" || a == "resources" || a == "customizations" {
			// Supports only getting metadata
			resp, err := r.client.R().
				SetHeader("Accept", "application/json").
				Get(url)
			if err != nil {
				fmt.Println(err)
				return err
			}
			printOutput(url, "GET", resp)
			return nil
		}
	}
	resp, err := r.client.R().
		Get(url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	printOutput(url, "GET", resp)
	return nil
}

func getUpdateUrl(anchor string, body []byte) (string, error) {
	var e emcoBody
	err := json.Unmarshal(body, &e)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	s := strings.Split(anchor, "/")
	a := s[len(s)-1]
	if e.Meta.Name != "" {
		name := e.Meta.Name
		anchor = anchor + "/" + name
		if a == "composite-apps" {
			var cav emcoCompositeAppSpec
			err := mapstructure.Decode(e.Spec, &cav)
			if err != nil {
				fmt.Println("Unable to decode CompositeApp Spec")
				return "", err
			}
			anchor = anchor + "/" + cav.CompositeAppVersion
		}
	} else if e.Label != "" {
		anchor = anchor + "/" + e.Label
	}
	return anchor, nil
}

// RestClientGet gets resource
func (r RestyClient) RestClientGet(anchor string, body []byte) error {
	if anchor == "" {
		return pkgerrors.Errorf("Anchor can't be empty")
	}
	s := strings.Split(anchor, "/")
	a := s[len(s)-1]
	if a == "instantiate" || a == "apply" || a == "approve" || a == "terminate" || a == "migrate" || a == "update" || a == "rollback" || a == "stop" {
		// No get for these
		return nil
	}
	c, err := getUpdateUrl(anchor, body)
	if err != nil {
		return err
	}
	return r.RestClientGetAnchor(c)
}

// RestClientDeleteAnchor returns all resource in the input file
func (r RestyClient) RestClientDeleteAnchor(anchor string) error {
	url, err := GetURL(anchor)
	if err != nil {
		return err
	}
	resp, err := r.client.R().Delete(url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	printOutput(url, "DELETE", resp)
	if resp.StatusCode() >= 200 && resp.StatusCode() <= 299 {
		return nil
	} else {
		return pkgerrors.Errorf("API Error")
	}
}

// RestClientDelete calls rest delete command
func (r RestyClient) RestClientDelete(anchor string, body []byte) error {

	s := strings.Split(anchor, "/")
	a := s[len(s)-1]
	if a == "instantiate" {
		// Change instantiate to destroy
		s[len(s)-1] = "terminate"
		anchor = strings.Join(s[:], "/")
		return r.RestClientPost(anchor, []byte{})
	} else if a == "apply" {
		// Change apply to terminate
		s[len(s)-1] = "terminate"
		anchor = strings.Join(s[:], "/")
		return r.RestClientPost(anchor, []byte{})
	} else if a == "approve" || a == "status" || a == "migrate" || a == "update" || a == "rollback" {
		// No delete required for these
		return nil
	}
	var e emcoBody
	err := json.Unmarshal(body, &e)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if e.Meta.Name != "" {
		s := strings.Split(anchor, "/")
		a := s[len(s)-1]
		name := e.Meta.Name
		anchor = anchor + "/" + name
		if a == "composite-apps" {
			var cav emcoCompositeAppSpec
			err := mapstructure.Decode(e.Spec, &cav)
			if err != nil {
				fmt.Println("Unable to decode CompositeApp Spec")
				return err
			}
			anchor = anchor + "/" + cav.CompositeAppVersion
		}
	} else if e.Label != "" {
		anchor = anchor + "/" + e.Label
	}
	return r.RestClientDeleteAnchor(anchor)
}

// GetURL reads the configuration file to get URL
func GetURL(anchor string) (string, error) {
	var baseUrl string
	s := strings.Split(anchor, "/")
	if len(s) < 1 {
		return "", fmt.Errorf("Invalid Anchor: %s", s)
	}

	switch s[0] {
	case "cluster-providers":
		if len(s) >= 5 && (s[4] == "networks" || s[4] == "provider-networks" ||
			s[4] == "apply" || s[4] == "terminate" || strings.HasPrefix(s[4], "status") || s[4] == "stop") {
			baseUrl = GetNcmURL()
			break
		}
		if len(s) >= 3 && s[2] == "ca-certs" {
			baseUrl = GetCaCertUrl()
			break
		}
		baseUrl = GetClmURL()
	case "controllers":
		baseUrl = GetOrchestratorURL()
	case "clm-controllers":
		baseUrl = GetClmURL()
	case "dtc-controllers":
		baseUrl = GetDtcURL()
	case "projects":
		if len(s) >= 3 && s[2] == "logical-clouds" {
			baseUrl = GetDcmURL()
			break
		}
		if len(s) >= 3 && s[2] == "ca-certs" {
			baseUrl = GetCaCertUrl()
			break
		}
		if len(s) >= 3 && s[2] == "policy" {
			baseUrl = GetPolicyURL()
			break
		}
		if len(s) >= 8 && s[7] == "network-chains" {
			baseUrl = GetSfcURL()
			break
		}
		if len(s) >= 8 && s[7] == "sfc-clients" {
			baseUrl = GetSfcClientURL()
			break
		}
		if len(s) >= 8 && s[7] == "network-controller-intent" {
			baseUrl = GetOvnactionURL()
			break
		}
		if len(s) >= 8 && s[7] == "traffic-group-intents" {
			baseUrl = GetDtcURL()
			break
		}
		if len(s) >= 8 && s[7] == "generic-k8s-intents" {
			baseUrl = GetGacURL()
			break
		}
		if len(s) >= 8 && s[7] == "hpa-intents" {
			baseUrl = GetHpaPlacementURL()
			break
		}
		if len(s) >= 8 && s[7] == "temporal-workflow-intents" {
			baseUrl = GetWorkflowMgrURL()
			break
		}
		if len(s) >= 8 && s[7] == "temporal-action-controller" {
			baseUrl = GetTacURL()
			break
		}
		if len(s) >= 8 && s[7] == "policy-intents" {
			baseUrl = GetPolicyURL()
			break
		}
		if (len(s) >= 8 && s[7] == "migration-intents") || s[2] == "migration-policies" {
			baseUrl = GetMigrationURL()
			break
		}
		// All other paths go to Orchestrator
		baseUrl = GetOrchestratorURL()
	default:
		return "", fmt.Errorf("Invalid baseUrl: %s", baseUrl)
	}
	return (baseUrl + "/" + anchor), nil
}

// WatchGrpcEndpoint reads the configuration file to get gRPC Endpoint
// and makes a connection to watch status notifications.
func WatchGrpcEndpoint(args ...string) {
	var endpoint string
	var anchor string
	var reg statusnotifypb.StatusRegistration

	reg.Output = statusnotifypb.OutputType_SUMMARY
	reg.StatusType = statusnotifypb.StatusValue_READY
	reg.Apps = make([]string, 0)
	reg.Clusters = make([]string, 0)
	reg.Resources = make([]string, 0)

	for i, arg := range args {
		if i == 0 {
			anchor = arg
			continue
		}
		s := strings.Split(arg, "=")
		if len(s) != 2 {
			fmt.Errorf("Invalid argument: %s\n", s)
			fmt.Println("Use: 'emcoctl watch --help'")
			return
		}
		switch s[0] {
		case "format":
			if s[1] == "summary" {
				reg.Output = statusnotifypb.OutputType_SUMMARY
			} else if s[1] == "all" {
				reg.Output = statusnotifypb.OutputType_ALL
			} else {
				fmt.Errorf("Invalid output format parameter: %s\n", s[1])
				fmt.Println("Use: 'emcoctl watch --help'")
				return
			}
		case "status":
			if s[1] == "deployed" {
				reg.StatusType = statusnotifypb.StatusValue_DEPLOYED
			} else if s[1] == "ready" {
				reg.StatusType = statusnotifypb.StatusValue_READY
			} else {
				fmt.Errorf("Invalid output format parameter: %s\n", s[1])
				fmt.Println("Use: 'emcoctl watch --help'")
				return
			}
		case "app":
			reg.Apps = append(reg.Apps, s[1])
		case "cluster":
			reg.Clusters = append(reg.Clusters, s[1])
		case "resource":
			reg.Resources = append(reg.Resources, s[1])
		}
	}

	s := strings.Split(anchor, "/")
	if len(s) < 1 {
		fmt.Errorf("Invalid Anchor: %s\n", s)
		fmt.Println("Use: 'emcoctl watch --help'")
		return
	}

	switch s[0] {
	case "cluster-providers":
		if len(s) == 5 && s[2] == "clusters" && s[4] == "status" {
			endpoint = GetNcmGrpcEndpoint()
			reg.Key = &statusnotifypb.StatusRegistration_ClusterKey{
				ClusterKey: &statusnotifypb.ClusterKey{
					ClusterProvider: s[1],
					Cluster:         s[3],
				},
			}
			break
		}
		fmt.Errorf("Invalid Anchor: %s\n", s)
		fmt.Println("Use: 'emcoctl watch --help'")
		return
	case "projects":
		if len(s) == 5 && s[2] == "logical-clouds" && s[4] == "status" {
			endpoint = GetDcmGrpcEndpoint()
			reg.Key = &statusnotifypb.StatusRegistration_LcKey{
				LcKey: &statusnotifypb.LcKey{
					Project:      s[1],
					LogicalCloud: s[3],
				},
			}
			break
		}
		if len(s) == 8 && s[2] == "composite-apps" && s[5] == "deployment-intent-groups" && s[7] == "status" {
			endpoint = GetOrchestratorGrpcEndpoint()
			reg.Key = &statusnotifypb.StatusRegistration_DigKey{
				DigKey: &statusnotifypb.DigKey{
					Project:               s[1],
					CompositeApp:          s[3],
					CompositeAppVersion:   s[4],
					DeploymentIntentGroup: s[6],
				},
			}
			break
		}
		fmt.Errorf("Invalid status anchor: %s\n", s)
		fmt.Println("Use: 'emcoctl watch --help'")
		return
	default:
		fmt.Errorf("Invalid status anchor: %s\n", s)
		fmt.Println("Use: 'emcoctl watch --help'")
		return
	}

	reg.ClientId = uuid.New().String()

	conn, err := newGrpcClient(endpoint)
	if err != nil {
		fmt.Errorf("Error connecting to gRPC status endpoint: %s, Error: %s\n", endpoint, err)
		fmt.Println("Use: 'emcoctl watch --help'")
		return
	}

	client := statusnotifypb.NewStatusNotifyClient(conn)

	stream, err := client.StatusRegister(context.Background(), &reg, grpc.WaitForReady(true))
	if err != nil {
		fmt.Errorf("Error registering for status notifications, gRPC status endpoint: %s, Error: %s\n", endpoint, err)
		fmt.Println("Use: 'emcoctl watch --help'")
		return
	}

	for true {
		resp, err := stream.Recv()
		if err != nil {
			fmt.Errorf("error reading from status stream: %s\n", err)
			time.Sleep(5 * time.Second) // protect against potential deluge of errors in the for loop
			continue
		}
		printResponse(resp)
	}
}

func printResponse(resp *statusnotifypb.StatusNotification) {
	jsonStatus, err := protojson.Marshal(resp)
	if err != nil {
		fmt.Println("Error Marshalling Status Notification to JSON:", err)
		return
	}
	fmt.Printf("%v\n", string(jsonStatus))
}

func printOutput(url, op string, resp *resty.Response) {
	fmt.Println("---")
	fmt.Println(op, " --> URL:", url)
	fmt.Println("Response Code:", resp.StatusCode())
	if len(resp.Body()) > 0 {
		fmt.Println("Response:", resp)
	}
}

func HandleError(err error, op string, anchor string) bool {
	if err != nil {
		if err.Error() == "API Error" {
			// On API Error stop processing if stopFlag is set
			return stopFlag
		} else {
			// On any other error than API error stop processing
			fmt.Println(op, anchor, "Error: ", err)
			return true
		}
	} else {
		// Not an error
		return false
	}
}
