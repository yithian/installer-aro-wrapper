package dnsmasq

import (
	"github.com/coreos/ignition/v2/config/v3_2/types"
	"github.com/openshift/installer/pkg/asset/ignition"
	"github.com/openshift/installer/pkg/asset/ignition/bootstrap"
	"github.com/openshift/installer/pkg/asset/installconfig"
	mcv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	"sigs.k8s.io/yaml"
)

type DNSConfig struct {
	APIIntIP                 string
	IngressIP                string
	GatewayDomains           []string
	GatewayPrivateEndpointIP string
}

func CreatednsmasqIgnitionFiles(bootstrapAsset *bootstrap.Bootstrap, installConfig *installconfig.InstallConfig, dnsConfig DNSConfig) error {
	dnsmasqIgnConfig, err := Ignition3Config(installConfig.Config.ClusterDomain(), dnsConfig.APIIntIP, dnsConfig.IngressIP, dnsConfig.GatewayDomains, dnsConfig.GatewayPrivateEndpointIP, true)
	if err != nil {
		return err
	}
	bootstrapAsset.Config.Storage.Files = ReplaceOrAppend(bootstrapAsset.Config.Storage.Files, dnsmasqIgnConfig.Storage.Files)
	bootstrapAsset.Config.Systemd.Units = ReplaceOrAppendSystemd(bootstrapAsset.Config.Systemd.Units, dnsmasqIgnConfig.Systemd.Units)

	dnsmasqMasterMachineConfig, err := MachineConfig(installConfig.Config.ClusterDomain(), dnsConfig.APIIntIP, dnsConfig.IngressIP, "master", dnsConfig.GatewayDomains, dnsConfig.GatewayPrivateEndpointIP, true)
	if err != nil {
		return err
	}
	dnsmasqWorkerMachineConfig, err := MachineConfig(installConfig.Config.ClusterDomain(), dnsConfig.APIIntIP, dnsConfig.IngressIP, "worker", dnsConfig.GatewayDomains, dnsConfig.GatewayPrivateEndpointIP, true)
	if err != nil {
		return err
	}
	AppendMachineConfigToBootstrap(dnsmasqMasterMachineConfig, bootstrapAsset, "/opt/openshift/openshift/99_openshift-machineconfig_99-master-aro-dns.yaml")
	AppendMachineConfigToBootstrap(dnsmasqWorkerMachineConfig, bootstrapAsset, "/opt/openshift/openshift/99_openshift-machineconfig_99-worker-aro-dns.yaml")
	return nil
}

func AppendMachineConfigToBootstrap(machineConfig *mcv1.MachineConfig, bootstrapAsset *bootstrap.Bootstrap, path string) error {
	data, err := yaml.Marshal(machineConfig)
	if err != nil {
		return err
	}
	config := ignition.FileFromBytes(path, "root", 0644, data)
	bootstrapAsset.Config.Storage.Files = ReplaceOrAppend(bootstrapAsset.Config.Storage.Files, []types.File{config})
	return nil
}

func ReplaceOrAppend(bootstrapFiles []types.File, file []types.File) []types.File {
	for _, iff := range file {
		flag := false
		for i, f := range bootstrapFiles {
			if f.Node.Path == iff.Node.Path {
				bootstrapFiles[i] = iff
				flag = true
				break
			}
		}
		if !flag {
			bootstrapFiles = append(bootstrapFiles, iff)
		}
	}
	return bootstrapFiles
}

func ReplaceOrAppendSystemd(bootstrapFiles []types.Unit, file []types.Unit) []types.Unit {
	for _, iff := range file {
		flag := false
		for i, f := range bootstrapFiles {
			if f.Name == iff.Name {
				bootstrapFiles[i] = iff
				flag = true
				break
			}
		}
		if !flag {
			bootstrapFiles = append(bootstrapFiles, iff)
		}
	}
	return bootstrapFiles
}
