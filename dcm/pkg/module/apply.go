// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2022 Intel Corporation

package module

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"
	"strings"

	"context"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/common"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext/subresources"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/installappclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/updateappclient"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
	rsync "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	"gopkg.in/yaml.v2"
	k8scertsv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// rsyncName denotes the name of the rsync controller
const rsyncName = "rsync"

// lcAppName denotes the technical/internal name of *any* logical cloud inside an appcontext
const lcAppName = "logical-cloud"

type Resource struct {
	ApiVersion    string         `yaml:"apiVersion"`
	Kind          string         `yaml:"kind"`
	MetaData      MetaDatas      `yaml:"metadata"`
	Specification Specs          `yaml:"spec,omitempty"`
	Rules         []RoleRules    `yaml:"rules,omitempty"`
	Subjects      []RoleSubjects `yaml:"subjects,omitempty"`
	RoleRefs      RoleRef        `yaml:"roleRef,omitempty"`
}

type MetaDatas struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace,omitempty"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

type Specs struct {
	Request    string   `yaml:"request,omitempty"`    // for CSR
	Usages     []string `yaml:"usages,omitempty"`     // for CSR
	SignerName string   `yaml:"signerName,omitempty"` // for CSR
	// TODO: validate quota keys
	// //Hard           logicalcloud.QSpec    `yaml:"hard,omitempty"`
	// Hard QSpec `yaml:"hard,omitempty"`
	Hard map[string]string `yaml:"hard,omitempty"` // for Quotas
	Git  AnthosGit         `yaml:"git,omitempty"`  // for Anthos RepoSync
}

type RoleRules struct {
	ApiGroups []string `yaml:"apiGroups"`
	Resources []string `yaml:"resources"`
	Verbs     []string `yaml:"verbs"`
}

type RoleSubjects struct {
	Kind      string `yaml:"kind"`
	Name      string `yaml:"name"`
	ApiGroup  string `yaml:"apiGroup"`
	NameSpace string `yaml:"namespace,omitempty"`
}

type RoleRef struct {
	Kind     string `yaml:"kind"`
	Name     string `yaml:"name"`
	ApiGroup string `yaml:"apiGroup"`
}

type AnthosGit struct {
	Repo        string `yaml:"repo"`
	Revision    string `yaml:"revision"`
	Branch      string `yaml:"branch"`
	Dir         string `yaml:"dir"`
	Auth        string `yaml:"auth"`
	NoSSLVerify bool   `yaml:"noSSLVerify"`
}

func cleanupCompositeApp(ctx context.Context, appCtx appcontext.AppContext, err error, reason string, details []string) error {
	if err == nil {
		// create an error object to avoid wrap failures
		err = pkgerrors.New("Composite App cleanup.")
	}

	cleanuperr := appCtx.DeleteCompositeApp(ctx)
	newerr := pkgerrors.Wrap(err, reason)
	if cleanuperr != nil {
		log.Warn("Error cleaning AppContext, ", log.Fields{
			"Related details": details,
		})
		// this would be useful: https://godoc.org/go.uber.org/multierr
		return pkgerrors.Wrap(err, "After previous error, cleaning the AppContext also failed.")
	}
	return newerr
}

func createNamespace(logicalcloud common.LogicalCloud) (string, string, error) {

	name := logicalcloud.Specification.NameSpace
	labels := logicalcloud.Specification.Labels

	namespace := Resource{
		ApiVersion: "v1",
		Kind:       "Namespace",
		MetaData: MetaDatas{
			Name:   name,
			Labels: labels,
		},
	}

	nsData, err := yaml.Marshal(&namespace)
	if err != nil {
		return "", "", err
	}

	return string(nsData), strings.Join([]string{name, "+Namespace"}, ""), nil
}

func createServiceAccount(logicalcloud common.LogicalCloud) (string, string, error) {

	name := logicalcloud.Specification.NameSpace
	labels := logicalcloud.Specification.Labels

	sa := Resource{
		ApiVersion: "v1",
		Kind:       "ServiceAccount",
		MetaData: MetaDatas{
			Name:      strings.Join([]string{name, "-sa"}, ""),
			Labels:    labels,
			Namespace: name,
		},
	}

	nsData, err := yaml.Marshal(&sa)
	if err != nil {
		return "", "", err
	}

	return string(nsData), strings.Join([]string{name, "+ServiceAccount"}, ""), nil
}

func createRoles(logicalcloud common.LogicalCloud, userpermissions []UserPermission) ([]string, []string, error) {
	var name string
	var kind string
	var datas []string
	var names []string

	roleCount := len(userpermissions)
	datas = make([]string, roleCount, roleCount)
	names = make([]string, roleCount, roleCount)

	for i, up := range userpermissions {
		if up.Specification.Namespace == "" {
			name = strings.Join([]string{logicalcloud.MetaData.Name, "-clusterRole", strconv.Itoa(i)}, "")
			kind = "ClusterRole"
		} else {
			name = strings.Join([]string{logicalcloud.MetaData.Name, "-role", strconv.Itoa(i)}, "")
			kind = "Role"
		}

		role := Resource{
			ApiVersion: "rbac.authorization.k8s.io/v1",
			Kind:       kind,
			MetaData: MetaDatas{
				Name: name,
				// Namespace: logicalcloud.Specification.NameSpace,
			},
			Rules: []RoleRules{RoleRules{
				ApiGroups: up.Specification.APIGroups,
				Resources: up.Specification.Resources,
				Verbs:     up.Specification.Verbs,
			},
			},
		}
		if up.Specification.Namespace != "" {
			role.MetaData.Namespace = up.Specification.Namespace
		}

		roleData, err := yaml.Marshal(&role)
		if err != nil {
			return []string{}, []string{}, err
		}

		datas[i] = string(roleData)
		names[i] = strings.Join([]string{name, "+", kind}, "")
	}

	return datas, names, nil
}

func createRoleBindings(ctx context.Context, logicalcloud common.LogicalCloud, userpermissions []UserPermission, clusterName string, gitOpsSupport bool) ([]string, []string, error) {
	var name string
	var kind string
	var kindbinding string
	var datas []string
	var names []string

	roleCount := len(userpermissions)
	datas = make([]string, roleCount, roleCount)
	names = make([]string, roleCount, roleCount)

	for i, up := range userpermissions {
		// empty namespace here means a cluster-scoped user permission
		if up.Specification.Namespace == "" {
			name = strings.Join([]string{logicalcloud.MetaData.Name, "-clusterRoleBinding", strconv.Itoa(i)}, "")
			kind = "ClusterRole"
			kindbinding = "ClusterRoleBinding"
			// otherwise it's a namespace-scoped user permission
		} else {
			name = strings.Join([]string{logicalcloud.MetaData.Name, "-roleBinding", strconv.Itoa(i)}, "")
			kind = "Role"
			kindbinding = "RoleBinding"
		}
		var subjects []RoleSubjects

		subjects = []RoleSubjects{RoleSubjects{
			Kind:     "User",
			Name:     logicalcloud.Specification.User.UserName,
			ApiGroup: "",
		},
		}

		if gitOpsSupport {
			gitops_type, err := GetGitOpsType(ctx, clusterName)
			if err != nil {
				return nil, nil, err
			}
			if gitops_type == "anthos" {
				// Google Anthos GitOps backend supports Privileged Logical Clouds using RepoSync.
				// This means Anthos auto-creates a Service Account for RepoSync, so we need to bind to it:
				// (Note: the actual RepoSync resource is automatically generated by rsync's anthos plugin)
				repoSyncName := strings.Join([]string{"repo-sync-", logicalcloud.MetaData.Name}, "")
				subjects = append(subjects, RoleSubjects{
					Kind:      "ServiceAccount",
					Name:      strings.Join([]string{"ns-reconciler-", logicalcloud.Specification.NameSpace, "-", repoSyncName, "-", fmt.Sprintf("%d", len(repoSyncName))}, ""),
					NameSpace: "config-management-system", // all Anthos Kubernetes Service Accounts belong to this namespace
				})
			} else {
				subjects = append(subjects, RoleSubjects{
					Kind:      "ServiceAccount",
					Name:      strings.Join([]string{logicalcloud.MetaData.Name, "-sa"}, ""),
					NameSpace: logicalcloud.Specification.NameSpace, // here we want the namespace where the ServiceAccount was created in
				})
			}
		}
		roleBinding := Resource{
			ApiVersion: "rbac.authorization.k8s.io/v1",
			Kind:       kindbinding,
			MetaData: MetaDatas{
				Name: name,
			},
			Subjects: subjects,

			RoleRefs: RoleRef{
				Kind:     kind,
				ApiGroup: "",
			},
		}
		if up.Specification.Namespace != "" {
			roleBinding.MetaData.Namespace = up.Specification.Namespace
			roleBinding.RoleRefs.Name = strings.Join([]string{logicalcloud.MetaData.Name, "-role", strconv.Itoa(i)}, "")
		} else {
			roleBinding.RoleRefs.Name = strings.Join([]string{logicalcloud.MetaData.Name, "-clusterRole", strconv.Itoa(i)}, "")
		}

		rBData, err := yaml.Marshal(&roleBinding)
		if err != nil {
			return []string{}, []string{}, err
		}
		datas[i] = string(rBData)
		names[i] = strings.Join([]string{name, "+", kindbinding}, "")
	}

	return datas, names, nil
}

func createQuotas(quotaList []Quota, namespace string) ([]string, []string, error) {

	var name string
	var datas []string
	var names []string

	quotaCount := len(quotaList)
	datas = make([]string, quotaCount, quotaCount)
	names = make([]string, quotaCount, quotaCount)

	for i, lcQuota := range quotaList {

		name = lcQuota.MetaData.QuotaName

		q := Resource{
			ApiVersion: "v1",
			Kind:       "ResourceQuota",
			MetaData: MetaDatas{
				Name:      name,
				Namespace: namespace,
			},
			Specification: Specs{
				Hard: lcQuota.Specification,
			},
		}

		qData, err := yaml.Marshal(&q)
		if err != nil {
			return []string{}, []string{}, err
		}
		datas[i] = string(qData)
		names[i] = strings.Join([]string{name, "+ResourceQuota"}, "")
	}

	return datas, names, nil
}

func createUserCSR(logicalcloud common.LogicalCloud, pkData string) (string, string, error) {
	pa, err := base64.StdEncoding.DecodeString(strings.Trim(pkData, "\""))
	if err != nil {
		return "", "", err
	}
	pb, _ := pem.Decode([]byte(pa))
	if pb == nil {
		return "", "", pkgerrors.New("Couldn't decode private key")
	}

	pk, err := x509.ParsePKCS1PrivateKey(pb.Bytes)
	if err != nil {
		return "", "", err
	}

	userName := logicalcloud.Specification.User.UserName
	name := strings.Join([]string{logicalcloud.MetaData.Name, "-", logicalcloud.Specification.NameSpace, "-user-csr"}, "")

	csrTemplate := x509.CertificateRequest{Subject: pkix.Name{CommonName: userName}}

	csrCert, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, pk)
	if err != nil {
		return "", "", err
	}

	//Encode csr
	csr := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrCert,
	})

	csrObj := Resource{
		ApiVersion: "certificates.k8s.io/v1",
		Kind:       "CertificateSigningRequest",
		MetaData: MetaDatas{
			Name: name,
		},
		Specification: Specs{
			Request:    base64.StdEncoding.EncodeToString(csr),
			Usages:     []string{"digital signature", "key encipherment", "client auth"},
			SignerName: "kubernetes.io/kube-apiserver-client",
		},
	}

	csrData, err := yaml.Marshal(&csrObj)
	if err != nil {
		return "", "", err
	}

	return string(csrData), strings.Join([]string{name, "+CertificateSigningRequest"}, ""), nil
}

func createApprovalSubresource(logicalcloud common.LogicalCloud) (string, error) {
	subresource := subresources.ApprovalSubresource{
		Message:        "Approved for Logical Cloud authentication",
		Reason:         "LogicalCloud",
		Type:           string(k8scertsv1.CertificateApproved),
		LastUpdateTime: metav1.Now().Format("2006-01-02T15:04:05Z"),
		Status:         "True",
	}
	csrData, err := json.Marshal(subresource)
	return string(csrData), err
}

/*
queryDBAndSetRsyncInfo queries the MCO db to find the record the sync controller
and then sets the RsyncInfo global variable.
*/
func queryDBAndSetRsyncInfo(ctx context.Context) (installappclient.RsyncInfo, error) {
	client := controller.NewControllerClient("resources", "data", "orchestrator")
	vals, _ := client.GetControllers(ctx)
	for _, v := range vals {
		if v.Metadata.Name == rsyncName {
			log.Info("Initializing RPC connection to resource synchronizer", log.Fields{
				"Controller": v.Metadata.Name,
			})
			rsyncInfo := installappclient.NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			return rsyncInfo, nil
		}
	}
	return installappclient.RsyncInfo{}, pkgerrors.Errorf("queryRsyncInfoInMCODB Failed - Could not get find rsync by name : %v", rsyncName)
}

// callRsyncInstall method shall take in the app context id and invoke the rsync service via grpc
func callRsyncInstall(ctx context.Context, contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo(ctx)
	log.Info("Calling rsync", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = installappclient.InvokeInstallApp(ctx, appContextID)
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}
	return nil
}

// callRsyncReadyNotify method shall take in the app context id and invoke the rsync ready-notify grpc api
func callRsyncReadyNotify(ctx context.Context, contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo(ctx)
	log.Info("Calling rsync", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	return InvokeReadyNotify(ctx, appContextID) // see dcm/pkg/module/client.go
}

// callRsyncUninstall method shall take in the app context id and invoke the rsync service via grpc
func callRsyncUninstall(ctx context.Context, contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo(ctx)
	log.Info("Calling rsync", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = installappclient.InvokeUninstallApp(ctx, appContextID)
	if err != nil {
		log.Error("", log.Fields{"err": err})
		return err
	}
	return nil
}

func callRsyncUpdate(ctx context.Context, FromContextid, ToContextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo(ctx)
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	fromAppContextID := fmt.Sprintf("%v", FromContextid)
	toAppContextID := fmt.Sprintf("%v", ToContextid)
	err = updateappclient.InvokeUpdateApp(ctx, fromAppContextID, toAppContextID)
	if err != nil {
		return err
	}
	return nil
}

// TODO: use appCtx.ctxid instead of passing cid
func prepL1ClusterAppContext(ctx context.Context, oldCid string, logicalcloud common.LogicalCloud, cluster common.Cluster, quotaList []Quota, userPermissionList []UserPermission, lcclient *LogicalCloudClient, lckey common.LogicalCloudKey, pkData string, appCtx appcontext.AppContext, cid string) error {
	logicalCloudName := logicalcloud.MetaData.Name
	clusterName := strings.Join([]string{cluster.Specification.ClusterProvider, "+", cluster.Specification.ClusterName}, "")
	appHandle, err := appCtx.GetAppHandle(ctx, lcAppName)                // caution: ignoring error
	clusterHandle, err := appCtx.AddCluster(ctx, appHandle, clusterName) // caution: ignoring error
	// pre-build array to pass to cleanupCompositeApp() [for performance]
	details := []string{logicalCloudName, clusterName, cid}

	if err != nil {
		return cleanupCompositeApp(ctx, appCtx, err, "Error adding Cluster to AppContext", details)
	}

	gitOps, err := IsGitOpsCluster(ctx, clusterName)
	if err != nil {
		return err
	}

	// Get resources to be added
	namespace, namespaceName, err := createNamespace(logicalcloud)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating Namespace YAML for logical cloud")
	}

	roles, roleNames, err := createRoles(logicalcloud, userPermissionList)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating Roles/ClusterRoles YAMLs for logical cloud")
	}

	roleBindings, roleBindingNames, err := createRoleBindings(ctx, logicalcloud, userPermissionList, clusterName, gitOps)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating RoleBindings/ClusterRoleBindings YAMLs for logical cloud")
	}

	quotas, quotaNames, err := createQuotas(quotaList, logicalcloud.Specification.NameSpace)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating Quota YAMLs for logical cloud")
	}

	// Add namespace resource to each cluster
	_, err = appCtx.AddResource(ctx, clusterHandle, namespaceName, namespace)
	if err != nil {
		return cleanupCompositeApp(ctx, appCtx, err, "Error adding Namespace Resource to AppContext", details)
	}
	resorderList := []string{namespaceName}
	if !gitOps {
		// then use it to generate a CSR for the cluster being processed
		csr, csrName, err := createUserCSR(logicalcloud, pkData)
		if err != nil {
			return pkgerrors.Wrap(err, "Error Creating User CSR and Key for logical cloud")
		}

		// check to see if - in case of an update - if the csr was already created for this logical cloud
		gotCsr, oldCsr := alreadyGotCsr(ctx, oldCid, appCtx, lcAppName, clusterName, csrName)
		if gotCsr {
			csr = oldCsr
		}

		approval, err := createApprovalSubresource(logicalcloud)
		if err != nil {
			return pkgerrors.Wrap(err, "Error Creating approval subresource for logical cloud")
		}
		// Add csr resource to each cluster
		csrHandle, err := appCtx.AddResource(ctx, clusterHandle, csrName, csr)
		if err != nil {
			return cleanupCompositeApp(ctx, appCtx, err, "Error adding CSR Resource to AppContext", details)
		}

		// Add csr approval as a subresource of csr
		_, err = appCtx.AddLevelValue(ctx, csrHandle, "subresource/approval", approval)
		if err != nil {
			return cleanupCompositeApp(ctx, appCtx, err, "Error approving CSR via AppContext", details)
		}
		resorderList = append(resorderList, csrName)

		// Add Subresource Order and Subresource Dependency
		subresOrder, err := json.Marshal(map[string][]string{"subresorder": []string{"approval"}})
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating subresource order JSON")
		}

		_, err = appCtx.AddInstruction(ctx, csrHandle, "subresource", "order", string(subresOrder))
		if err != nil {
			return cleanupCompositeApp(ctx, appCtx, err, "Error adding instruction order to AppContext", details)
		}
	} else {
		gitops_type, err := GetGitOpsType(ctx, clusterName)
		if err != nil {
			return err
		}
		if gitops_type != "anthos" {
			sa, saName, err := createServiceAccount(logicalcloud)
			if err != nil {
				return pkgerrors.Wrap(err, "Error Creating SA for logical cloud")
			}
			_, err = appCtx.AddResource(ctx, clusterHandle, saName, sa)
			if err != nil {
				return cleanupCompositeApp(ctx, appCtx, err, "Error adding SA Resource to AppContext", details)
			}
			resorderList = append(resorderList, saName)
		}
	}

	// Add [Cluster]Role resources to each cluster
	for i, roleName := range roleNames {
		_, err = appCtx.AddResource(ctx, clusterHandle, roleName, roles[i])
		if err != nil {
			return cleanupCompositeApp(ctx, appCtx, err, "Error adding [Cluster]Role Resource to AppContext", details)
		}
	}

	// Add [Cluster]RoleBinding resource to each cluster
	for i, roleBindingName := range roleBindingNames {
		_, err = appCtx.AddResource(ctx, clusterHandle, roleBindingName, roleBindings[i])
		if err != nil {
			return cleanupCompositeApp(ctx, appCtx, err, "Error adding [Cluster]RoleBinding Resource to AppContext", details)
		}
	}

	// Add quota resources to each cluster
	for i, quotaName := range quotaNames {
		_, err = appCtx.AddResource(ctx, clusterHandle, quotaName, quotas[i])
		if err != nil {
			return cleanupCompositeApp(ctx, appCtx, err, "Error adding quota Resource to AppContext", details)
		}
	}

	// Add Resource Order
	resorderList = append(resorderList, quotaNames...)
	resorderList = append(resorderList, roleNames...)
	resorderList = append(resorderList, roleBindingNames...)
	resOrder, err := json.Marshal(map[string][]string{"resorder": resorderList})
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating resource order JSON")
	}

	// Add Resource-level Order
	_, err = appCtx.AddInstruction(ctx, clusterHandle, "resource", "order", string(resOrder))
	if err != nil {
		return cleanupCompositeApp(ctx, appCtx, err, "Error adding instruction order to AppContext", details)
	}

	return nil
}

// blindInstantiateL0 attempts to instantiate a Level-0 logical cloud into the appContext without making any
// judgements about that logical cloud. It doesn't check for any current status for that logical cloud, it
// simply adds it to a new appcontext, with a new appcontext ID, even if it was previously there. Any
// judgements about the logical cloud, and any calls to rsync, are done by either Instantiate() or Update().
func blindInstantiateL0(ctx context.Context, project string, logicalcloud common.LogicalCloud, lcclient *LogicalCloudClient,
	clusterList []common.Cluster) (appcontext.AppContext, string, error) {
	var err error
	var appCtx appcontext.AppContext
	l0ns := ""
	logicalCloudName := logicalcloud.MetaData.Name
	lckey := common.LogicalCloudKey{
		LogicalCloudName: logicalCloudName,
		Project:          project,
	}
	// cycle through all clusters to obtain and validate the single level-0 namespace to use
	// the namespace of each cluster is retrieved from CloudConfig in rsync
	for _, cluster := range clusterList {

		ccc := rsync.NewCloudConfigClient()
		log.Info("Asking rsync's CloudConfig for this cluster's namespace at level-0", log.Fields{"cluster": cluster.Specification.ClusterName})
		ns, err := ccc.GetNamespace(
			ctx,
			cluster.Specification.ClusterProvider,
			cluster.Specification.ClusterName,
		)
		if err != nil {
			if err.Error() == "No CloudConfig was returned" {
				return appCtx, "", pkgerrors.New("It looks like the cluster provided as reference does not exist")
			}
			return appCtx, "", pkgerrors.Wrap(err, "Couldn't determine namespace for L0 logical cloud")
		}
		// we're checking here if any of the clusters have a differently-named namespace at level 0 and, if so,
		// we abort the instantiate operation because a single namespace name for this logical cloud cannot be inferred
		if len(l0ns) > 0 && ns != l0ns {
			log.Error("The clusters associated to this L0 logical cloud don't all share the same namespace name", log.Fields{"logicalcloud": logicalCloudName})
			return appCtx, "", pkgerrors.New("The clusters associated to this L0 logical cloud don't all share the same namespace name")
		}
		l0ns = ns
	}
	// if l0ns is still empty, something definitely went wrong so we can't let this pass
	if len(l0ns) == 0 {
		log.Error("Something went wrong as no cluster namespaces got checked", log.Fields{"logicalcloud": logicalCloudName})
		return appCtx, "", pkgerrors.New("Something went wrong as no cluster namespaces got checked")
	}
	// at this point we know what namespace name to give to the logical cloud
	logicalcloud.Specification.NameSpace = l0ns
	// the following is an update operation:
	err = db.DBconn.Insert(ctx, lcclient.storeName, lckey, nil, lcclient.tagMeta, logicalcloud)
	if err != nil {
		log.Error("Failed to update L0 logical cloud with a namespace name", log.Fields{"logicalcloud": logicalCloudName, "namespace": l0ns})
		return appCtx, "", pkgerrors.Wrap(err, "Failed to update L0 logical cloud with a namespace name")
	}
	log.Info("The L0 logical cloud has been updated with a namespace name", log.Fields{"logicalcloud": logicalCloudName, "namespace": l0ns})

	// prepare empty-shell appcontext for the L0 LC in order to officially set it as Instantiated
	appCtx = appcontext.AppContext{}
	ctxVal, err := appCtx.InitAppContext()
	if err != nil {
		return appCtx, "", pkgerrors.Wrap(err, "Error creating L0 LC AppContext")
	}
	cid := ctxVal.(string)

	handle, err := appCtx.CreateCompositeApp(ctx)
	if err != nil {
		return appCtx, "", pkgerrors.Wrap(err, "Error creating L0 LC AppContext CompositeApp")
	}

	appHandle, err := appCtx.AddApp(ctx, handle, lcAppName)
	if err != nil {
		return appCtx, "", cleanupCompositeApp(ctx, appCtx, err, "Error adding App to L0 LC AppContext", []string{logicalCloudName, cid})
	}

	// Create a Logical Cloud Meta with basic information about the Logical Cloud:
	// project name and logical cloud name
	err = appCtx.AddCompositeAppMeta(
		ctx,
		appcontext.CompositeAppMeta{
			Project:      project,
			LogicalCloud: logicalCloudName})
	if err != nil {
		return appCtx, "", cleanupCompositeApp(ctx, appCtx, err, "Error Adding Logical Cloud Meta to AppContext", []string{logicalCloudName, cid})
	}

	// iterate through cluster list and add all the clusters (as empty-shells)
	for _, cluster := range clusterList {
		clusterName := strings.Join([]string{cluster.Specification.ClusterProvider, "+", cluster.Specification.ClusterName}, "")
		clusterHandle, err := appCtx.AddCluster(ctx, appHandle, clusterName)
		// pre-build array to pass to cleanupCompositeApp() [for performance]
		details := []string{logicalCloudName, clusterName, cid}

		if err != nil {
			return appCtx, "", cleanupCompositeApp(ctx, appCtx, err, "Error adding Cluster to L0 LC AppContext", details)
		}

		// resource-level order is mandatory too for an empty-shell appcontext
		resOrder, err := json.Marshal(map[string][]string{"resorder": []string{}})
		if err != nil {
			return appCtx, "", pkgerrors.Wrap(err, "Error creating resource order JSON")
		}
		_, err = appCtx.AddInstruction(ctx, clusterHandle, "resource", "order", string(resOrder))
		if err != nil {
			return appCtx, "", cleanupCompositeApp(ctx, appCtx, err, "Error adding resource-level order to L0 LC AppContext", details)
		}
		// TODO add resource-level dependency as well
		// app-level order is mandatory too for an empty-shell appcontext
		appOrder, err := json.Marshal(map[string][]string{"apporder": []string{lcAppName}})
		if err != nil {
			return appCtx, "", pkgerrors.Wrap(err, "Error creating app order JSON")
		}
		_, err = appCtx.AddInstruction(ctx, handle, "app", "order", string(appOrder))
		if err != nil {
			return appCtx, "", cleanupCompositeApp(ctx, appCtx, err, "Error adding app-level order to L0 LC AppContext", details)
		}
		// TODO add app-level dependency as well
		// TODO move app-level order/dependency out of loop
	}
	return appCtx, cid, nil
}

func alreadyGotCsr(ctx context.Context, oldCid string, newAc appcontext.AppContext, app, cluster, csrName string) (bool, string) {
	if oldCid == "" {
		return false, ""
	}
	var oldAc appcontext.AppContext
	_, err := oldAc.LoadAppContext(ctx, oldCid)
	if err != nil {
		log.Info("[CSR} Error getting old appcontext",
			log.Fields{"oldCid": oldCid, "app": app, "cluster": cluster, "csrName": csrName})
		return false, ""
	}

	csrH, err := oldAc.GetResourceHandle(ctx, app, cluster, csrName)
	if err != nil {
		log.Info("[CSR} Error getting old appcontext resource handle",
			log.Fields{"oldCid": oldCid, "app": app, "cluster": cluster, "csrName": csrName})
		return false, ""
	}
	csr, err := oldAc.GetValue(ctx, csrH)
	if err != nil {
		log.Info("[CSR} Error getting old appcontext resource handle",
			log.Fields{"oldCid": oldCid, "app": app, "cluster": cluster, "csrName": csrName})
		return false, ""
	}
	log.Trace("[CSR} Got old CSR",
		log.Fields{"oldCid": oldCid, "app": app, "cluster": cluster, "csrName": csrName, "csr": csr})
	return true, csr.(string)
}

// // blindInstantiateL1 is the equivalent of blindInstantiateL0 but for Level-1 Logical Clouds.
func blindInstantiateL1(ctx context.Context, oldCid string, project string, logicalcloud common.LogicalCloud, lcclient *LogicalCloudClient,
	clusterList []common.Cluster, quotaList []Quota, userPermissionList []UserPermission) (appcontext.AppContext, string, error) {
	var err error
	var appCtx appcontext.AppContext
	logicalCloudName := logicalcloud.MetaData.Name
	lckey := common.LogicalCloudKey{
		LogicalCloudName: logicalCloudName,
		Project:          project,
	}

	if len(userPermissionList) == 0 {
		return appCtx, "", pkgerrors.New("Level-1 Logical Clouds require at least a User Permission assigned to its primary namespace")
	}

	primaryUP := false
	for _, up := range userPermissionList {
		if up.Specification.Namespace == logicalcloud.Specification.NameSpace {
			primaryUP = true
			break
		}
	}
	if !primaryUP {
		return appCtx, "", pkgerrors.New("Level-1 Logical Clouds require a User Permission assigned to its primary namespace")
	}

	// From this point on, we are dealing with a new AppContext
	appCtx = appcontext.AppContext{}
	ctxVal, err := appCtx.InitAppContext()
	if err != nil {
		return appCtx, "", pkgerrors.Wrap(err, "Error creating AppContext")
	}
	cid := ctxVal.(string)

	handle, err := appCtx.CreateCompositeApp(ctx)
	if err != nil {
		return appCtx, "", pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}

	_, err = appCtx.AddApp(ctx, handle, lcAppName)
	if err != nil {
		return appCtx, "", cleanupCompositeApp(ctx, appCtx, err, "Error adding App to AppContext", []string{logicalCloudName, cid})
	}

	// Create a Logical Cloud Meta with all data needed for a successful L1 (standard/privileged) instantiation:
	// Project name, LogicalCloud name,
	// LogicalCloudNamespace (user-intended namespace for logical cloud),
	// LogicalCloudLevel (user-intended level of logical cloud),
	// Level (always "0" - level of logical cloud's compositeapp),
	// and Namespace (always "default" - namespace of logical cloud's compositeapp)
	err = appCtx.AddCompositeAppMeta(
		ctx,
		appcontext.CompositeAppMeta{
			Project:               project,
			LogicalCloud:          logicalCloudName,
			LogicalCloudNamespace: logicalcloud.Specification.NameSpace,
			LogicalCloudLevel:     logicalcloud.Specification.Level,
			Level:                 "0",
			Namespace:             "default"})
	if err != nil {
		return appCtx, "", cleanupCompositeApp(ctx, appCtx, err, "Error Adding Logical Cloud Meta to AppContext", []string{logicalCloudName, cid})
	}

	// Add App Order and App Dependency
	appOrder, err := json.Marshal(map[string][]string{"apporder": []string{lcAppName}})
	if err != nil {
		return appCtx, "", pkgerrors.Wrap(err, "Error creating app order JSON")
	}
	appDependency, err := json.Marshal(map[string]map[string]string{"appdependency": map[string]string{lcAppName: "go"}})
	if err != nil {
		return appCtx, "", pkgerrors.Wrap(err, "Error creating app dependency JSON")
	}

	// Add App-level Order and Dependency
	_, err = appCtx.AddInstruction(ctx, handle, "app", "order", string(appOrder))
	if err != nil {
		return appCtx, "", cleanupCompositeApp(ctx, appCtx, err, "Error adding app-level order to AppContext", []string{logicalCloudName})
	}
	_, err = appCtx.AddInstruction(ctx, handle, "app", "dependency", string(appDependency))
	if err != nil {
		return appCtx, "", cleanupCompositeApp(ctx, appCtx, err, "Error adding app-level dependency to AppContext", []string{logicalCloudName})
	}

	// get pkData from the database, which was created during logical cloud Create()
	pkDataArray, err := db.DBconn.Find(ctx, lcclient.storeName, lckey, "privatekey")
	if err != nil {
		return appCtx, "", cleanupCompositeApp(ctx, appCtx, err, "Error getting private key from logical cloud", []string{logicalCloudName})
	}
	if len(pkDataArray) == 0 {
		return appCtx, "", cleanupCompositeApp(ctx, appCtx, err, "Private key from logical cloud not found", []string{logicalCloudName})
	}
	privKey := common.PrivateKey{}
	err = db.DBconn.Unmarshal(pkDataArray[0], &privKey)
	if err != nil {
		return appCtx, "", cleanupCompositeApp(ctx, appCtx, err, "Private key unmarshal error", []string{logicalCloudName})
	}

	// Iterate through cluster list and add all the clusters
	for _, cluster := range clusterList {
		err = prepL1ClusterAppContext(ctx, oldCid, logicalcloud, cluster, quotaList, userPermissionList, lcclient, lckey, privKey.KeyValue, appCtx, cid)
		if err != nil {
			return appCtx, "", err
		}
	}

	return appCtx, cid, nil
}

// Instantiate prepares all yaml resources to be given to the clusters via rsync,
// then creates an appcontext with such resources and asks rsync to instantiate the logical cloud
func Instantiate(ctx context.Context, project string, logicalcloud common.LogicalCloud, clusterList []common.Cluster,
	quotaList []Quota, userPermissionList []UserPermission) error {

	logicalCloudName := logicalcloud.MetaData.Name
	level := logicalcloud.Specification.Level

	lcclient := NewLogicalCloudClient()

	// Check if there was a previous appCtx for this logical cloud
	s, err := lcclient.GetState(ctx, project, logicalCloudName)
	if err != nil {
		return err
	}
	cid := state.GetLastContextIdFromStateInfo(s)
	if cid != "" {
		ac, err := state.GetAppContextFromId(ctx, cid)
		if err != nil {
			return err
		}
		acStatus, err := state.GetAppContextStatus(ctx, cid) // new from state
		if err != nil {
			return err
		}

		// If we're trying to instantiate a stopped termination, first clear the stop flag
		stateVal, err := state.GetCurrentStateFromStateInfo(s)
		if err != nil {
			return err
		}
		if stateVal == state.StateEnum.TerminateStopped {
			err = state.UpdateAppContextStopFlag(ctx, cid, false)
			if err != nil {
				return err
			}
		}

		// Make sure rsync status for this logical cloud is Terminated,
		// otherwise we can't re-instantiate logical cloud yet
		switch acStatus.Status {
		case appcontext.AppContextStatusEnum.Terminated:
			// Fully delete the old AppContext
			err := ac.DeleteCompositeApp(ctx)
			if err != nil {
				log.Error("Error deleting AppContext CompositeApp Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
				return pkgerrors.Wrap(err, "Error deleting AppContext CompositeApp Logical Cloud")
			}
		case appcontext.AppContextStatusEnum.Terminating:
			log.Error("The Logical Cloud can't be re-instantiated yet, it is being terminated", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud can't be re-instantiated yet, it is being terminated")
		case appcontext.AppContextStatusEnum.Instantiated:
			log.Error("The Logical Cloud is already instantiated", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud is already instantiated")
		case appcontext.AppContextStatusEnum.Instantiating:
			log.Error("The Logical Cloud is already instantiating", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud is already instantiating")
		case appcontext.AppContextStatusEnum.InstantiateFailed:
			log.Error("The Logical Cloud has failed instantiating before, please terminate and try again", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud has failed instantiating before, please terminate and try again")
		case appcontext.AppContextStatusEnum.TerminateFailed:
			log.Error("The Logical Cloud has failed terminating, please delete the Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud has failed terminating, please delete the Logical Cloud")
		default:
			log.Error("The Logical Cloud isn't in an expected status so not taking any action", log.Fields{"logicalcloud": logicalCloudName, "status": acStatus.Status})
			return pkgerrors.New("The Logical Cloud isn't in an expected status so not taking any action")
		}
	}

	// still need this because one final cleanupCompositeApp might be needed
	var appCtx appcontext.AppContext
	var newcid string
	// TODO: use appCtx.ctxid instead of returning newcid
	if level == "0" {
		appCtx, newcid, err = blindInstantiateL0(ctx, project, logicalcloud, lcclient, clusterList)
	} else {
		appCtx, newcid, err = blindInstantiateL1(ctx, "", project, logicalcloud, lcclient, clusterList, quotaList, userPermissionList)
	}
	if err != nil {
		return err
	}

	// Call rsync to install Logical Cloud in clusters
	err = callRsyncInstall(ctx, newcid)
	if err != nil {
		log.Error("Failed calling rsync install-app", log.Fields{"err": err})
		return pkgerrors.Wrap(err, "Failed calling rsync install-app")
	}

	// Update state with switch to Instantiated state, along with storing the AppContext ID for future retrieval
	err = addState(ctx, lcclient, project, logicalCloudName, newcid, state.StateEnum.Instantiated)
	if err != nil {
		return cleanupCompositeApp(ctx, appCtx, err, "Error adding Logical Cloud AppContext to DB", []string{logicalCloudName, newcid})
	}

	if level == "1" {
		// Call rsync grpc streaming api, which launches a goroutine to wait for the response of
		// every cluster (function should know how many clusters are expected and only finish when
		// all respective certificates have been obtained and all kubeconfigs stored in CloudConfig)
		err = callRsyncReadyNotify(ctx, newcid)
		if err != nil {
			log.Error("Failed calling rsync ready-notify", log.Fields{"err": err})
			return pkgerrors.Wrap(err, "Failed calling rsync ready-notify")
		}
		// However, for GitOps the above doesn't apply, so let's iterate over all clusters and prep L1 CloudConfig for those which are GitOps:
		ccc := rsync.NewCloudConfigClient()
		for _, cluster := range clusterList {
			clusterName := strings.Join([]string{cluster.Specification.ClusterProvider, "+", cluster.Specification.ClusterName}, "")
			gitOps, err := IsGitOpsCluster(ctx, clusterName)
			if err != nil {
				return err
			}
			if gitOps {
				// With a GitOps backend, there is no need to wait for ReadyNotify.
				// Create a record in Rsync for logical cloud (level 1) and this cluster:
				_, err = ccc.CreateCloudConfig(
					ctx,
					cluster.Specification.ClusterProvider,
					cluster.Specification.ClusterName,
					level,
					logicalcloud.Specification.NameSpace,
					"")
				if err != nil {
					if err.Error() != "CloudConfig already exists" {
						return pkgerrors.Wrap(err, "Failed creating a new GitOps logical cloud in rsync's CloudConfig")
					}
				}
				// Fetch GitOps config for respective level-0 CloudConfig, in order to copy the contents into the level-1 entry
				goc, err := ccc.GetGitOpsConfig(
					ctx,
					cluster.Specification.ClusterProvider,
					cluster.Specification.ClusterName,
					"0",
					"default")
				if err != nil {
					return pkgerrors.Wrap(err, "Error fetching cloud config for GitOps at level-0")
				}
				_, err = ccc.CreateGitOpsConfig(
					ctx,
					cluster.Specification.ClusterProvider,
					cluster.Specification.ClusterName,
					goc.Config,
					level,
					logicalcloud.Specification.NameSpace,
				)
				if err != nil {
					return pkgerrors.Wrap(err, "Error creating cloud config for GitOps at level-1")
				}
				return nil
			}
		}
	}

	return nil
}

// Terminate asks rsync to terminate the logical cloud
func Terminate(ctx context.Context, project string, logicalcloud common.LogicalCloud, clusterList []common.Cluster,
	quotaList []Quota) error {

	logicalCloudName := logicalcloud.MetaData.Name
	level := logicalcloud.Specification.Level
	namespace := logicalcloud.Specification.NameSpace

	lcclient := NewLogicalCloudClient()

	// Check if there was a previous appCtx for this logical cloud
	s, err := lcclient.GetState(ctx, project, logicalCloudName)
	if err != nil {
		return err
	}
	cid := state.GetLastContextIdFromStateInfo(s)
	if cid != "" {

		// If we're trying to terminate a stopped instantiation, first clear the stop flag
		stateVal, err := state.GetCurrentStateFromStateInfo(s)
		if err != nil {
			return err
		}
		if stateVal == state.StateEnum.InstantiateStopped {
			err = state.UpdateAppContextStopFlag(ctx, cid, false)
			if err != nil {
				return err
			}
		}

		// Make sure rsync status for this logical cloud is Terminated,
		// otherwise we can't re-instantiate logical cloud yet
		acStatus, err := state.GetAppContextStatus(ctx, cid) // new from state
		if err != nil {
			return err
		}
		switch acStatus.Status {
		case appcontext.AppContextStatusEnum.Terminated:
			log.Error("The Logical Cloud has already been terminated", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud has already been terminated")
		case appcontext.AppContextStatusEnum.Terminating:
			log.Error("The Logical Cloud is already being terminated", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud is already being terminated")
		case appcontext.AppContextStatusEnum.Instantiating:
			log.Error("The Logical Cloud is still instantiating", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud is still instantiating")
		case appcontext.AppContextStatusEnum.TerminateFailed:
			log.Error("The Logical Cloud has failed terminating, please delete the Logical Cloud", log.Fields{"logicalcloud": logicalCloudName})
			return pkgerrors.New("The Logical Cloud has failed terminating, please delete the Logical Cloud")
		case appcontext.AppContextStatusEnum.InstantiateFailed:
			// try to terminate anyway
			fallthrough
		case appcontext.AppContextStatusEnum.Instantiated:
			// call resource synchronizer to delete the CRs from every cluster of the logical cloud
			err = callRsyncUninstall(ctx, cid)
			if err != nil {
				return err
			}
			// destroy kubeconfigs from CloudConfig if this is an L1 logical cloud
			if level == "1" {

				ccc := rsync.NewCloudConfigClient()
				for _, cluster := range clusterList {
					log.Info("Destroying CloudConfig of logicalcloud/cluster pair via rsync", log.Fields{"cluster": cluster.Specification.ClusterName, "logicalcloud": logicalCloudName, "level": level})
					err = ccc.DeleteCloudConfig(
						ctx,
						cluster.Specification.ClusterProvider,
						cluster.Specification.ClusterName,
						level,
						namespace,
					)

					if err != nil {
						log.Error("Failed destroying at least one CloudConfig of L1 LC", log.Fields{"cluster": cluster, "err": err})
						// continue terminating and removing any remaining CloudConfigs
						// (this happens when terminating a Logical Cloud before all kubeconfigs had a chance to be generated, such as after a stopped instantiation)
					}
				}
			}

			// Set State as Terminated
			err = addState(ctx, lcclient, project, logicalCloudName, cid, state.StateEnum.Terminated)
			if err != nil {
				return err // error already logged
			}

		default:
			log.Error("The Logical Cloud isn't in an expected status so not taking any action", log.Fields{"logicalcloud": logicalCloudName, "status": acStatus.Status})
			return pkgerrors.New("The Logical Cloud isn't in an expected status so not taking any action")
		}
	}
	return nil
}

// Stop asks rsync to stop the instantiation or termination of the logical cloud
func Stop(ctx context.Context, project string, logicalcloud common.LogicalCloud) error {

	logicalCloudName := logicalcloud.MetaData.Name
	lcclient := NewLogicalCloudClient()

	// Find and deal with state
	s, err := lcclient.GetState(ctx, project, logicalCloudName)
	if err != nil {
		return err
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting current state from LogicalCloud stateInfo: "+logicalCloudName)
	}

	cid := state.GetLastContextIdFromStateInfo(s)

	stopState := state.StateEnum.Undefined
	switch stateVal {
	case state.StateEnum.Created:
		return pkgerrors.New("LogicalCloud has not been asked to instantiate: " + logicalCloudName)
	case state.StateEnum.Instantiated:
		stopState = state.StateEnum.InstantiateStopped
	case state.StateEnum.Terminated:
		stopState = state.StateEnum.TerminateStopped
	case state.StateEnum.TerminateStopped:
		return pkgerrors.New("LogicalCloud termination already stopped: " + logicalCloudName)
	case state.StateEnum.InstantiateStopped:
		return pkgerrors.New("LogicalCloud instantiation already stopped: " + logicalCloudName)
	default:
		return pkgerrors.New("LogicalCloud is in an invalid state: " + logicalCloudName + " " + stateVal)
	}

	err = state.UpdateAppContextStopFlag(ctx, cid, true)
	if err != nil {
		return err
	}

	err = addState(ctx, lcclient, project, logicalCloudName, cid, stopState)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of the LogicalCloud: "+logicalCloudName)
	}

	return nil
}

func GetGitOpsType(ctx context.Context, clustername string) (string, error) {
	if !strings.Contains(clustername, "+") {
		log.Debug("GitOps Not a valid cluster name", log.Fields{})
		return "", pkgerrors.New("GitOps Not a valid cluster name")
	}
	strs := strings.Split(clustername, "+")
	if len(strs) != 2 {
		log.Debug("GitOps Not a valid cluster name", log.Fields{})
		return "", pkgerrors.New("GitOps Not a valid cluster name")
	}

	ccc := rsync.NewCloudConfigClient()

	gc, err := ccc.GetGitOpsConfig(ctx, strs[0], strs[1], "0", "")
	if err != nil {
		log.Debug("Error getting GitOps config", log.Fields{"err": err})
		return "", pkgerrors.New("Cluster is not GitOps")
	}
	return gc.Config.Props.GitOpsType, nil
}

func IsGitOpsCluster(ctx context.Context, clustername string) (bool, error) {
	if !strings.Contains(clustername, "+") {
		log.Debug("GitOps Not a valid cluster name", log.Fields{})
		return false, pkgerrors.New("GitOps Not a valid cluster name")
	}
	strs := strings.Split(clustername, "+")
	if len(strs) != 2 {
		log.Debug("GitOps Not a valid cluster name", log.Fields{})
		return false, pkgerrors.New("GitOps Not a valid cluster name")
	}

	ccc := rsync.NewCloudConfigClient()

	gc, err := ccc.GetGitOpsConfig(ctx, strs[0], strs[1], "0", "")
	if err != nil {
		log.Debug("Error getting GitOps config", log.Fields{"err": err})
		return false, nil
	}
	if gc.Config.Props.GitOpsType == "fluxcd" || gc.Config.Props.GitOpsType == "azureArcV2" || gc.Config.Props.GitOpsType == "anthos" {
		return true, nil
	} else {
		log.Info("GitOps Type not supported:", log.Fields{"GitOpsType": gc.Config.Props.GitOpsType})
		return false, pkgerrors.New("GitOps Type not supported: " + gc.Config.Props.GitOpsType)
	}
}
