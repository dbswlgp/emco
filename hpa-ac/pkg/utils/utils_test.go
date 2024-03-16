package utils_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/project-emco/core/emco-base/src/hpa-ac/pkg/utils"
	orchLog "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"github.com/sirupsen/logrus"
)

func TestUtils(t *testing.T) {

	fmt.Printf("\n================== TestUtils .. start ==================\n")

	orchLog.SetLoglevel(logrus.InfoLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "ActionUtils")

	fmt.Printf("\n================== TestUtils .. end ==================\n")
}

var _ = Describe("ActionUtils", func() {
	It("unsuccessful DecodeYAMLData", func() {
		_, err := utils.DecodeYAMLData("", nil)
		Expect(err).To(HaveOccurred())
	})

	It("unsuccessful DecodeYAMLFile", func() {
		_, err := utils.DecodeYAMLFile("", nil)
		Expect(err).To(HaveOccurred())
	})
})
