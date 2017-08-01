package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"

	"gopkg.in/yaml.v2"

	"github.com/AlecAivazis/survey"
	"github.com/olekukonko/tablewriter"

	"github.com/elbow-jason/dgraph_helper/prompt"
)

const systemDpath = "/etc/systemd/system/"

func main() {
	Install()
}

func ensureLinux() error {
	if runtime.GOOS != "linux" {
		return errors.New("Currently dgraph_helper can only be used on Linux systems")
	}
	return nil
}

func ensurePermissions() error {
	err := unix.Access(systemDpath, unix.W_OK)
	if err != nil {
		return errors.New("Invalid Permissions (try running as root or use sudo)")
	}
	return nil
}

type allConfig struct {
	// helper fields
	installDir     string
	yamlFilename   string
	PeerIP         string
	PeerPort       int
	MyIP           string
	MyPort         int
	TotalGroups    int
	SelectedGroups []int
	// yaml.config fields
	P            string  `yaml:"p"`            // (default "p") Directory to store posting lists.
	W            string  `yaml:"w"`            // (default "w") Directory to store raft write-ahead logs.
	Export       string  `yaml:"export"`       // (default "exports") Directory to store exports.
	Port         int     `yaml:"port"`         // (default 8080) Port to run HTTP service on.
	GrpcPort     int     `yaml:"grpc_port"`    // (default 9080) Port to run gRPC service on.
	Workerport   int     `yaml:"workerport"`   // (default 12345) Port used by worker for internal communication.
	Idx          int     `yaml:"idx"`          // (default 1) RAFT ID that this server will use to join RAFT groups.
	Groups       string  `yaml:"groups"`       // (default "0,1") RAFT groups handled by this server.
	Gentlecommit float64 `yaml:"gentlecommit"` // (default 0.1)
	Trace        float64 `yaml:"trace"`
	Debugmode    bool    `yaml:"debugmode"`
	MemoryMb     float64 `yaml:"memory_mb"`
	// command-line fields
	Bindall bool
	Peer    string // IP_ADDRESS:PORT of any healthy peer.
	My      string // addr:port of this server, so other Dgraph servers can talk to this

	// // Engine Tuning Fields
	// Pending          int    // (default 1000)  Number of pending queries. Useful for rate limiting.
	// PendingProposals int    // (default 2000) Number of pending mutation proposals. Useful for rate limiting.
	// PortOffset       int    //(default 0) Value added to all listening port numbers.
	// PostingTables    string // (default "loadtoram")(oneof ["loadtoram", "memorymap", "nothing"]) Specifies how Badger LSM tree is stored. Options are loadtoram, memorymap and nothing; which consume most to least RAM while providing best to worst performance respectively.
	// Sc               uint   // Max number of pending entries in wal after which snapshot is taken (default 1000)
	// UI               string // (default "/usr/local/share/dgraph/assets") Directory which contains assets for the user interface
	// Cpu          bool
	// Block        int
	// Dumpsg       string
	// ExpandEdge   bool // (default true)
	// ExposeTrace  bool // (default false)
	// GroupConf    string //
	// Mem          string //
	// Nomutations     bool //Don't allow mutations on this server

	// // TLS Fields
	// TlsCaCerts           string // CA Certs file path.
	// TlsCert              string //  Certificate file path.
	// TlsCertKey           string // Certificate key file path.
	// TlsCertKeyPassphrase string // Certificate key passphrase.
	// TlsClientAuth        string // Enable TLS client authentication
	// TlsMaxVersion        string // (default "TLS12") TLS max version.
	// TlsMinVersion        string // (default "TLS11") TLS min version.
	// TlsOn                string //  Use TLS connections with clients.
	// TlsUseSystemCa       bool   //  Include System CA into CA Certs.

}

func defaultConfig() allConfig {
	return allConfig{
		installDir:     "/var/lib/dgraph",
		yamlFilename:   "config.yaml",
		Groups:         "0,1",
		PeerIP:         "",
		PeerPort:       12345,
		MyIP:           "",
		Workerport:     12345,
		Port:           8080,
		GrpcPort:       9080,
		Trace:          0.33,
		Gentlecommit:   0.1,
		MemoryMb:       1024.00,
		Debugmode:      false,
		SelectedGroups: []int{},
		TotalGroups:    2,
		Idx:            1,
	}
}

func (cfg allConfig) toYAML() ([]byte, error) {
	return yaml.Marshal(map[string]interface{}{
		"p":            cfg.P,
		"w":            cfg.W,
		"export":       cfg.Export,
		"port":         cfg.Port,
		"grpc_port":    cfg.GrpcPort,
		"workerport":   cfg.Workerport,
		"idx":          cfg.Idx,
		"groups":       cfg.Groups,
		"gentlecommit": cfg.Gentlecommit,
		"trace":        cfg.Trace,
		"debugmode":    cfg.Debugmode,
		"memory_mb":    cfg.MemoryMb,
	})
}

func (cfg *allConfig) wantsToChangeSubdirectories() bool {
	return prompt.InputYesOrNo("Change dgraph's subdirectories?", false)
}

func (cfg *allConfig) wantsToChangeInstallDir() bool {
	message := fmt.Sprintf("Change dgraph's base directory? [%s]", cfg.installDir)
	return prompt.InputYesOrNo(message, false)
}

func (cfg *allConfig) wantsToChangePorts() bool {
	return prompt.InputYesOrNo("Change dgraph's ports config?", false)
}

func (cfg *allConfig) wantsToChangeCluster() bool {
	return prompt.InputYesOrNo("Change dgraph's cluster config?", false)
}

func (cfg *allConfig) wantsToChangeEngine() bool {
	return prompt.InputYesOrNo("Change dgraph's engine config?", false)
}

func (cfg *allConfig) wantsToCommitConfig() bool {
	return prompt.InputYesOrNo("Proceed with install?", true)
}

func (cfg *allConfig) writeSystemDUnit() {
	filename := path.Join(systemDpath, "dgraph.service")
	err := ioutil.WriteFile(filename, []byte(cfg.systemDUnit()), os.ModePerm)
	if err != nil {
		panic(err)
	}
	reloadDaemons()
}

func (cfg *allConfig) writeConfigDotYaml() {
	yamlBytes, err := cfg.toYAML()
	if err != nil {
		panic(err)
	}
	// install config.yaml
	err = ioutil.WriteFile(cfg.configDotYamlFilepath(), yamlBytes, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func (cfg *allConfig) configDotYamlFilepath() string {
	return path.Join(cfg.installDir, cfg.yamlFilename)
}

func (cfg *allConfig) isFirstServer() bool {
	return prompt.InputYesOrNo("Is this the first server in the cluster?", false)
}

func (cfg *allConfig) changeInstallDir() {
	cfg.installDir = prompt.InputString("The directory to store data folders and config files", cfg.installDir, prompt.AlwaysValid)
}

func (cfg *allConfig) changeP() {
	cfg.P = prompt.InputString("The directory to store posting lists?", cfg.P, prompt.AlwaysValid)
}

func (cfg *allConfig) changeW() {
	cfg.W = prompt.InputString("The directory to store write-ahead logs?", cfg.W, prompt.AlwaysValid)
}

func (cfg *allConfig) changeExport() {
	cfg.Export = prompt.InputString("The directory to store exports?", cfg.Export, prompt.AlwaysValid)
}

func (cfg *allConfig) changePort() {
	cfg.Port = prompt.InputInteger("The port to serve http?", cfg.Port, true, prompt.PortValidator)
}

func (cfg *allConfig) changeGrpcPort() {
	cfg.GrpcPort = prompt.InputInteger("The port to serve grpc?", cfg.GrpcPort, true, prompt.PortValidator)
}

func (cfg *allConfig) changeWorkerport() {
	cfg.Workerport = prompt.InputInteger("The port for worker communication?", cfg.Workerport, true, prompt.PortValidator)
}

func (cfg *allConfig) changePeer() {
	cfg.changePeerIP()
	cfg.changePeerPort()
}
func (cfg *allConfig) changePeerIP() {
	cfg.PeerIP = prompt.InputString("The IP of a healty peer in the cluster?", cfg.PeerIP, prompt.IPv4Validator)
	cfg.updatePeer()
}

func (cfg *allConfig) changePeerPort() {
	cfg.PeerPort = prompt.InputInteger("The workerport of the same peer", cfg.PeerPort, true, prompt.PortValidator)
	cfg.updatePeer()
}

func (cfg *allConfig) updatePeer() {
	if cfg.PeerIP == "" {
		cfg.Peer = ""
	} else {
		cfg.Peer = fmt.Sprintf("%s:%d", cfg.PeerIP, cfg.PeerPort)
	}
}

func (cfg *allConfig) changeTotalGroups() {
	cfg.TotalGroups = prompt.InputInteger("The total number of groups?", cfg.TotalGroups, true, prompt.AtLeast2)
	cfg.changeSelectedGroups()
}

func (cfg *allConfig) changeSelectedGroups() {
	if cfg.TotalGroups > 10 {
		cfg.changeGroupsText()
	} else {
		cfg.changeGroupsMenu()
	}
}

func (cfg *allConfig) changeGroupsText() {
	validators := survey.ComposeValidators(prompt.GroupsRegexValidator, cfg.ensureGroupsRangeValidator())
	cfg.Groups = prompt.InputString("Enter the groups for this server (comma separated ints and int ranges accepted)", cfg.Groups, validators)
}

func (cfg *allConfig) changeGroupsMenu() {
	// exclusive range where start is 0 and count is 2
	selected := prompt.MultiSelectInts("Select the groups (must choose at least one option)\n<<space to select/deselect, arrows to move, enter when done>>", 0, cfg.TotalGroups)
	cfg.Groups = strings.Join(selected, ",")
}

func (cfg *allConfig) changeMyIP() {
	cfg.MyIP = prompt.InputString("The IP of this server?", cfg.MyIP, prompt.IPv4Validator)
}

func (cfg *allConfig) bindallFlag() string {
	return fmt.Sprintf("--bindall=%s", strconv.FormatBool(cfg.Bindall))
}

func (cfg *allConfig) configFlag() string {
	filepath := path.Join(cfg.installDir, cfg.yamlFilename)
	return fmt.Sprintf("--config=%s", filepath)
}

func (cfg *allConfig) downloadAndInstallBinary() {
	filename := "install_dgraph.sh"
	if err := runCommand("curl", "https://nightly.dgraph.io", "-o", filename); err != nil {
		panic(err)
	}
	err := os.Chmod(filename, 0777)
	if err != nil {
		panic(err)
	}
	defer os.Remove(filename)
	if err := runCommand("./install_dgraph.sh"); err != nil {
		panic(err)
	}
}

func (cfg *allConfig) startDgraphCommand() string {
	return fmt.Sprintf("/usr/local/bin/dgraph %s %s", cfg.bindallFlag(), cfg.configFlag())
}

func (cfg *allConfig) systemDUnit() string {
	return fmt.Sprintf(`
[Unit]
Description = Dgraph graph database
Wants=network-online.target
After=network.target network-online.target

[Service]	
ExecStart = %s
`, cfg.startDgraphCommand())
}

func (cfg *allConfig) updateMy() {
	if cfg.MyIP == "" {
		cfg.My = ""
	} else {
		cfg.My = fmt.Sprintf("%s:%d", cfg.MyIP, cfg.Workerport)
	}
}

func (cfg *allConfig) ensureGroupsRangeValidator() survey.Validator {
	return func(answer interface{}) error {
		answerStr := answer.(string)
		nums := splitGroups(answerStr)
		maxGroup, err := maxIntOfSlice(nums)
		if err != nil {
			return fmt.Errorf("At least one group is required.")
		}
		if maxGroup > cfg.TotalGroups-1 {
			return fmt.Errorf("The max group (%d) exceed the highest allowed group (%d) according configured total groups (total-1).", maxGroup, cfg.TotalGroups-1)
		}
		return nil
	}
}

func (cfg *allConfig) changeTrace() {
	cfg.Trace = prompt.InputFloat64("The ratio of queries to trace", cfg.Trace, prompt.ZeroToOneOnly)
}

func (cfg *allConfig) changeGentlecommit() {
	cfg.Gentlecommit = prompt.InputFloat64("Fraction of dirty posting lists to commit every few seconds", cfg.Gentlecommit, prompt.ZeroToOneOnly)
}

func (cfg *allConfig) changeMemoryMb() {
	cfg.MemoryMb = prompt.InputFloat64("Estimated memory the process can take", cfg.MemoryMb, prompt.AtLeast1024)
}

func (cfg *allConfig) changeDebugMode() {
	cfg.Debugmode = prompt.InputYesOrNo("Debug Mode?", cfg.Debugmode)
}

func (cfg *allConfig) changeIdx() {
	cfg.Idx = prompt.InputInteger("RAFT ID that this server will use to join RAFT groups?", cfg.Idx, true, prompt.PositiveIntValidator)
}

func (cfg *allConfig) printConfigTable() {
	yamlFilepath := path.Join(cfg.installDir, cfg.yamlFilename)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Value", "Description", "Destination"})
	data := [][]string{
		[]string{"p", cfg.P, "Postings Files Directory", cfg.P},
		[]string{"w", cfg.W, "Write-Ahead Logs Directory", cfg.W},
		[]string{"export", cfg.Export, "Exports Directory", cfg.Export},
		[]string{"port", int2string(cfg.Port), "HTTP port", yamlFilepath},
		[]string{"grpc_port", int2string(cfg.GrpcPort), "gRPC port", yamlFilepath},
		[]string{"workerport", int2string(cfg.Workerport), "Internal worker port", yamlFilepath},
		[]string{"idx", int2string(cfg.Idx), "Raft ID for joining groups", yamlFilepath},
		[]string{"total groups", int2string(cfg.TotalGroups), "the total number of groups", "nil"},
		[]string{"groups", cfg.Groups, "Groups for this server", yamlFilepath},
		[]string{"memory_mb", float2string(cfg.MemoryMb), "Estimated Memory in MB", yamlFilepath},
		[]string{"gentlecommit", float2string(cfg.Gentlecommit), "Dirty posting commit freq", yamlFilepath},
		[]string{"trace", float2string(cfg.Trace), "Ratio of queries to trace", yamlFilepath},
		[]string{"debugmode", bool2string(cfg.Debugmode), "Debug mode", yamlFilepath},
		[]string{"bindall", bool2string(cfg.Bindall), cfg.serverStartsOn(), cfg.bindallFlag()},
	}
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()
}

func (cfg *allConfig) createInstallDir() {
	os.MkdirAll(cfg.installDir, os.ModePerm)
}

func (cfg *allConfig) createSubirs() {
	os.MkdirAll(cfg.P, os.ModePerm)
	os.MkdirAll(cfg.W, os.ModePerm)
	os.MkdirAll(cfg.Export, os.ModePerm)
}

func (cfg *allConfig) serverStartsOn() string {
	if cfg.Bindall {
		return "Server host is 0.0.0.0"
	}
	return "Server host is 127.0.0.1"
}

// Install runs prompts for configuration files info and command-line flags,
// writes files to selected directories, installs a systemd unit dgraph,
// and starts dgraph as a service
func Install() {

	if err := ensureLinux(); err != nil {
		log.Fatal(err)
	}
	if err := ensurePermissions(); err != nil {
		log.Fatal(err)
	}

	cfg := defaultConfig()
	if cfg.wantsToChangeInstallDir() {
		cfg.changeInstallDir()
	}
	cfg.P = path.Join(cfg.installDir, "p")
	cfg.W = path.Join(cfg.installDir, "w")
	cfg.Export = path.Join(cfg.installDir, "exports")
	if cfg.wantsToChangeSubdirectories() {
		cfg.changeP()
		cfg.changeW()
		cfg.changeExport()
	}
	if cfg.wantsToChangePorts() {
		cfg.changePort()
		cfg.changeGrpcPort()
		cfg.changeWorkerport()
	}
	if cfg.wantsToChangeEngine() {
		cfg.changeMemoryMb()
		cfg.changeDebugMode()
		cfg.changeGentlecommit()
		cfg.changeTrace()
	}
	if cfg.wantsToChangeCluster() {
		cfg.Bindall = true
		cfg.changeIdx()
		if !cfg.isFirstServer() {
			cfg.changePeer()
		}
		cfg.changeTotalGroups()
		cfg.changeMyIP()
	}
	cfg.printConfigTable()

	if cfg.wantsToCommitConfig() {
		fmt.Println("Installing...")
		cfg.downloadAndInstallBinary()
		cfg.writeConfigDotYaml()
		cfg.writeSystemDUnit()
		startDgraphService()
		statusDgraphService()
	}
}

func reloadDaemons() {
	err := runCommand("systemctl", "daemon-reload")
	if err != nil {
		log.Fatal(err)
	}
}

func startDgraphService() {
	err := runCommand("systemctl", "start", "dgraph")
	if err != nil {
		log.Fatal(err)
	}
}

func statusDgraphService() {
	err := runCommand("systemctl", "status", "dgraph")
	if err != nil {
		log.Fatal(err)
	}
}
