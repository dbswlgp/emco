/// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package sdewancc

import (
	"encoding/json"
	"fmt"
	"strings"

	"context"

	pkgerrors "github.com/pkg/errors"
	clusterPkg "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	"gitlab.com/project-emco/core/emco-base/src/dtc/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
)

type clusterData struct {
	Reslist     []map[string][]byte //resname: res
	ClusterName string
	CNFPort     string
	ServicePort string
	AllowedCIDR string
}
type client struct {
	ClientName        string
	ClientServiceName string
	InstallClientRes  bool
	ClusterData       []clusterData
}
type serverData struct {
	AppName          string
	ServiceName      string
	ClusterData      []clusterData
	Clients          []client
	InstallServerRes bool
}

// Action applies the supplied intent against the given AppContext ID
func UpdateAppContext(ctx context.Context, intentName, appContextId string) error {
	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(ctx, appContextId)
	if err != nil {
		log.Error("Error loading AppContext", log.Fields{
			"error": err,
		})
		return pkgerrors.Wrapf(err, "Error loading AppContext with Id: %v", appContextId)
	}

	caMeta, err := ac.GetCompositeAppMeta(ctx)
	if err != nil {
		log.Error("Error getting metadata from AppContext", log.Fields{
			"error": err,
		})
		return pkgerrors.Wrapf(err, "Error getting metadata from AppContext with Id: %v", appContextId)
	}

	project := caMeta.Project
	compositeapp := caMeta.CompositeApp
	compositeappversion := caMeta.Version
	deployIntentGroup := caMeta.DeploymentIntentGroup
	namespace := caMeta.Namespace

	// Get all server inbound intents
	iss, err := module.NewServerInboundIntentClient().GetServerInboundIntents(ctx, project, compositeapp, compositeappversion, deployIntentGroup, intentName)
	if err != nil {
		log.Error("Error getting server inbound intents", log.Fields{
			"error": err,
		})
		return pkgerrors.Wrapf(err, "Error getting server inbound intents %v for %v/%v%v/%v not found", intentName, project, compositeapp, deployIntentGroup, compositeappversion)
	}

	l := len(iss)
	servers := make([]serverData, l)
	index := 0

	for _, is := range iss {
		if is.Spec.EdgeCNF != "sdewan" {
			log.Error("Error SDEWAN not enabled for this server", log.Fields{
				"error":    err,
				"app name": is.Spec.AppName,
			})
			return pkgerrors.Wrapf(err, "Error SDEWAN not enabled for this server")
		}
		clusters, err := ac.GetClusterNames(ctx, is.Spec.AppName)
		if err != nil {
			log.Error("Error retrieving clusters from App Context", log.Fields{
				"error":    err,
				"app name": is.Spec.AppName,
			})
			return pkgerrors.Wrapf(err,
				"Error retrieving clusters from App Context for app %v", is.Spec.AppName)
		}

		servers[index].AppName = is.Spec.AppName
		servers[index].ServiceName = is.Spec.ServiceName
		lc := len(clusters)
		servers[index].ClusterData = make([]clusterData, lc)
		for ci, c := range clusters {
			obj, err := getClusterKvPair(ctx, c, "SdewanCnfPort")
			if err != nil {
				log.Error("Error getting sdewan cnf port", log.Fields{
					"error": err,
				})
				return pkgerrors.Wrapf(err,
					"Error getting sdewan cnf port")
			}
			servers[index].ClusterData[ci].CNFPort = obj
			obj, err = getClusterKvPair(ctx, c, "SdewanServicePort")
			if err != nil {
				log.Error("Error getting sdewan service port", log.Fields{
					"error": err,
				})
				return pkgerrors.Wrapf(err,
					"Error getting sdewan service port")
			}
			servers[index].ClusterData[ci].ServicePort = obj
			obj, err = getClusterKvPair(ctx, c, "SdewanServiceAllowedCidr")
			if err != nil {
				log.Error("Error getting sdewan service allowed CIDR", log.Fields{
					"error": err,
				})
				return pkgerrors.Wrapf(err,
					"Error getting sdewan service allowed CIDR")
			}
			servers[index].ClusterData[ci].AllowedCIDR = obj
			servers[index].ClusterData[ci].ClusterName = c
			servers[index].ClusterData[ci].Reslist = make([]map[string][]byte, 0)
		}
		ics, err := module.NewClientsInboundIntentClient().GetClientsInboundIntents(ctx, project,
			compositeapp,
			compositeappversion,
			deployIntentGroup,
			intentName,
			is.Metadata.Name)
		if err != nil {
			log.Error("Error getting clients inbound intents", log.Fields{
				"error": err,
			})
			return pkgerrors.Wrapf(err,
				"Error getting clients inbound intents %v under server inbound intent %v for %v/%v%v/%v not found",
				is.Metadata.Name, intentName, project, compositeapp, compositeappversion, deployIntentGroup)
		}

		li := len(ics)
		servers[index].Clients = make([]client, li)
		for i, ic := range ics {
			servers[index].Clients[i].ClientName = ic.Spec.AppName
			servers[index].Clients[i].ClientServiceName = ic.Spec.ServiceName
			clusters, err = ac.GetClusterNames(ctx, ic.Spec.AppName)
			if err != nil {
				log.Error("Error retrieving clusters from App Context", log.Fields{
					"error":    err,
					"app name": ic.Spec.AppName,
				})
				return pkgerrors.Wrapf(err,
					"Error retrieving clusters from App Context for app %v", is.Spec.AppName)
			}
			lc := len(clusters)
			servers[index].Clients[i].ClusterData = make([]clusterData, lc)
			for cci, c := range clusters {
				servers[index].Clients[i].ClusterData[cci].ClusterName = c
				servers[index].Clients[i].ClusterData[cci].Reslist = make([]map[string][]byte, 0)

				err = createClientResources(is, c, servers, namespace, index, i, cci)
				if err != nil {
					log.Error("Error creating client resources for SDEWAN Application", log.Fields{
						"error":    err,
						"svc name": ic.Spec.ServiceName,
					})
					return err
				}
			}
		}
		// create the specified server resources
		for ci, scd := range servers[index].ClusterData {
			err = createServerResources(is, scd.ClusterName, servers, namespace, index, ci)
			if err != nil {
				log.Error("Error creating server resources", log.Fields{
					"error":    err,
					"svc name": is.Spec.ServiceName,
				})
				return pkgerrors.Wrapf(err,
					"Error creating server resources")
			}
		}
		index = index + 1

	}
	for _, s := range servers {
		// Add server resources
		for _, cd := range s.ClusterData {
			if len(cd.Reslist) <= 0 {
				continue
			}
			for _, r := range cd.Reslist {
				err = addClusterResource(ctx, ac, s.AppName, cd.ClusterName, r)
				if err != nil {
					log.Error("Error adding cluster Resource", log.Fields{
						"error":    err,
						"app name": s.AppName,
					})
					return pkgerrors.Wrapf(err, "Error adding cluster resource for %v", s.AppName)
				}
			}
		}
		for ci, cc := range s.Clients {
			//Add client resources
			for _, clu := range s.Clients[ci].ClusterData {
				if len(clu.Reslist) <= 0 {
					continue
				}
				for _, r := range clu.Reslist {
					err = addClusterResource(ctx, ac, cc.ClientName, clu.ClusterName, r)
					if err != nil {
						log.Error("Error adding cluster Resource", log.Fields{
							"error":    err,
							"app name": cc.ClientName,
						})
						return pkgerrors.Wrapf(err, "Error adding cluster resource for %v", s.AppName)
					}
				}
			}
		}
	}

	return nil
}

func addClusterResource(ctx context.Context, ac appcontext.AppContext, appname string, c string, res map[string][]byte) error {
	ch, err := ac.GetClusterHandle(ctx, appname, c)
	if err != nil {
		log.Error("Error getting clusters handle App Context", log.Fields{
			"error":        err,
			"app name":     appname,
			"cluster name": c,
		})
		return pkgerrors.Wrapf(err,
			"Error getting clusters from App Context for app %v and cluster %v", appname, c)
	}
	// Add resource to the cluster

	if len(res) != 1 {
		log.Error("Error validating  resource value", log.Fields{
			"error":        err,
			"app name":     appname,
			"cluster name": c,
		})
		return pkgerrors.Wrapf(err, "Error validating resource value")
	}
	var resname string
	var r []byte
	for rname, ro := range res {
		resname = rname
		r = ro
	}

	_, err = ac.AddResource(ctx, ch, resname, string(r))
	if err != nil {
		log.Error("Error adding Resource to AppContext", log.Fields{
			"error":        err,
			"app name":     appname,
			"cluster name": c,
		})
		return pkgerrors.Wrap(err, "Error adding Resource to AppContext")
	}
	resorder, err := ac.GetResourceInstruction(ctx, appname, c, "order")
	if err != nil {
		log.Error("Error getting Resource order", log.Fields{
			"error":        err,
			"app name":     appname,
			"cluster name": c,
		})
		return pkgerrors.Wrap(err, "Error getting Resource order")
	}
	aov := make(map[string][]string)
	json.Unmarshal([]byte(resorder.(string)), &aov)
	aov["resorder"] = append(aov["resorder"], resname)
	jresord, _ := json.Marshal(aov)

	_, err = ac.AddInstruction(ctx, ch, "resource", "order", string(jresord))
	if err != nil {
		log.Error("Error updating Resource order", log.Fields{
			"error":        err,
			"app name":     appname,
			"cluster name": c,
		})
		return pkgerrors.Wrap(err, "Error updating Resource order")
	}
	return nil
}

func createServerResources(is module.InboundServerIntent, c string, servers []serverData, namespace string, index, ci int) error {

	cnfport := servers[index].ClusterData[ci].CNFPort
	serviceport := servers[index].ClusterData[ci].ServicePort
	cidr := servers[index].ClusterData[ci].AllowedCIDR
	fullname := is.Spec.ServiceName + "." + namespace + "." + "svc.cluster.local"
	res, err := createSdewanService(fullname, namespace, cnfport, serviceport, cidr)
	if err != nil {
		log.Error("Error creating SDEWAN Service", log.Fields{
			"error":        err,
			"svc name":     is.Spec.ServiceName,
			"cluster name": c,
		})
		return pkgerrors.Wrap(err, "Error creating SDEWAN Service")
	}
	servers[index].ClusterData[ci].Reslist = append(servers[index].ClusterData[ci].Reslist, res)

	return nil
}

func createSdewanApplication(svcname, namespace string, pslabel string, serviceport string, cnfport string) (map[string][]byte, error) {
	salabel := createPodSelector(pslabel)
	saspec := createSdewanApplicationSpec(namespace, salabel, serviceport, cnfport)
	meta := createGenericMetadata(svcname, namespace, "")
	out, err := createSdewanApplicationResource(meta, saspec)

	if err != nil {
		log.Error("Error creating SdewanApplication Resource", log.Fields{
			"error":    err,
			"svc name": svcname,
		})
		return nil, pkgerrors.Wrap(err, "Error creating SdewanApplication Resource")
	}

	res := make(map[string][]byte)
	resname := svcname + "-sa"
	res[resname] = out
	return res, nil
}

func createSdewanService(svcname, namespace string, port string, dport string, cidr string) (map[string][]byte, error) {
	ssspec := createSdewanServiceSpec(svcname, port, dport, cidr)
	meta := createGenericMetadata(svcname, namespace, "")
	out, err := createSdewanServiceResource(meta, ssspec)

	if err != nil {
		log.Error("Error creating SdewanApplication Resource", log.Fields{
			"error":    err,
			"svc name": svcname,
		})
		return nil, pkgerrors.Wrap(err, "Error creating SdewanApplication Resource")
	}

	res := make(map[string][]byte)
	resname := svcname + "-sa"
	res[resname] = out
	return res, nil
}

func createClientResources(is module.InboundServerIntent, c string, servers []serverData, namespace string, index, ci, cci int) error {
	cnfport := servers[index].ClusterData[ci].CNFPort
	serviceport := servers[index].ClusterData[ci].ServicePort
	res, err := createSdewanApplication(is.Spec.ServiceName, namespace, is.Spec.AppLabel, serviceport, cnfport)
	if err != nil {
		log.Error("Error creating SDEWAN Application", log.Fields{
			"error":        err,
			"app name":     is.Spec.ServiceName,
			"cluster name": c,
		})
		return pkgerrors.Wrap(err, "Error creating SDEWAN Application")
	}
	servers[index].Clients[ci].ClusterData[cci].Reslist = append(servers[index].Clients[ci].ClusterData[cci].Reslist, res)
	return nil
}

func getClusterKvPair(ctx context.Context, c, kvkey string) (string, error) {

	parts := strings.Split(c, "+")
	if len(parts) != 2 {
		log.Error("Not a valid cluster name", log.Fields{
			"cluster name": c,
		})
		return "", pkgerrors.New("Not a valid cluster name")
	}
	ckv, err := clusterPkg.NewClusterClient().GetAllClusterKvPairs(ctx, parts[0], parts[1])
	var val string
	if err == nil {
		for _, kvp := range ckv {
			for _, mkey := range kvp.Spec.Kv {
				if v, ok := mkey[kvkey]; ok {
					val = fmt.Sprintf("%v", v)
					return val, nil
				}
			}
		}
	}

	return "", pkgerrors.New("Cluster kvpair not found")

}

func getProviderAndCluster(c string) (string, string, error) {
	s := strings.Split(c, "+")
	if len(s) != 2 {
		return "", "", pkgerrors.New("Not a valid cluster name")
	}

	return s[0], s[1], nil
}
