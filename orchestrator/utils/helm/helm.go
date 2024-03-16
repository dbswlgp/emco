// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package helm

import (
	"bytes"
	ejson "encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	helmOptions "helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/validation"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"

	pkgerrors "github.com/pkg/errors"
	logger "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	utils "gitlab.com/project-emco/core/emco-base/src/orchestrator/utils"
)

//KubernetesResourceTemplate - Represents the template that is used to create a particular
//resource in Kubernetes
type KubernetesResourceTemplate struct {
	// Tracks the apiVersion and Kind of the resource
	GVK schema.GroupVersionKind
	// Path to the file that contains the resource info
	FilePath string
}

// Hook is internal container for Helm Hook Definition
type Hook struct {
	Hook release.Hook
	KRT  KubernetesResourceTemplate
}

// Custom Marshal implementation to satisfy external interface
func (h Hook) MarshalJSON() ([]byte, error) {
	return ejson.Marshal(&struct {
		Name     string              `json:"name"`
		Kind     string              `json:"kind"`
		Path     string              `json:"path"`
		Manifest string              `json:"manifest"`
		Events   []release.HookEvent `json:"events"`
	}{h.Hook.Name, h.Hook.Kind, h.Hook.Path,
		h.Hook.Manifest, h.Hook.Events})
}

// Template is the interface for all helm templating commands
// Any backend implementation will implement this interface and will
// access the functionality via this.
type Template interface {
	GenerateKubernetesArtifacts(
		chartPath string,
		valueFiles []string,
		values []string) (map[string][]string, error)
}

// TemplateClient implements the Template interface
// It will also be used to maintain any localized state
type TemplateClient struct {
	whitespaceRegex *regexp.Regexp
	kubeVersion     string
	kubeNameSpace   string
	releaseName     string
	manifestName    string
}

// NewTemplateClient returns a new instance of TemplateClient
func NewTemplateClient(k8sversion, namespace, releasename, manifestFileName string) *TemplateClient {
	return &TemplateClient{
		whitespaceRegex: regexp.MustCompile(`^\s*$`),
		// defaultKubeVersion is the default value of --kube-version flag
		kubeVersion:   k8sversion,
		kubeNameSpace: namespace,
		releaseName:   releasename,
		manifestName:  manifestFileName,
	}
}

// Combines valueFiles and values into a single values stream.
// values takes precedence over valueFiles
func (h *TemplateClient) processValues(valueFiles []string, values []string) (map[string]interface{}, error) {
	settings := cli.New()
	providers := getter.All(settings)
	options := helmOptions.Options{
		ValueFiles: valueFiles,
		Values:     values,
	}
	base, err := options.MergeValues(providers)
	if err != nil {
		return nil, err
	}
	return base, nil
}

// GenerateKubernetesArtifacts a mapping of type to fully evaluated helm template
func (h *TemplateClient) GenerateKubernetesArtifacts(inputPath string, valueFiles []string,
	values []string) ([]KubernetesResourceTemplate, []*Hook, error) {

	var outputDir, chartPath, namespace, releaseName string
	var retData []KubernetesResourceTemplate
	var hookList []*Hook

	releaseName = h.releaseName
	namespace = h.kubeNameSpace

	// verify chart path exists
	if _, err := os.Stat(inputPath); err == nil {
		if chartPath, err = filepath.Abs(inputPath); err != nil {
			return retData, hookList, err
		}
	} else {
		return retData, hookList, err
	}

	//Create a temp directory in the system temp folder
	outputDir, err := ioutil.TempDir("", "helm-tmpl-")
	if err != nil {
		return retData, hookList, pkgerrors.Wrap(err, "Got error creating temp dir")
	}
	logger.Info(":: The o/p dir:: ", logger.Fields{"OutPutDirectory ": outputDir})
	if namespace == "" {
		namespace = "default"
	}

	// get combined values and create config
	rawVals, err := h.processValues(valueFiles, values)
	if err != nil {
		return retData, hookList, err
	}

	if msgs := validation.IsDNS1123Label(releaseName); releaseName != "" && len(msgs) > 0 {
		return retData, hookList, fmt.Errorf("release name %s is not a valid DNS label: %s", releaseName, strings.Join(msgs, ";"))
	}

	// Initialize the install client
	client := action.NewInstall(&action.Configuration{})
	client.DryRun = true
	client.ClientOnly = true
	client.ReleaseName = releaseName
	client.IncludeCRDs = true
	client.APIVersions = []string{}
	client.DisableHooks = true //to ensure no duplicates in case of defined pre/post install hooks

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		logger.Error("Requested helm chart is not present", logger.Fields{"Error": err.Error()})
		return retData, hookList, err
	}

	if chartRequested.Metadata.Type != "" && chartRequested.Metadata.Type != "application" {
		return nil, hookList, fmt.Errorf(
			"chart %q has an unsupported type and is not installable: %q",
			chartRequested.Metadata.Name,
			chartRequested.Metadata.Type,
		)
	}

	if chartRequested.Metadata.Deprecated {
		logger.Warn("This helm chart is deprecated", logger.Fields{})
	}

	client.Namespace = namespace
	releese, err := client.Run(chartRequested, rawVals)
	if err != nil {
		logger.Error("Error in processing the helm chart", logger.Fields{"Error": err.Error()})
		// Unwrap error returned by Helm library to have the complete cause of the error
		errString := pkgerrors.Unwrap(err).Error()
		return nil, hookList, pkgerrors.Errorf("%s", errString)
	}
	// SplitManifests returns integer-sortable so that manifests get output
	// in the same order as the input by `BySplitManifestsOrder`.
	rmap := releaseutil.SplitManifests(releese.Manifest)
	keys := make([]string, 0, len(rmap))
	for k := range rmap {
		keys = append(keys, k)
	}
	// Sort Keys to get Sort Order
	sort.Sort(releaseutil.BySplitManifestsOrder(keys))
	for _, k := range keys {
		data := rmap[k]
		b := filepath.Base(k)
		if b == "NOTES.txt" {
			continue
		}
		if strings.HasPrefix(b, "_") {
			continue
		}
		// blank template after execution
		if h.whitespaceRegex.MatchString(data) {
			continue
		}
		mfilePath := filepath.Join(outputDir, k)
		utils.EnsureDirectory(mfilePath)
		err = ioutil.WriteFile(mfilePath, []byte(data), 0600)
		if err != nil {
			return retData, hookList, err
		}
		gvk, err := getGroupVersionKind(data)
		if err != nil {
			return retData, hookList, err
		}
		kres := KubernetesResourceTemplate{
			GVK:      gvk,
			FilePath: mfilePath,
		}
		retData = append(retData, kres)
	}
	// Handle Hooks
	sort.Stable(hookByWeight(releese.Hooks))
	for i, h := range releese.Hooks {
		testhook := false
		for _, e := range h.Events {
			if e == release.HookTest {
				testhook = true
			}
		}
		// if hook is only a test hook, then skip it
		if testhook && len(h.Events) == 1 {
			continue
		}
		hFilePath := filepath.Join(outputDir, "hook-"+fmt.Sprint(i))
		utils.EnsureDirectory(hFilePath)
		err = ioutil.WriteFile(hFilePath, []byte(h.Manifest), 0600)
		if err != nil {
			return retData, hookList, err
		}
		gvk, err := getGroupVersionKind(h.Manifest)
		if err != nil {
			return retData, hookList, err
		}
		hookList = append(hookList, &Hook{*h, KubernetesResourceTemplate{gvk, hFilePath}})
	}
	return retData, hookList, nil
}

func getGroupVersionKind(data string) (schema.GroupVersionKind, error) {
	out, err := k8syaml.ToJSON([]byte(data))
	if err != nil {
		return schema.GroupVersionKind{}, pkgerrors.Wrap(err, "Converting yaml to json")
	}

	simpleMeta := json.SimpleMetaFactory{}
	gvk, err := simpleMeta.Interpret(out)
	if err != nil {
		return schema.GroupVersionKind{}, pkgerrors.Wrap(err, "Parsing apiversion and kind")
	}

	return *gvk, nil
}

// Resolver is an interface exposes the helm related functionalities
type Resolver interface {
	Resolve(appContent, appProfileContent []byte, overrideValuesOfAppStr []string, appName string) ([]KubernetesResourceTemplate, error)
}

func cleanupTempFiles(fp string) error {
	sa := strings.Split(fp, "/")
	dp := "/" + sa[1] + "/" + sa[2] + "/"
	err := os.RemoveAll(dp)
	if err != nil {
		logger.Error("Error while deleting dir", logger.Fields{"Dir: ": dp})
		return err
	}
	logger.Info("Clean up k8s-ext-tmp-dir::", logger.Fields{"Dir: ": dp})
	return nil
}

// Resolve function
func (h *TemplateClient) Resolve(appContent []byte, appProfileContent []byte, overrideValuesOfAppStr []string, appName string) ([]KubernetesResourceTemplate, []*Hook, error) {

	var sortedTemplates []KubernetesResourceTemplate
	var hookList []*Hook
	//chartBasePath is the tmp path where the appContent(rawHelmCharts) is extracted.
	chartBasePath, err := utils.ExtractTarBall(bytes.NewBuffer(appContent))
	defer cleanupTempFiles(chartBasePath)
	if err != nil {
		logger.Error("Error while extracting appContent", logger.Fields{})
		return sortedTemplates, hookList, pkgerrors.Wrap(err, "Error while extracting appContent")
	}
	logger.Info("The chartBasePath ::", logger.Fields{"chartBasePath": chartBasePath})

	//prPath is the tmp path where the appProfileContent is extracted.
	prPath, err := utils.ExtractTarBall(bytes.NewBuffer(appProfileContent))
	defer cleanupTempFiles(prPath)
	if err != nil {
		logger.Error("Error while extracting Profile Content", logger.Fields{})
		return sortedTemplates, hookList, pkgerrors.Wrap(err, "Error while extracting Profile Content")
	}
	logger.Info("The profile path:: ", logger.Fields{"Profile Path": prPath})

	prYamlClient, err := ProcessProfileYaml(prPath, h.manifestName)
	if err != nil {
		logger.Error("Error while processing Profile Manifest", logger.Fields{})
		return sortedTemplates, hookList, pkgerrors.Wrap(err, "Error while processing Profile Manifest")
	}
	logger.Info("Got the profileYamlClient..", logger.Fields{})

	err = prYamlClient.CopyConfigurationOverrides(chartBasePath)
	if err != nil {
		logger.Error("Error while copying configresources to chart", logger.Fields{})
		return sortedTemplates, hookList, pkgerrors.Wrap(err, "Error while copying configresources to chart")
	}

	chartPath := filepath.Join(chartBasePath, appName)
	sortedTemplates, hookList, err = h.GenerateKubernetesArtifacts(chartPath, []string{prYamlClient.GetValues()}, overrideValuesOfAppStr)
	if err != nil {
		logger.Error("Error while generating final k8s yaml", logger.Fields{})
		return sortedTemplates, hookList, pkgerrors.Wrap(err, "Error while generating final k8s yaml")
	}
	return sortedTemplates, hookList, nil
}

func GetHooksByEvent(hs []*Hook) (map[string][]*Hook, error) {
	resources := make(map[string][]*Hook)
	for _, h := range hs {
		for _, e := range h.Hook.Events {
			resources[e.String()] = append(resources[e.String()], h)
		}
	}
	return resources, nil
}

// Copied from https://github.com/helm/helm/blob/a499b4b179307c267bdf3ec49b880e3dbd2a5591/pkg/action/hooks.go#L110

type hookByWeight []*release.Hook

func (x hookByWeight) Len() int      { return len(x) }
func (x hookByWeight) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
func (x hookByWeight) Less(i, j int) bool {
	if x[i].Weight == x[j].Weight {
		return x[i].Name < x[j].Name
	}
	return x[i].Weight < x[j].Weight
}
