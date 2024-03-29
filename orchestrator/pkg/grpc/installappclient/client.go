// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package installappclient

import (
	"context"
	"sync"
	"time"
	"fmt"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"
	installpb "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/installapp"
	//"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/state"
//	orchmodule "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
	//"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module"
)

const rsyncName = "rsync"

/*
RsyncInfo consists of rsyncName, hostName and portNumber.
*/
type RsyncInfo struct {
	RsyncName  string
	hostName   string
	portNumber int
}

var rsyncInfo RsyncInfo
var mutex = &sync.Mutex{}

type _testvars struct {
	UseGrpcMock   bool
	InstallClient installpb.InstallappClient
}

var Testvars _testvars

// InitRsyncClient initializes connctions to the Resource Synchronizer service
func InitRsyncClient() bool {
	if (RsyncInfo{}) == rsyncInfo {
		mutex.Lock()
		defer mutex.Unlock()
		log.Error("RsyncInfo not set. InitRsyncClient failed", log.Fields{
			"Rsyncname":  rsyncInfo.RsyncName,
			"Hostname":   rsyncInfo.hostName,
			"PortNumber": rsyncInfo.portNumber,
		})
		return false
	}
	rpc.UpdateRpcConn(rsyncInfo.RsyncName, rsyncInfo.hostName, rsyncInfo.portNumber)
	return true
}

// NewRsyncInfo shall return a newly created RsyncInfo object
func NewRsyncInfo(rName, h string, pN int) RsyncInfo {
	mutex.Lock()
	defer mutex.Unlock()
	rsyncInfo = RsyncInfo{RsyncName: rName, hostName: h, portNumber: pN}
	return rsyncInfo

}

// InvokeInstallApp will make the grpc call to the resource synchronizer
// or rsync controller.
// rsync will deploy the resources in the app context to the clusters as
// prepared in the app context.
func InvokeInstallApp(ctx context.Context, appContextId string) error {
	var err error
	var rpcClient installpb.InstallappClient
	var installRes *installpb.InstallAppResponse

	ctx, cancel := context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	// Unit test helper code
	if Testvars.UseGrpcMock {
		rpcClient = Testvars.InstallClient
		installReq := new(installpb.InstallAppRequest)
		installReq.AppContext = appContextId
		installRes, err = rpcClient.InstallApp(ctx, installReq)
		if err == nil {
			log.Info("Response from InstappApp GRPC call", log.Fields{
				"Succeeded": installRes.AppContextInstalled,
				"Message":   installRes.AppContextInstallMessage,
			})
		}
		return nil
	}

	conn := rpc.GetRpcConn(ctx, rsyncName)
	if conn == nil {
		InitRsyncClient()
		conn = rpc.GetRpcConn(ctx, rsyncName)
	}

	if conn != nil {
		rpcClient = installpb.NewInstallappClient(conn)
		installReq := new(installpb.InstallAppRequest)
		installReq.AppContext = appContextId
		installRes, err = rpcClient.InstallApp(ctx, installReq)
		if err == nil {
			log.Info("Response from InstappApp GRPC call", log.Fields{
				"Succeeded": installRes.AppContextInstalled,
				"Message":   installRes.AppContextInstallMessage,
			})
		}
	} else {
		return pkgerrors.Errorf("InstallApp Failed - Could not get InstallAppClient: %v", "rsync")
	}

	if err == nil {
		if installRes.AppContextInstalled {
			log.Info("InstallApp Success", log.Fields{
				"AppContext": appContextId,
				"Message":    installRes.AppContextInstallMessage,
			})
			return nil
		} else {
			return pkgerrors.Errorf("InstallApp Failed: %v", installRes.AppContextInstallMessage)
		}
	}
	return err
}

func InvokeUninstallApp(ctx context.Context, appContextId string) error {
	var err error
	var rpcClient installpb.InstallappClient
	var uninstallRes *installpb.UninstallAppResponse

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn := rpc.GetRpcConn(ctx, rsyncName)
	if conn == nil {
		InitRsyncClient()
		conn = rpc.GetRpcConn(ctx, rsyncName)
	}

	if conn != nil {
		rpcClient = installpb.NewInstallappClient(conn)
		uninstallReq := new(installpb.UninstallAppRequest)
		uninstallReq.AppContext = appContextId
		uninstallRes, err = rpcClient.UninstallApp(ctx, uninstallReq)
		if err == nil {
			log.Info("Response from UninstappApp GRPC call", log.Fields{
				"Succeeded": uninstallRes.AppContextUninstalled,
				"Message":   uninstallRes.AppContextUninstallMessage,
			})
		}
	} else {
		return pkgerrors.Errorf("UninstallApp Failed - Could not get InstallAppClient: %v", "rsync")
	}

	if err == nil {
		if uninstallRes.AppContextUninstalled {
			log.Info("UninstallApp Success", log.Fields{
				"AppContext": appContextId,
				"Message":    uninstallRes.AppContextUninstallMessage,

			})

			getappcontextfromid, _ := state.GetAppContextFromId(ctx, appContextId)
			fmt.Println("\n",getappcontextfromid,"\n")
/*			err = getappcontextfromid.DeleteCompositeApp(ctx)
			if err != nil{
				fmt.Println("\n\n DeleteCompositeApp Failed\n\n")
			}
*/
			return nil


		} else {
			return pkgerrors.Errorf("UninstallApp Failed: %v", uninstallRes.AppContextUninstallMessage)
		}
	}


	return err
}
