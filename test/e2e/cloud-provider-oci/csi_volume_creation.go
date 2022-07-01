// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"time"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"

	. "github.com/onsi/ginkgo"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

var _ = Describe("CSI Volume Creation", func() {
	f := framework.NewDefaultFramework("csi-basic")
	Context("[cloudprovider][storage][csi]", func() {
		It("Create PVC and POD for CSI.", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.NewPodForCSI("app1", f.Namespace.Name, pvc.Name, setupF.AdLabel)
		})

		It("Create PVC with VolumeSize 1Gi but should use default 50Gi", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-pvc-with-1gi")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.VolumeFss, scName, nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.NewPodForCSI("app2", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
		})

		It("Create PVC with VolumeSize 100Gi should use 100Gi", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-pvc-with-100gi")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MaxVolumeBlock, scName, nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.NewPodForCSI("app3", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumeCapacity("100Gi", pvc.Name, f.Namespace.Name)
		})

		It("Data should persist on CSI volume on pod restart", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-pod-restart-data-persistence")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.CheckDataPersistenceWithDeployment(pvc.Name, f.Namespace.Name)
		})

		It("FsGroup test for CSI", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-pod-nginx")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false)
			f.CreateCSIDriverOrFail("blockvolume.csi.oraclecloud.com", nil, storagev1.FileFSGroupPolicy)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)

			pvcJig.CheckVolumeDirectoryOwnership(f.Namespace.Name, pvc)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteCSIDriver("blockvolume.csi.oraclecloud.com")
		})
	})
})

var _ = Describe("CSI Volume Expansion iSCSI", func() {
	f := framework.NewDefaultFramework("csi-expansion")
	Context("[cloudprovider][storage][csi][expand][iSCSI]", func() {
		It("Expand PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for iSCSI volumes with existing storage class", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-resizer-pvc-expand-to-100gi-iscsi")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI},
				pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			podName := pvcJig.NewPodForCSI("expanded-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckExpandedVolumeReadWrite(f.Namespace.Name, podName)
			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podName)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCICSIExpand)
		})
	})
})

var _ = Describe("CSI Volume Expansion iSCSI", func() {
	f := framework.NewDefaultFramework("csi-expansion")
	Context("[cloudprovider][storage][csi][expand][iSCSI]", func() {
		It("Expand PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for iSCSI volumes with new storage class", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-resizer-pvc-expand-to-100gi-iscsi")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSIExpand, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI},
				pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			podName := pvcJig.NewPodForCSI("expanded-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckExpandedVolumeReadWrite(f.Namespace.Name, podName)
			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podName)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCICSIExpand)
		})
	})
})

var _ = Describe("CSI Volume Performance Level", func() {
	f := framework.NewBackupFramework("csi-perf-level")
	Context("[cloudprovider][storage][csi][perf][iSCSI]", func() {
		It("Create CSI block volume with Performance Level as Low Cost", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-iscsi-lowcost")

			scName := f.CreateStorageClassOrFail(framework.ClassOCILowCost, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "0"},
				pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			pvcJig.NewPodForCSI("low-cost-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.LowCostPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCILowCost)
		})
		It("Create CSI block volume with no Performance Level and verify default", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-iscsi-default")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			pvcJig.NewPodForCSI("default-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.BalancedPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})
		It("Create CSI block volume with Performance Level as High", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-iscsi-high")

			scName := f.CreateStorageClassOrFail(framework.ClassOCIHigh, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "20"},
				pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			podName := pvcJig.NewPodForCSI("high-perf-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running
			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.HigherPerformanceOption)
			pvcJig.CheckISCSIQueueDepthOnNode(f.Namespace.Name, podName)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCIHigh)
		})
	})
	Context("[cloudprovider][storage][csi][perf][paravirtualized]", func() {
		It("Create CSI block volume with Performance Level as Low Cost", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-paravirtual-lowcost")

			scName := f.CreateStorageClassOrFail(framework.ClassOCILowCost, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "0"},
				pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			pvcJig.NewPodForCSI("low-cost-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.LowCostPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCILowCost)
		})
		It("Create CSI block volume with no Performance Level and verify default", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-paravirtual-balanced")

			scName := f.CreateStorageClassOrFail(framework.ClassOCIBalanced, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "10"},
				pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			pvcJig.NewPodForCSI("default-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.BalancedPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCIBalanced)
		})
		It("Create CSI block volume with Performance Level as High", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-paravirtual-high")

			scName := f.CreateStorageClassOrFail(framework.ClassOCIHigh, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "20"},
				pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			pvcJig.NewPodForCSI("high-perf-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running
			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.HigherPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCIHigh)
		})
	})

	Context("[cloudprovider][storage][csi][perf][static]", func() {
		It("High Performance Static Provisioning CSI", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-static-high")

			scName := f.CreateStorageClassOrFail(framework.ClassOCIHigh, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "20"},
				pvcJig.Labels, "WaitForFirstConsumer", true)

			compartmentId := ""
			if setupF.Compartment1 != "" {
				compartmentId = setupF.Compartment1
			} else if f.CloudProviderConfig.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.CompartmentID
			} else if f.CloudProviderConfig.Auth.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.Auth.CompartmentID
			} else {
				framework.Failf("Compartment Id undefined.")
			}
			pvc, volumeId := pvcJig.CreateAndAwaitStaticPVCOrFailCSI(f.BlockStorageClient, f.Namespace.Name, framework.MinVolumeBlock, csi_util.HigherPerformanceOption, scName, setupF.AdLocation, compartmentId, nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			podName := pvcJig.NewPodForCSI("app4", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
			pvcJig.CheckISCSIQueueDepthOnNode(pvc.Namespace, podName)
			f.VolumeIds = append(f.VolumeIds, volumeId)
			_ = f.DeleteStorageClass(framework.ClassOCIHigh)
		})
	})
})

var _ = Describe("CSI Volume Expansion Paravirtualized", func() {
	f := framework.NewDefaultFramework("csi-expansion")
	Context("[cloudprovider][storage][csi][expand][paravirtualized]", func() {
		It("Expand PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for paravirtualized volumes with new storage class", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-resizer-pvc-expand-to-100gi-paravirtualized")

			scParameter := map[string]string{
				framework.KmsKey:         setupF.CMEKKMSKey,
				framework.AttachmentType: framework.AttachmentTypeParavirtualized,
			}
			scName := f.CreateStorageClassOrFail(framework.ClassOCICSIExpand,
				"blockvolume.csi.oraclecloud.com", scParameter, pvcJig.Labels,
				"WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			podName := pvcJig.NewPodForCSI("expanded-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckExpandedVolumeReadWrite(f.Namespace.Name, podName)
			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podName)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCICSIExpand)
		})
	})
})

var _ = Describe("CSI Static Volume Creation", func() {
	f := framework.NewBackupFramework("csi-static")
	Context("[cloudprovider][storage][csi][static]", func() {
		It("Static Provisioning CSI", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-pvc-with-static")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com",
				nil, pvcJig.Labels, "WaitForFirstConsumer", false)

			compartmentId := ""
			if setupF.Compartment1 != "" {
				compartmentId = setupF.Compartment1
			} else if f.CloudProviderConfig.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.CompartmentID
			} else if f.CloudProviderConfig.Auth.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.Auth.CompartmentID
			} else {
				framework.Failf("Compartment Id undefined.")
			}
			pvc, volumeId := pvcJig.CreateAndAwaitStaticPVCOrFailCSI(f.BlockStorageClient, f.Namespace.Name, framework.MinVolumeBlock, 10, scName, setupF.AdLocation, compartmentId, nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.NewPodForCSI("app4", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
			f.VolumeIds = append(f.VolumeIds, volumeId)
		})
	})
})

var _ = Describe("CSI CMEK,PV attachment and in-transit encryption test", func() {
	f := framework.NewDefaultFramework("csi-basic")
	Context("[cloudprovider][storage][csi][cmek][paravirtualized]", func() {
		It("Create PVC and POD for CSI with CMEK,PV attachment and in-transit encryption", func() {
			TestCMEKAttachmentTypeAndEncryptionType(f, framework.AttachmentTypeParavirtualized)
		})
	})

	Context("[cloudprovider][storage][csi][cmek][iscsi]", func() {
		It("Create PVC and POD for CSI with CMEK,ISCSI attachment and in-transit encryption", func() {
			TestCMEKAttachmentTypeAndEncryptionType(f, framework.AttachmentTypeISCSI)
		})
	})

})

func TestCMEKAttachmentTypeAndEncryptionType(f *framework.CloudProviderFramework, expectedAttachmentType string) {
	pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-cmek-iscsi-in-transit-e2e-tests")
	scParameter := map[string]string{
		framework.KmsKey:         setupF.CMEKKMSKey,
		framework.AttachmentType: expectedAttachmentType,
	}
	scName := f.CreateStorageClassOrFail(framework.ClassOCIKMS, "blockvolume.csi.oraclecloud.com", scParameter, pvcJig.Labels, "WaitForFirstConsumer", false)
	pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
	podName := pvcJig.NewPodForCSI("app1", f.Namespace.Name, pvc.Name, setupF.AdLabel)
	pvcJig.CheckCMEKKey(f.Client.BlockStorage(), pvc.Name, f.Namespace.Name, setupF.CMEKKMSKey)
	pvcJig.CheckAttachmentTypeAndEncryptionType(f.Client.Compute(), pvc.Name, f.Namespace.Name, podName, expectedAttachmentType)
	f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
	_ = f.DeleteStorageClass(framework.ClassOCIKMS)
}

var _ = Describe("CSI backup policy addition tests", func() {
	f := framework.NewBackupFramework("csi-basic")
	Context("[cloudprovider][storage][csi][backup-policy]", func() {
		It("can assign an Oracle-defined backup policy", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-backup-policy-oracle-first")

			backupPolicies := pvcJig.GetOracleDefinedBackupPolicies(f.BlockStorageClient)
			scName := f.CreateStorageClassOrFail("backup-policy-first", "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.BackupPolicyId: backupPolicies[0]},
				pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			pvcJig.NewPodForCSI("backup-policy-check-first-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			// time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckBackupPolicy(f.BlockStorageClient, pvc.Namespace, pvc.Name, backupPolicies[0])
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass("backup-policy-first")
		})

		It("creates a User-Defined Backup Policy", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "checking-user-defined-policy")

			compartmentId := ""
			if setupF.Compartment1 != "" {
				compartmentId = setupF.Compartment1
			} else if f.CloudProviderConfig.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.CompartmentID
			} else if f.CloudProviderConfig.Auth.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.Auth.CompartmentID
			} else {
				framework.Failf("Compartment Id undefined.")
			}

			backupPolicyOcid := pvcJig.CreateUserDefinedBackupPolicy(f.BlockStorageClient, "test-policy", compartmentId)
			framework.Logf("BackupPolicyID : %s", backupPolicyOcid)

			scName := f.CreateStorageClassOrFail("backup-policy-user-defined", "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.BackupPolicyId: backupPolicyOcid},
				pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			pvcJig.NewPodForCSI("backup-policy-check-ud-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			// time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckBackupPolicy(f.BlockStorageClient, pvc.Namespace, pvc.Name, backupPolicyOcid)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass("backup-policy-user-defined")
			pvcJig.DeleteUserDefinedBackupPolicy(f.BlockStorageClient, backupPolicyOcid)
		})

		It("assigns no backup policy when the parameter is not provided", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-backup-policy-none")

			scName := f.CreateStorageClassOrFail("backup-policy-none", "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			pvcJig.NewPodForCSI("backup-policy-check-none-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			// time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckBackupPolicy(f.BlockStorageClient, pvc.Namespace, pvc.Name, "")
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass("backup-policy-none")
		})

		It("assigns no backup policy when the provided id is an empty string", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-backup-policy-empty")

			scName := f.CreateStorageClassOrFail("backup-policy-empty", "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.BackupPolicyId: ""},
				pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			pvcJig.NewPodForCSI("backup-policy-check-empty-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			// time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckBackupPolicy(f.BlockStorageClient, pvc.Namespace, pvc.Name, "")
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass("backup-policy-empty")
		})

		It("allows volume expansion after policy assignment", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-backup-policy-expand")

			backupPolicies := pvcJig.GetOracleDefinedBackupPolicies(f.BlockStorageClient)
			scName := f.CreateStorageClassOrFail("backup-policy-expand", "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.BackupPolicyId: backupPolicies[0]},
				pvcJig.Labels, "WaitForFirstConsumer", true)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)
			podName := pvcJig.NewPodForCSI("backup-policy-check-expand-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			// time.Sleep(60 * time.Second) //waiting for pod to up and running
			pvcJig.CheckBackupPolicy(f.BlockStorageClient, pvc.Namespace, pvc.Name, backupPolicies[0])

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckExpandedVolumeReadWrite(f.Namespace.Name, podName)
			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podName)
			pvcJig.CheckBackupPolicy(f.BlockStorageClient, pvc.Namespace, pvc.Name, backupPolicies[0])
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass("backup-policy-expand")
		})
	})
})

var _ = Describe("CSI raw block volume tests", func() {
	f := framework.NewBackupFramework("csi-raw")

	Context("[cloudprovider][storage][csi][raw-block]", func() {
		It("can publish a raw block volume and expand its size [iscsi]", func() {
			TestRawBlockProvisionAndExpansion(f, framework.AttachmentTypeISCSI)
		})

		It("can publish a raw block volume and expand its size [paravirtualized]", func() {
			TestRawBlockProvisionAndExpansion(f, framework.AttachmentTypeParavirtualized)
		})
	})
})

func TestRawBlockProvisionAndExpansion(f *framework.CloudProviderFramework, attachmentType string) {
	pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-raw-block-expand")

	By("creating a Custom Storage Class")
	scName := f.CreateStorageClassOrFail("raw-sc", "blockvolume.csi.oraclecloud.com",
		map[string]string{framework.AttachmentType: attachmentType},
		pvcJig.Labels, "WaitForFirstConsumer", true)

	By("creating a PVC with VolumeMode - Block")
	tweakFunc := func(pvc *v1.PersistentVolumeClaim) {
		blockMode := v1.PersistentVolumeBlock
		pvc.Spec.VolumeMode = &blockMode
	}
	pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, tweakFunc)

	By("creating a new pod")
	podName := pvcJig.NewPodForRawCSI("raw-check-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

	By("expanding the size in PVC")
	expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, "60Gi", nil)
	time.Sleep(120 * time.Second) // Wait for volume rescan

	By("checking the expanded size and already written data")
	pvcJig.CheckVolumeCapacity("60Gi", expandedPvc.Name, f.Namespace.Name)
	pvcJig.CheckExpandedSizeAndDataForRawVolume(f.Namespace.Name, podName)
	f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
	_ = f.DeleteStorageClass("raw-sc")
}
