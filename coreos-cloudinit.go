package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/initialize"
	"github.com/coreos/coreos-cloudinit/network"
	"github.com/coreos/coreos-cloudinit/system"
)

const version = "0.7.1+git"

func main() {
	var printVersion bool
	flag.BoolVar(&printVersion, "version", false, "Print the version and exit")

	var ignoreFailure bool
	flag.BoolVar(&ignoreFailure, "ignore-failure", false, "Exits with 0 status in the event of malformed input from user-data")

	var file string
	flag.StringVar(&file, "from-file", "", "Read user-data from provided file")

	var configdrive string
	flag.StringVar(&configdrive, "from-configdrive", "", "Read user-data from provided cloud-drive directory")

	var url string
	flag.StringVar(&url, "from-url", "", "Download user-data from provided url")

	var useProcCmdline bool
	flag.BoolVar(&useProcCmdline, "from-proc-cmdline", false, fmt.Sprintf("Parse %s for '%s=<url>', using the cloud-config served by an HTTP GET to <url>", datasource.ProcCmdlineLocation, datasource.ProcCmdlineCloudConfigFlag))

	var convertNetconf string
	flag.StringVar(&convertNetconf, "convert-netconf", "", "Read the network config provided in cloud-drive and translate it from the specified format into networkd unit files (requires the -from-configdrive flag)")

	var workspace string
	flag.StringVar(&workspace, "workspace", "/var/lib/coreos-cloudinit", "Base directory coreos-cloudinit should use to store data")

	var sshKeyName string
	flag.StringVar(&sshKeyName, "ssh-key-name", initialize.DefaultSSHKeyName, "Add SSH keys to the system with the given name")

	flag.Parse()

	if printVersion == true {
		fmt.Printf("coreos-cloudinit version %s\n", version)
		os.Exit(0)
	}

	var ds datasource.Datasource
	if file != "" {
		ds = datasource.NewLocalFile(file)
	} else if url != "" {
		ds = datasource.NewMetadataService(url)
	} else if configdrive != "" {
		ds = datasource.NewConfigDrive(configdrive)
	} else if useProcCmdline {
		ds = datasource.NewProcCmdline()
	} else {
		fmt.Println("Provide one of --from-file, --from-configdrive, --from-url or --from-proc-cmdline")
		os.Exit(1)
	}

	if convertNetconf != "" && configdrive == "" {
		fmt.Println("-convert-netconf flag requires the use of -from-configdrive")
		os.Exit(1)
	}

	switch convertNetconf {
	case "debian":
	default:
		fmt.Printf("Invalid option to -convert-netconf: '%s'. Supported options: 'debian'\n", convertNetconf)
		os.Exit(1)
	}

	fmt.Printf("Fetching user-data from datasource of type %q\n", ds.Type())
	userdataBytes, err := ds.Fetch()
	if err != nil {
		fmt.Printf("Failed fetching user-data from datasource: %v\n", err)
		if ignoreFailure {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	env := initialize.NewEnvironment("/", workspace)
	if len(userdataBytes) > 0 {
		if err := processUserdata(string(userdataBytes), env); err != nil {
			fmt.Printf("Failed resolving user-data: %v\n", err)
			if !ignoreFailure {
				os.Exit(1)
			}
		}
	} else {
		fmt.Println("No user data to handle.")
	}

	if convertNetconf != "" {
		if err := processNetconf(convertNetconf, configdrive); err != nil {
			fmt.Printf("Failed to process network config: %v\n", err)
			if !ignoreFailure {
				os.Exit(1)
			}
		}
	}
}

func processUserdata(userdata string, env *initialize.Environment) error {
	userdata = env.Apply(userdata)

	parsed, err := initialize.ParseUserData(userdata)
	if err != nil {
		fmt.Printf("Failed parsing user-data: %v\n", err)
		return err
	}

	err = initialize.PrepWorkspace(env.Workspace())
	if err != nil {
		fmt.Printf("Failed preparing workspace: %v\n", err)
		return err
	}

	switch t := parsed.(type) {
	case initialize.CloudConfig:
		err = initialize.Apply(t, env)
	case system.Script:
		var path string
		path, err = initialize.PersistScriptInWorkspace(t, env.Workspace())
		if err == nil {
			var name string
			name, err = system.ExecuteScript(path)
			initialize.PersistUnitNameInWorkspace(name, env.Workspace())
		}
	}

	return err
}

func processNetconf(convertNetconf, configdrive string) error {
	openstackRoot := path.Join(configdrive, "openstack")
	metadataBytes, err := ioutil.ReadFile(path.Join(openstackRoot, "latest/meta_data.json"))
	if err != nil {
		return err
	}

	var metadata struct {
		NetworkConfig struct {
			ContentPath string `json:"content_path"`
		} `json:"network_config"`
	}
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return err
	}
	configPath := metadata.NetworkConfig.ContentPath
	if configPath == "" {
		return nil
	}

	netconfBytes, err := ioutil.ReadFile(path.Join(openstackRoot, configPath))
	if err != nil {
		return err
	}

	var interfaces []network.InterfaceGenerator
	switch convertNetconf {
	case "debian":
		interfaces, err = network.ProcessDebianNetconf(string(netconfBytes))
	default:
		return fmt.Errorf("Unsupported network config format '%s'", convertNetconf)
	}

	if err != nil {
		return err
	}

	if err := system.WriteNetworkdConfigs(interfaces); err != nil {
		return err
	}
	return system.RestartNetwork(interfaces)
}
