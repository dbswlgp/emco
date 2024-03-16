package module_test

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	readynotifyclient "gitlab.com/project-emco/core/emco-base/src/dcm/pkg/module"
	installappclient "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/grpc/installappclient"
	mockinstallpb "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/mock_installapp"
	mockreadynotifypb "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/grpc/mock_readynotify"
)

var mockinstallapp *mockinstallpb.MockInstallappClient       // for gRPC communication
var mockreadynotify *mockreadynotifypb.MockReadyNotifyClient // for gRPC communication

var buf bytes.Buffer

func TestModule(t *testing.T) {
	RegisterFailHandler(Fail)
	buf.Reset()
	ctrl := gomock.NewController(t)
	mockinstallapp = mockinstallpb.NewMockInstallappClient(ctrl)
	installappclient.Testvars.UseGrpcMock = true
	installappclient.Testvars.InstallClient = mockinstallapp
	mockreadynotify = mockreadynotifypb.NewMockReadyNotifyClient(ctrl)
	readynotifyclient.Testvars.UseGrpcMock = true
	readynotifyclient.Testvars.ReadyNotifyClient = mockreadynotify
	RunSpecs(t, "Module Suite")
	ctrl.Finish()

}
