// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package azurearcv2

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
)

const subscriptionURL = "https://management.azure.com/subscriptions/"
const azureCoreURL = "https://management.core.windows.net/"
const azureLoginURL = "https://login.microsoftonline.com/"

type Token struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	ExtExpiresIn string `json:"ext_expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	AccessToken  string `json:"access_token"`
}

type PropertiesFlux struct {
	Scope          string             `json:"scope"`
	Namespace      string             `json:"namespace"`
	SourceKind     string             `json:"sourceKind"`
	Suspend        bool               `json:"suspend"`
	GitRepository  RepoProperties     `json:"gitRepository"`
	Kustomizations KustomizationsUnit `json:"kustomizations"`
}

type RepoProperties struct {
	Url           string  `json:"url"`
	RepositoryRef RepoRef `json:"repositoryRef"`
}

type RepoRef struct {
	Branch string `json:"branch"`
}

type KustomizationsUnit struct {
	FirstKustomization KustomizationProperties `json:"kustomization-1"`
}

type KustomizationProperties struct {
	Path                   string `json:"path"`
	TimeoutInSeconds       int    `json:"timeoutInSeconds"`
	SyncIntervalInSeconds  int    `json:"syncIntervalInSeconds"`
	RetryIntervalInSeconds int    `json:"retryIntervalInSeconds"`
	Prune                  bool   `json:"prune"`
	Force                  bool   `json:"force"`
}

type RequestbodyFlux struct {
	Properties PropertiesFlux `json:"properties"`
}

type FluxExtension struct {
	AKSIdentityType AKSIdentityTypeProp `json:"aksIdentityType"`
	Properties      ExtensionProp       `json:"properties"`
}

type AKSIdentityTypeProp struct {
	Type string `json:"type"`
}

type ExtensionProp struct {
	ExtensionType           string `json:"extensionType"`
	AutoUpgradeMinorVersion bool   `json:"autoUpgradeMinorVersion"`
}

/*
	Function to get the access token for azure arc
	params: clientId, ClientSecret, tenantIdValue
	return: Token, error
*/
func (p *AzureArcV2Provider) getAccessToken(clientId string, clientSecret string, tenantIdValue string) (string, error) {

	client := http.Client{}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Add("client_id", clientId)
	data.Add("resource", azureCoreURL)
	data.Add("client_secret", clientSecret)

	urlPost := azureLoginURL + tenantIdValue + "/oauth2/token"

	//Rest api to get the access token
	req, err := http.NewRequest("POST", urlPost, bytes.NewBufferString(data.Encode()))
	if err != nil {
		//Handle Error
		log.Error("Couldn't create Azure Access Token request", log.Fields{"err": err, "req": req})
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	res, err := client.Do(req)
	if err != nil {
		log.Error(" Azure Access Token response error", log.Fields{"err": err, "res": res})
		return "", err
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error(" Azure Access Token response marshall error", log.Fields{"err": err, "responseData": responseData})
		return "", err
	}

	// Unmarshall the response body into json and get token value
	newToken := Token{}
	json.Unmarshal(responseData, &newToken)

	return newToken.AccessToken, nil
}

/*
	Function to create gitconfiguration of fluxv1 type in azure
	params : ctx context.Context, config interface{}
	return : error
*/
func (p *AzureArcV2Provider) ApplyConfig(ctx context.Context, config interface{}) error {

	//get accesstoken for azure
	accessToken, err := p.getAccessToken(p.clientID, p.clientSecret, p.tenantID)

	log.Info("Obtained AccessToken: ", log.Fields{"accessToken": accessToken})

	if err != nil {
		log.Error("Couldn't obtain access token", log.Fields{"err": err, "accessToken": accessToken})
		return err
	}

	//Get the Namespace
	acUtils, err := utils.NewAppContextReference(ctx, p.gitProvider.Cid)
	if err != nil {
		return nil
	}
	_, level := acUtils.GetNamespace(ctx)
	if err != nil {
		return err
	}
	var gitConfiguration, operatorScope string

	// select according to logical cloud level
	if level == "0" {
		gitConfiguration = "config-" + p.gitProvider.Cid
		operatorScope = "cluster"
	} else {
		gitConfiguration = p.gitProvider.Namespace
		operatorScope = "namespace"
	}

	gitPath := "clusters/" + p.gitProvider.Cluster + "/context/" + p.gitProvider.Cid
	gitBranch := p.gitProvider.Branch

	// Install flux extension
	_, err = p.installFluxExtension(accessToken, p.subscriptionID, p.arcResourceGroup, p.arcCluster)

	if err != nil {
		log.Error("Error in installing flux extension", log.Fields{"err": err})
		return err
	}

	// Create flux configuration
	_, err = p.createFluxConfiguration(accessToken, p.gitProvider.Url, gitConfiguration, operatorScope, p.subscriptionID,
		p.arcResourceGroup, p.arcCluster, gitBranch, gitPath, p.timeOut, p.syncInterval, p.retryInterval)

	if err != nil {
		log.Error("Error in creating flux configuration", log.Fields{"err": err})
		return err
	}

	return nil

}

/*
	Function to delete the git configuration
	params : ctx context.Context, config interface{}
	return : error
*/
func (p *AzureArcV2Provider) DeleteConfig(ctx context.Context, config interface{}) error {

	//get accesstoken for azure
	accessToken, err := p.getAccessToken(p.clientID, p.clientSecret, p.tenantID)

	if err != nil {
		log.Error("Couldn't obtain access token", log.Fields{"err": err, "accessToken": accessToken})
		return err
	}

	//Get the Namespace
	acUtils, err := utils.NewAppContextReference(ctx, p.gitProvider.Cid)
	if err != nil {
		return nil
	}
	_, level := acUtils.GetNamespace(ctx)
	if err != nil {
		return err
	}
	var gitConfiguration string

	// select according to logical cloud level
	if level == "0" {
		gitConfiguration = "config-" + p.gitProvider.Cid
	} else {
		gitConfiguration = p.gitProvider.Namespace
	}

	_, err = p.deleteFluxConfiguration(accessToken, p.subscriptionID, p.arcResourceGroup, p.arcCluster, gitConfiguration)

	if err != nil {
		log.Error("Error in deleting flux configuration", log.Fields{"err": err})
		return err
	}

	return nil
}

/*
	Function to create a FluxV2 configuration for the mentioned user repo
	params : Access Token, RepositoryUrl, Flux configuration name, scope of flux, Subscription Id
			Arc Resource group Name, Arc Cluster Name, git branch and git path, timeOut(seconds), SyncInterval(seconds)
	return : response, error
*/
func (p *AzureArcV2Provider) createFluxConfiguration(accessToken string, repositoryUrl string, gitConfiguration string, operatorScopeType string, subscriptionIdValue string, arcClusterResourceGroupName string, arcClusterName string, gitbranch string, gitpath string, timeOut int, syncInterval int, retryInterval int) (string, error) {
	// PUT request for creating git configuration
	// PUT request body
	client := http.Client{}
	if gitpath != "" {
		gitpath = "./" + gitpath
	}

	properties := RequestbodyFlux{
		PropertiesFlux{
			Scope:      operatorScopeType,
			Namespace:  gitConfiguration,
			SourceKind: "GitRepository",
			Suspend:    false,
			GitRepository: RepoProperties{
				Url: repositoryUrl,
				RepositoryRef: RepoRef{
					Branch: gitbranch}},
			Kustomizations: KustomizationsUnit{
				FirstKustomization: KustomizationProperties{
					Path:                   gitpath,
					TimeoutInSeconds:       timeOut,
					SyncIntervalInSeconds:  syncInterval,
					RetryIntervalInSeconds: retryInterval,
					Prune:                  true,
					Force:                  false}}}}

	dataProperties, err := json.Marshal(properties)
	if err != nil {
		log.Error("Error in Marshalling data for creation of flux configuration", log.Fields{"err": err})
		return "", err
	}

	urlPut := subscriptionURL + subscriptionIdValue + "/resourceGroups/" + arcClusterResourceGroupName + "/providers/Microsoft.Kubernetes/connectedClusters/" + arcClusterName + "/providers/Microsoft.KubernetesConfiguration/fluxConfigurations/" + gitConfiguration + "?api-version=2022-03-01"
	reqPut, err := http.NewRequest(http.MethodPut, urlPut, bytes.NewBuffer(dataProperties))

	if err != nil {
		log.Error("Error in creating http request for creation of flux configuration", log.Fields{"err": err})
		return "", err
	}
	// Add request header
	authorizationString := "Bearer " + accessToken
	reqPut.Header.Set("Content-Type", "application/json; charset=UTF-8")
	reqPut.Header.Add("Authorization", authorizationString)
	resPut, err := client.Do(reqPut)
	if err != nil {
		log.Error("Error in http response for creation of flux configuration", log.Fields{"err": err})
		return "", err
	}
	responseDataPut, err := ioutil.ReadAll(resPut.Body)
	if err != nil {
		log.Error("Error in reading data from http response for creation of flux configuration", log.Fields{"err": err})
		return "", err
	}
	return string(responseDataPut), nil

}

/*
	Function to install Microsoft.flux extension
	params: Access token, Subscription Id Value, Arc Cluster Resource Group Name, Arc Cluster Name
	return: response, error
*/
func (p *AzureArcV2Provider) installFluxExtension(accessToken string, subscriptionIdValue string, arcClusterResourceGroupName string, arcClusterName string) (string, error) {
	// PUT request for installing microsoft.flux extension
	// PUT request body
	client := http.Client{}
	properties := FluxExtension{AKSIdentityTypeProp{"SystemAssigned"}, ExtensionProp{"microsoft.flux", true}}
	dataProperties, err := json.Marshal(properties)
	if err != nil {
		log.Error("Error in Marshalling data for Flux extension installation", log.Fields{"err": err})
		return "", err
	}

	urlPut := subscriptionURL + subscriptionIdValue + "/resourceGroups/" + arcClusterResourceGroupName + "/providers/Microsoft.Kubernetes/connectedClusters/" + arcClusterName + "/providers/Microsoft.KubernetesConfiguration/extensions/flux?api-version=2021-09-01"
	reqPut, err := http.NewRequest(http.MethodPut, urlPut, bytes.NewBuffer(dataProperties))

	if err != nil {
		log.Error("Error in http request for Flux extension installation", log.Fields{"err": err})
		return "", err
	}
	// Add request header
	authorizationString := "Bearer " + accessToken
	reqPut.Header.Set("Content-Type", "application/json; charset=UTF-8")
	reqPut.Header.Add("Authorization", authorizationString)
	resPut, err := client.Do(reqPut)
	if err != nil {
		log.Error("Error in http response for Flux extension installation", log.Fields{"err": err})
		return "", err
	}
	responseDataPut, err := ioutil.ReadAll(resPut.Body)
	if err != nil {
		log.Error("Error in reading data from http response for Flux extension installation", log.Fields{"err": err})
		return "", err
	}

	return string(responseDataPut), nil
}

/*
	Function to Delete Flux configuration
	params : Access Token, Subscription Id, Arc Cluster ResourceName, Arc Cluster Name, Flux Configuration name
	return : Response, error

*/

func (p *AzureArcV2Provider) deleteFluxConfiguration(accessToken string, subscriptionIdValue string, arcClusterResourceGroupName string, arcClusterName string, gitConfiguration string) (string, error) {
	// Create client
	client := &http.Client{}
	// Create request
	urlDelete := subscriptionURL + subscriptionIdValue + "/resourceGroups/" + arcClusterResourceGroupName + "/providers/Microsoft.Kubernetes/connectedClusters/" + arcClusterName + "/providers/Microsoft.KubernetesConfiguration/fluxConfigurations/" + gitConfiguration + "?api-version=2021-11-01-preview"
	reqDelete, err := http.NewRequest("DELETE", urlDelete, nil)

	if err != nil {
		log.Error("Error in http request for flux configuration deletion", log.Fields{"err": err})
		return "", err
	}
	// Add request header
	authorizationString := "Bearer " + accessToken
	reqDelete.Header.Set("Content-Type", "application/json; charset=UTF-8")
	reqDelete.Header.Add("Authorization", authorizationString)
	resPut, err := client.Do(reqDelete)
	if err != nil {
		log.Error("Error in http response for deleting flux configuration", log.Fields{"err": err})
		return "", err
	}
	responseDataPut, err := ioutil.ReadAll(resPut.Body)
	if err != nil {
		log.Error("Error in reading data from http response for deleting flux configuration", log.Fields{"err": err})
		return "", err
	}

	return string(responseDataPut), nil
}
