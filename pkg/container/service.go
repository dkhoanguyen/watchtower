package container

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/network"
)

type ShellCommand []string
type Labels map[string]string

type Image struct {
	ID      string
	Name    string
	Tag     string
	Created string
}

type Volume struct {
	Name string `json:"name"`
}

type Network struct {
	Name           string `json:"name"`
	ID             string
	CheckDuplicate bool
	Labels         Labels
	Internal       bool
	Attachable     bool
	Driver         string `json:"driver"`
	Ipam           network.IPAM
	EnableIPv6     bool
}

type Service struct {
	Name          string            `json:"name"`
	Action        string            `json:"action"`
	Hostname      string            `json:"hostname"`
	User          string            `json:"user"`
	CapAdd        []string          `json:"cap_add"`
	CapDrop       []string          `json:"cap_drop"`
	BuildOpt      ServiceBuild      `json:"build_opt"`
	CgroupParent  string            `json:"cgroup_parent"`
	Command       ShellCommand      `json:"command"`
	ContainerName string            `json:"container_name"`
	Domainname    string            `json:"domain_name"`
	DependsOn     []string          `json:"depends_on"`
	Devices       []string          `json:"devices"`
	EntryPoint    ShellCommand      `json:"entrypoint"`
	Environment   []string          `json:"environment"`
	EnvFile       []string          `json:"env_file"`
	Expose        []string          `json:"expose"`
	ExtraHosts    []string          `json:"extra_hosts"`
	IpcMode       string            `json:"ipc_mode"`
	Resources     ServiceResources  `json:"resources"`
	Networks      []ServiceNetwork  `json:"networks"`
	NetworkMode   string            `json:"network_mode"`
	Ports         []ServicePort     `json:"ports"`
	Privileged    bool              `json:"privileged"`
	Sysctls       map[string]string `json:"sysctls"`
	Restart       string            `json:"restart"`
	Tty           bool              `json:"tty"`
	Volumes       []ServiceVolume   `json:"volumes"`
	WorkingDir    string            `json:"working_dir"`
	Image         string            `json:"image"`
}

type ServiceBuild struct {
	Context    string
	Dockerfile string
	Args       map[string]*string
}

type ServiceNetwork struct {
	Name    string
	Aliases []string
	IPv4    string
	IPv6    string
}

type ServiceVolume struct {
	Type        string
	Source      string
	Destination string
	Option      string
}

type ServicePort struct {
	Target   string
	Protocol string
	HostIp   string
	HostPort string
}

type ServiceResources struct {
	CPUPeriod         int64
	CPUQuota          int64
	CpusetCpus        string
	CpusetMems        string
	MemoryLimit       int64
	MemoryReservation int64
	MemorySwap        int64
	MemorySwappiness  int64
	OomKillDisable    bool
}

const (
	VolumeTypeBind  = "bind"
	VolumeTypeMount = "mount"
)

const (
	RestartAlways        = "always"
	RestartOnFailure     = "on-failure"
	RestartNoRetry       = "no"
	RestartUnlessStopped = "unless-stopped"
)

const (
	ActionRun     = "start"
	ActionPause   = "pause"
	ActionStop    = "stop"
	ActionRestart = "restart"
)

func GetBuildConfig(rawBuildConfig map[string]interface{}) ServiceBuild {
	return ServiceBuild{
		Context:    rawBuildConfig["context"].(string),
		Dockerfile: rawBuildConfig["dockerfile"].(string),
	}
}

func MakeService(
	config map[string]interface{},
	name string) Service {

	output := Service{}

	// Service Name
	output.Name = name

	// COmpose file from a to z (begining to end)
	// Build options
	// output.BuildOpt = MakeBuildOpt(config, path)
	// Image
	output.Image = MakeImage(config)

	// Action
	output.Action = MakeAction(config)

	// Container Name
	output.ContainerName = MakeContainerName(config)

	// Commands
	output.Command = MakeCommand(config, "command")

	// Dependencies
	output.DependsOn = MakeDependsOn(config)

	// Deployment and Resources
	output.Resources = MakeDeployResources(config)

	// Entrypoint
	output.EntryPoint = MakeCommand(config, "entrypoint")

	// Environment Variables
	output.Environment = MakeEnviroment(config)

	// Network
	output.Networks = MakeNetworks(config)

	// Ports
	output.Ports = MakePortBinding(config)

	// Privileged
	output.Privileged = MakePrivileged(config)

	// Restart
	output.Restart = MakeRestartOpt(config)

	// TTY
	output.Tty = MakeTTY(config)

	// Volumes
	output.Volumes = MakeVolumes(config)

	return output
}

func MakeBuildOpt(config map[string]interface{}, path string) ServiceBuild {
	output := ServiceBuild{}
	buildOpt := config["build"].(map[string]interface{})
	output.Context = path
	output.Dockerfile = buildOpt["dockerfile"].(string)

	// Only extract build arg if arg exists
	if buildArgs, exist := buildOpt["args"].([]interface{}); exist {
		formattedArg := make(map[string]*string)
		for _, arg := range buildArgs {
			if _, ok := arg.(string); ok {
				splittedString := strings.Split(arg.(string), "=")
				key := splittedString[0]
				value := arg.(string)[len(key+"="):]
				formattedArg[key] = &value
			}
		}
		output.Args = formattedArg
	}

	return output
}

func MakeContainerName(config map[string]interface{}) string {
	return config["container_name"].(string)
}

func MakeAction(config map[string]interface{}) string {
	return config["action"].(string)
}

func MakeCommand(config map[string]interface{}, cmdType string) ShellCommand {
	output := make(ShellCommand, 0)
	if cmdOpt, exist := config[cmdType].([]interface{}); exist {
		for _, args := range cmdOpt {
			output = append(output, args.(string))
		}
	} else {
		return nil
	}
	return output
}

func MakeDependsOn(config map[string]interface{}) []string {
	output := make([]string, 0)
	if dependsOnOpt, exist := config["depends_on"].([]interface{}); exist {
		for _, dependsOn := range dependsOnOpt {
			output = append(output, dependsOn.(string))
		}
	}
	return output
}

func MakeDeployResources(config map[string]interface{}) ServiceResources {
	resources := ServiceResources{}
	if deployOpt, exist := config["deploy"].(map[string]interface{}); exist {
		if resourcesOpt, exist := deployOpt["resources"].(map[string]interface{}); exist {
			limitOpt := resourcesOpt["limits"].(map[string]interface{})
			// CPU usage
			var cpuPeriod float64 = 100000                                   // Default value of 100000
			cpuQuota, _ := strconv.ParseFloat(limitOpt["cpus"].(string), 64) // Combination of period and quota to determine cpu limitation
			resources.CPUQuota = int64(cpuQuota * cpuPeriod)
			resources.CPUPeriod = int64(cpuPeriod)

			// Memory usage
			memoryInString := limitOpt["memory"].(string)
			memory, _ := strconv.ParseInt(memoryInString[:len(memoryInString)-1], 10, 64)
			suffix := string(memoryInString[len(memoryInString)-1])
			switch {
			case suffix == "k" || suffix == "K":
				memory = memory * 1024
			case suffix == "m" || suffix == "M":
				memory = memory * 1048576
			case suffix == "g" || suffix == "G":
				memory = memory * 1073741824
			}
			resources.MemoryLimit = memory
		}
	}
	return resources
}

func MakeEnviroment(config map[string]interface{}) []string {
	env := make([]string, 0)
	// Environment variables
	if envVarsOpt, exist := config["environment"].([]interface{}); exist {
		for _, envVars := range envVarsOpt {
			env = append(env, envVars.(string))
		}
	}
	return env
}

func MakeExtraHosts(config map[string]interface{}, hostname string) []string {
	hosts := make([]string, 0)
	if extraHostsOpt, exist := config["extra_hosts"].([]interface{}); exist {
		for _, host := range extraHostsOpt {
			hosts = append(hosts, host.(string))
		}

	}
	// Default expose host machine
	hosts = append(hosts, fmt.Sprintf("%s:127.0.0.1", hostname))
	return hosts
}

func MakeRestartOpt(config map[string]interface{}) string {
	output := "no"
	if restartOpt, exist := config["restart"].(string); exist {
		output = restartOpt
	}
	return output
}

func MakePortBinding(config map[string]interface{}) []ServicePort {
	ports := make([]ServicePort, 0)
	if portOpt, exist := config["ports"].([]interface{}); exist {
		for _, portData := range portOpt {
			// We need to properly split the string to port and host ip address
			splittedPort := strings.Split(portData.(string), ":")
			port := ServicePort{
				Target:   splittedPort[0],
				Protocol: "tcp",
				HostIp:   "0.0.0.0",
				HostPort: splittedPort[1],
			}
			ports = append(ports, port)
		}
	}
	return ports
}

func MakePrivileged(config map[string]interface{}) bool {
	if privileged, exist := config["privileged"].(bool); exist {
		return privileged
	}
	return false
}

func MakeTTY(config map[string]interface{}) bool {
	if tty, exist := config["tty"].(bool); exist {
		return tty
	}
	return false
}

func MakeNetworks(config map[string]interface{}) []ServiceNetwork {
	network := make([]ServiceNetwork, 0)
	if networkOpts, exist := config["networks"].(map[string]interface{}); exist {
		for name, networkData := range networkOpts {
			network = append(network, ServiceNetwork{
				Name: name,
				IPv4: networkData.(map[string]interface{})["ipv4_address"].(string),
			})
		}
	}
	return network
}

func MakeVolumes(config map[string]interface{}) []ServiceVolume {
	volumes := make([]ServiceVolume, 0)
	if volumeOpt, exist := config["volumes"].([]interface{}); exist {
		fromStringToVolume := func(volStr string) ServiceVolume {
			separateValues := strings.Split(volStr, ":")
			if len(separateValues) >= 2 {
				return ServiceVolume{
					Type:        VolumeTypeBind,
					Source:      separateValues[0],
					Destination: separateValues[1],
				}
			}
			return ServiceVolume{}
		}
		for _, volData := range volumeOpt {
			volumes = append(volumes, fromStringToVolume(volData.(string)))
		}
	}
	return volumes
}

func MakeImage(config map[string]interface{}) string {
	output := ""
	if image, exist := config["image"].(string); exist {
		output = image
	}
	return output
}
