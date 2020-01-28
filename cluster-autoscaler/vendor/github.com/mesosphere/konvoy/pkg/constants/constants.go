package constants

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/blang/semver"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/mesosphere/konvoy/pkg/printer"
)

var (
	KonvoyVersion   string
	KonvoyBuildDate string

	DefaultKubernetesVersion string
	// NOTE: This must be kept up to date with value of `CONTAINERD_VERSION` in Makefile.
	// This is specified here so that it can be used by KCL directly without setting
	// linker flags.
	DefaultContainerdVersion       = "1.2.6"
	DefaultKubeAddonsConfigVersion string

	MaxNumWorkerPoolsString string
)

const (
	KubernetesVersion            = "1.16.4"
	KubernetesVersionRangeString = ">=1.15.0 <1.17.0"

	MinimumRequiredDockerVersion = "18.09.2"

	InventoryFileVersion = "v1beta1"

	CalicoEncapsulationModeIPIP          = "ipip"
	CalicoEncapsulationModeVXLAN         = "vxlan"
	CalicoEncapsulationModeIPIPOverhead  = 20
	CalicoEncapsulationModeVXLANOverhead = 50
	DefaultCalicoVersion                 = "v3.10.1"
	DefaultMTUProvisionerAWS             = 1500
	DefaultMTUProvisionerAzure           = 1500
	DefaultMTUProvisionerNone            = 1500
	DefaultMTUProvisionerDocker          = 1500

	DefaultClusterName = "kubernetes"

	DefaultPodSubnet                    = "192.168.0.0/16"
	DefaultServiceSubnet                = "10.0.0.0/18"
	DefaultControlPlaneEndpointOverride = ""

	DefaultEnableAdditionalRepositories = true

	DefaultIngressServiceName  = "traefik-kubeaddons"
	DefaultOpsPortalSecretName = "ops-portal-credentials"

	DefaultMaxNumWorkerPools = 10

	DefaultInventoryFileName                   = "inventory.yaml"
	DefaultClusterConfigurationFileName        = "cluster.yaml"
	MutatedDefaultClusterConfigurationFileName = "cluster.tmp.yaml"
	DefaultImagesFileName                      = "images.json"

	// The name of the kubeconfig file that gets copied from a control-plane node
	AdminConfFileName = "admin.conf"

	PrefixProvisionerVariableFilename           = "variables.*.tfvars.json"
	DefaultProvisionerPlanFilename              = "konvoy-tf-plan.out"
	DefaultTerraformStateFilename               = "terraform.tfstate"
	DefaultTerraformParallelism                 = 7
	DefaultAWSControlPlaneCount                 = 3
	DefaultAWSWorkerCount                       = 4
	DefaultAWSRegion                            = "us-west-2"
	DefaultAWSSSHUser                           = "centos"
	DefaultAWSWorkerMachineType                 = "m5.2xlarge"
	DefaultAWSControlPlaneMachineType           = "m5.xlarge"
	DefaultAWSControlPlaneMachineRootVolumeType = "io1"
	DefaultAWSControlPlaneMachineRootVolumeIOPS = int32(1000)
	DefaultAWSBastionMachineType                = "m5.large"
	DefaultAWSMachineRootVolumeType             = "gp2"
	DefaultAWSMachineRootVolumeSize             = 80
	DefaultAWSBastionMachineRootVolumeSize      = 10
	DefaultAWSMachineImagefsVolumeType          = "gp2"
	DefaultAWSMachineImagefsVolumeSize          = 160
	DefaultAWSMachineImagefsVolumeDevice        = "xvdb"

	DefaultAzurePodSubnet                    = "10.0.128.0/18"
	DefaultAzureControlPlaneCount            = 3
	DefaultAzureWorkerCount                  = 6
	DefaultAzureLocation                     = "westus"
	DefaultAzureFaultDomainCount             = 3
	DefaultAzureUpdateDomainCount            = 3
	DefaultAzureSSHUser                      = "centos"
	DefaultAzureWorkerMachineType            = "Standard_DS3_v2"
	DefaultAzureControlPlaneMachineType      = "Standard_DS3_v2"
	DefaultAzureBastionMachineType           = "Standard_DS2_v2"
	DefaultAzureMachineRootVolumeType        = "Standard_LRS"
	DefaultAzureMachineRootVolumeSize        = 80
	DefaultAzureBastionMachineRootVolumeSize = 40
	DefaultAzureMachineImagefsVolumeType     = "Standard_LRS"
	DefaultAzureMachineImagefsVolumeSize     = 160

	// NOTE: maxAWSClusterNameLen is the max clusterName length
	MaxAWSClusterNameLen = 32
	// NOTE: MinAWSClusterNameLen, a cluster name shorter than 3 characters is not meaningful
	MinAWSClusterNameLen = 3

	DefaultDockerPodSubnet                     = "10.254.0.0/16"
	DefaultDockerServiceSubnet                 = "10.255.0.0/16"
	DefaultDockerControlPlaneEndpointOverride  = "172.17.1.251:6443"
	DefaultDockerControlPlaneCount             = 1
	DefaultDockerWorkerCount                   = 2
	DefaultDockerBaseImage                     = "mesosphere/konvoy-base-centos7"
	DefaultDockerSSHUser                       = "root"
	DefaultDockerControlPlaneMappedPortBase    = 46000
	DefaultDockerSSHControlPlaneMappedPortBase = 22000
	DefaultDockerSSHWorkerMappedPortBase       = 22010
	DefaultDockerDedicatedNetwork              = false
	DefaultDockerKubeletCgroupRoot             = "/kubelet"

	// Printout messages
	SuccessDeployMsg                    = "Kubernetes cluster and addons deployed successfully!"
	SuccessDeployKubernetesMsg          = "Kubernetes cluster deployed successfully!"
	SuccessDeployContainerNetworkingMsg = "Container networking deployed successfully!"
	SuccessDeployAddonControllerMsg     = "Addon Controller Deployed Successfully!"
	SuccessDeployAddonsMsg              = "Addons deployed successfully!"

	PromptToContinue                  = ", do you want to continue [y/n]: "
	EstimatedTime                     = "This process will take about %d %s to complete (additional time may be required for larger clusters)"
	PromptToContinueWithEstimatedTime = EstimatedTime + PromptToContinue

	// Estimated times(worker #) 6		20		40
	TimeUp                        = 15 // 13 	19  	23.5
	TimeProvision                 = 3  // 3		6.5 	11
	TimeDeployKubernetes          = 7  // this also includes prepare
	TimeDeployContainerNetworking = 1
	TimeDeployAddons              = 3
	TimeDeploy                    = 7  // 7		7.5 	9.5
	TimeUpgrade                   = 20 // 20		40 		60
	TimeDown                      = 3
	TimeReset                     = 5

	MarkerFile = "/etc/konvoy-marker.yaml"

	NodeLabel = "konvoy.mesosphere.com/inventory_hostname"

	KonvoyRepoName                  = "mesosphere/konvoy"
	DefaultDockerRegistryURL        = "https://registry-1.docker.io"
	DefaultDockerRegistryScheme     = "https"
	DefaultKonvoyCLIVersionFileName = "cli_version"

	DefaultAirgappedHelmRepo = "http://base-addons-chart-repo:8879"
)

// executer consts
const (
	PrepareName                       = "Preparing Machines"
	PreparePlaybook                   = "prepare.yaml"
	PreflightsName                    = "Running Preflights"
	PreflightsPlaybook                = "preflights.yaml"
	PreflightsBasicPlaybook           = "preflights-basic.yaml"
	DeployKubernetesName              = "Deploying Kubernetes"
	DeployKubernetesPlaybook          = "deploy-kubernetes.yaml"
	CheckNodesName                    = "Checking Nodes"
	CheckNodesPlaybook                = "check-nodes.yaml"
	CheckKubernetesName               = "Checking Kubernetes"
	CheckKubernetesPlaybook           = "check-kubernetes.yaml"
	FetchKubeconfigName               = "Fetching Admin Kubeconfig"
	FetchKubeconfigPlaybook           = "fetch-kubeconfig.yaml"
	FetchNodeConfigurationName        = "Fetching Node Configuration"
	FetchNodeConfigurationPlaybook    = "fetch-node-configuration.yaml"
	FetchNodeConfigurationVar         = "fetch_node_configuration_local_dir"
	ResetName                         = "Resetting Machines"
	ResetPlaybook                     = "reset.yaml"
	UpgradeKubernetesName             = "Upgrading Kubernetes"
	UpgradeKubernetesPlaybook         = "upgrade-kubernetes.yaml"
	DeployNodeLabelsAndTaints         = "Adding Node Labels and Taints"
	DeployNodeLabelsAndTaintsPlaybook = "deploy-node-labels-taints.yaml"
	DeployAdditionalResources         = "Deploying Additional Kubernetes Resources"
	DeployContainerNetworkingName     = "Deploying Container Networking"
	DeployContainerNetworkingPlaybook = "deploy-container-networking.yaml"
	DeployAddonControllerName         = "Deploying Addon Controller"
	DeployAddonControllerPlaybook     = "deploy-addon-controller.yaml"
	DiagnoseName                      = "Diagnosing Cluster"
	DiagnosePlaybook                  = "diagnose.yaml"
)

// up and provision cloud-providers
const (
	DefaultProvisionerProvider = "aws"
	ProvisionerAWS             = "aws"
	ProvisionerAzure           = "azure"
	ProvisionerDocker          = "docker"
	ProvisionerNone            = "none"
)

const (
	KonvoyCommandArgEnv    = "KONVOY_COMMAND_ARG"
	KonvoyExecutableDirEnv = "KONVOY_EXECUTABLE_DIR"
	KonvoyUserOSEnv        = "KONVOY_USER_OS"
)

var (
	SupportedProviders = []string{ProvisionerAWS, ProvisionerAzure, ProvisionerDocker, ProvisionerNone}

	DefaultEnabledAdmissionPlugins = []string{"NodeRestriction", "AlwaysPullImages"}
)

var (
	DefaultAWSAvailabilityZones = []string{"us-west-2c"}
	DefaultAWSAdminCIDRBlocks   = []string{"0.0.0.0/0"}
)

var (
	KubernetesVersionRange           = semver.MustParseRange(KubernetesVersionRangeString)
	MaxNumWorkerPools                = maxNumWorkerPools()
	AnsibleDir                       = ansibleDir()
	TerraformPath                    = terraformPath()
	ProvidersDir                     = providersDir()
	ProvisionerExtrasDir             = filepath.Join(WorkingDir, "extras", "provisioner")
	KubernetesExtrasDir              = filepath.Join(WorkingDir, "extras", "kubernetes")
	ExecutableDir                    = executableDir()
	WrapperExecutableDir             = wrapperExecutableDir()
	WorkingDir                       = workingDir()
	RunsDir                          = filepath.Join(WorkingDir, "runs/")
	PythonPath                       = pythonPath()
	RPMsTarFileName                  = "konvoy_%s_x86_64_rpms.tar.gz"
	DebsTarFileName                  = "konvoy_%s_amd64_debs.tar.gz"
	DefaultRPMsTarFile               = filepath.Join(WrapperExecutableDir, fmt.Sprintf(RPMsTarFileName, KonvoyVersion))
	DefaultDebsTarFile               = filepath.Join(WrapperExecutableDir, fmt.Sprintf(DebsTarFileName, KonvoyVersion))
	DefaultImagesFilePath            = filepath.Join(WrapperExecutableDir, DefaultImagesFileName)
	User                             = currentUser()
	HomeDir                          = homeDir()
	KubeconfigPath                   = filepath.Join(WorkingDir, AdminConfFileName)
	DefaultStateDir                  = filepath.Join(WorkingDir, "state/")
	DefaultKonvoyOptionsDir          = filepath.Join(WorkingDir, ".konvoy/")
	UserOS                           = userOS()
	InstructionRunApplyKubeconfigMsg = fmt.Sprintf("Run `%s apply kubeconfig` to update kubectl credentials.", executable())
)

const (
	AnsibleCallbackWhiteListVerbose = "profile_tasks"
)

func maxNumWorkerPools() int32 {
	num, err := strconv.Atoi(MaxNumWorkerPoolsString)
	if err != nil {
		printer.PrintColor(os.Stderr, printer.Red, "Warning: max number worker pools not configured, default to %d", DefaultMaxNumWorkerPools)
		num = DefaultMaxNumWorkerPools
	}
	return int32(num)
}

// executable returns the string with the executable name. If KONVOY_COMMAND_ARG is set,
// it's set by the konvoy bash script when running this binary in a container. This allows
// this function to return the command name used when running the konvoy script.
//
// TODO return error but there is no way to recover
func executable() string {
	if ex := os.Getenv(KonvoyCommandArgEnv); ex != "" {
		return ex
	}
	ex, err := os.Executable()
	if err != nil {
		printer.PrintColor(os.Stderr, printer.Red, "Error: could not get executable %v", err)
		os.Exit(1)
	}
	return ex
}

// TODO return error but there is no way to recover
func executableDir() string {
	ex, err := os.Executable()
	if err != nil {
		printer.PrintColor(os.Stderr, printer.Red, "Error: could not get executable directory %v", err)
		os.Exit(1)
	}
	return filepath.Dir(ex)
}

func wrapperExecutableDir() string {
	if ex := os.Getenv(KonvoyExecutableDirEnv); ex != "" {
		return ex
	}
	return executableDir()
}

func workingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		printer.PrintColor(os.Stderr, printer.Red, "Error: could not get current working directory %v", err)
		os.Exit(1)
	}
	return wd
}

func pythonPath() string {
	pythonVersion, err := pythonVersion()
	if err != nil {
		return ""
	}
	return filepath.Join(AnsibleDir, "lib", "python"+pythonVersion[:3], "site-packages")
}

func pythonVersion() (string, error) {
	pythonBinary, err := exec.LookPath("/usr/bin/python")
	if err != nil {
		return "", err
	}
	out, err := exec.Command(pythonBinary, "-V").CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.Split(string(out), " ")[1], nil
}

func currentUser() string {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	if user.Name == "" {
		// It could be blank, fallback to username
		return user.Username
	}
	return user.Name
}

func homeDir() string {
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	return homeDir
}

func ansibleDir() string {
	ap := os.Getenv("ANSIBLE_PATH")
	if ap == "" {
		ap = filepath.Join(ExecutableDir, "ansible")
	}

	return ap
}

func terraformPath() string {
	tp := os.Getenv("TERRAFORM_PATH")
	if tp == "" {
		tp = filepath.Join(ExecutableDir, "terraform")
		// Prefer vendored terraform if exists, otherwise assume its in PATH
		_, err := exec.LookPath(tp)
		if err != nil {
			tp = "terraform"
		}
	}

	return tp
}

func providersDir() string {
	ps := os.Getenv("PROVIDERS_PATH")
	if ps == "" {
		ps = filepath.Join(ExecutableDir, "providers")
	}

	return ps
}

func userOS() string {
	override := os.Getenv(KonvoyUserOSEnv)
	if override != "" {
		return override
	}
	return runtime.GOOS
}
