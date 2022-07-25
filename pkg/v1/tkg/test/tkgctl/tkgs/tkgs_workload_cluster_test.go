// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package tkgs

import (
	"bytes"
	"fmt"
	"io"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/cluster-api/util"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

const TKC_KIND = "kind: TanzuKubernetesCluster"

var _ = Describe("TKGS - Create workload cluster use cases", func() {
	var (
		stdoutOld *os.File
		r         *os.File
		w         *os.File
	)
	JustBeforeEach(func() {
		err = tkgctlClient.CreateCluster(clusterOptions)
	})

	BeforeEach(func() {
		tkgctlClient, err = tkgctl.New(tkgctlOptions)
		Expect(err).To(BeNil())
		deleteClusterOptions = getDeleteClustersOptions(e2eConfig)
	})

	Context("when input file is legacy config file (TKC cluster)", func() {
		BeforeEach(func() {
			Expect(e2eConfig.TkrVersion).ToNot(BeEmpty(), fmt.Sprintf("the kubernetes_version should not be empty to create legacy TKGS cluster"))
			clusterOptions.TkrVersion = e2eConfig.TkrVersion
			e2eConfig.WorkloadClusterOptions.ClusterName = "tkc-e2e-" + util.RandomString(4)
			deleteClusterOptions.ClusterName = e2eConfig.WorkloadClusterOptions.ClusterName
			clusterOptions.ClusterConfigFile = createClusterConfigFile(e2eConfig)
		})
		AfterEach(func() {
			defer os.Remove(clusterOptions.ClusterConfigFile)
		})
		Context("when cluster Plan is dev", func() {
			BeforeEach(func() {
				e2eConfig.WorkloadClusterOptions.ClusterPlan = "dev"
			})
			When("create cluster is invoked", func() {
				AfterEach(func() {
					err = tkgctlClient.DeleteCluster(deleteClusterOptions)
				})
				It("should create TKC Workload Cluster and delete it", func() {
					Expect(err).To(BeNil())
				})
			})

			When("create cluster dry-run is invoked", func() {
				BeforeEach(func() {
					// set dry-run mode
					clusterOptions.GenerateOnly = true

					stdoutOld = os.Stdout
					r, w, _ = os.Pipe()
					os.Stdout = w
				})
				It("should give TKC configuration as output", func() {
					Expect(err).ToNot(HaveOccurred())

					w.Close()
					os.Stdout = stdoutOld
					var buf bytes.Buffer
					io.Copy(&buf, r)
					r.Close()
					str := buf.String()
					Expect(str).To(ContainSubstring(TKC_KIND))
					Expect(str).To(ContainSubstring("name: " + clusterOptions.ClusterName))
				})
			})
		})
		Context("when cluster Plan is prod", func() {
			BeforeEach(func() {
				clusterOptions.Plan = "prod"
				clusterOptions.GenerateOnly = false
			})

			When("create cluster is invoked", func() {
				AfterEach(func() {
					err = tkgctlClient.DeleteCluster(deleteClusterOptions)
				})
				It("should create TKC Workload Cluster and delete it", func() {
					Expect(err).To(BeNil())
				})
			})
		})
	})

	Context("when input file is cluster class based", func() {
		BeforeEach(func() {
			clusterName, namespace := ValidateClusterClassConfigFile(e2eConfig.WorkloadClusterOptions.ClusterClassFilePath)
			clusterOptions.ClusterConfigFile = e2eConfig.WorkloadClusterOptions.ClusterClassFilePath
			clusterOptions.ClusterName = clusterName
			clusterOptions.Namespace = namespace
		})

		It("should fail to create a clusterclass based cluster", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(constants.ErrorMsgFeatureGateNotActivated, constants.ClusterClassFeature, clusterOptions.Namespace)))
		})
	})
})

// createLegacyClusterTest creates and deletes (if created successfully) workload cluster
func createLegacyClusterTest(tkgctlClient tkgctl.TKGClient, deleteClusterOptions tkgctl.DeleteClustersOptions, cliFlag bool, clusterName, namespace string) {
	if isTKCAPIFeatureActivated {
		By(fmt.Sprintf("creating TKC workload cluster, TKC-API feature-gate is activated and cli feature flag set %v", cliFlag))
		By(fmt.Sprintf("creating TKC workload cluster %v in namespace: %v, cli feature flag is %v", clusterName, namespace, cliFlag))
		err = tkgctlClient.CreateCluster(clusterOptions)
		Expect(err).To(BeNil())
		By(fmt.Sprintf("deleting TKC workload cluster %v in namespace: %v", clusterName, namespace))
		err = tkgctlClient.DeleteCluster(deleteClusterOptions)
		Expect(err).To(BeNil())

	} else {
		By(fmt.Sprintf("creating TKC workload cluster, TKC-API feature-gate is deactivated and cli feature flag set %v", cliFlag))
		err = tkgctlClient.CreateCluster(clusterOptions)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(constants.ErrorMsgFeatureGateNotActivated, constants.TKCAPIFeature, constants.TKGSTKCAPINamespace)))
	}
}
