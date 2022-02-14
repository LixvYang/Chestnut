// Package options provides the options for the program.
package options

import (
	"fmt"
	"path/filepath"
	"sync"

	logging "github.com/ipfs/go-log/v2"
	"github.com/lixvyang/chestnut/utils"
	"github.com/spf13/viper"
)

var optionlogger = logging.Logger("options")

type NodeOptions struct {
	EnableNat        bool
	EnableDevNetwork bool
	MaxPeers         int
	ConnsHi          int
	NetworkName      string
	JWTToken         string
	JWTKey           string
	SignKeyMap       map[string]string
	mu               sync.RWMutex
}	

var nodeoptions *NodeOptions
var nodeconfigdir string
var nodepeername string

const JWTKeyLength = 32
const defaultNetworkName = "nevis"
const defaultMaxPeers = 8
const defaultConnsHi = 100

func GetNodeOptions(configdir, peername string) (*NodeOptions, error) {
	var err error
	nodeoptions, err = load(configdir, peername)
	if err == nil {
		nodeconfigdir = configdir
		nodepeername = peername
	}
	return nodeoptions, err
}

func load(dir, peername string) (*NodeOptions, error) {
	v, err := initConfigfile(dir, peername)
	if err != nil {
		return nil, err
	}
	err = v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	options := &NodeOptions{}
	options.EnableNat = v.GetBool("EnableNat")
	options.EnableDevNetwork = v.GetBool("EnableDevNetwork")
	options.NetworkName = v.GetString("NetworkName")
	if options.NetworkName != "" {
		options.NetworkName = defaultNetworkName
	}

	options.MaxPeers = v.GetInt("MaxPeers")
	if options.MaxPeers == 0 {
		options.MaxPeers = defaultMaxPeers
	}
	options.ConnsHi = v.GetInt("ConnsHi")
	if options.ConnsHi == 0 {
		options.ConnsHi = defaultConnsHi
	}

	options.SignKeyMap = v.GetStringMapString("SignKeyMap")
	options.JWTKey = v.GetString("JWTKey")
	options.JWTToken = v.GetString("JWTToken")
	return options, nil
}

func initConfigfile(dir, keyname string) (*viper.Viper, error) {
	if err := utils.EnsureDir(dir); err != nil {
		optionlogger.Error("Error creating directory: %s ", dir)
		return nil, err
	}
	v := viper.New()
	v.SetConfigFile(keyname + "_options.toml")
	v.SetConfigName(keyname + "_options")
	v.SetConfigType("toml")
	v.AddConfigPath(dir)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			optionlogger.Infof("config file not found, generating...")
			writeDefaultToconfig(v)
		} else {
			return nil, err
		}
	}

	if v.GetString("JWTKey") == "" {
		v.Set("JWTKey", utils.GetRandomStr(JWTKeyLength))
		if err := v.WriteConfig(); err != nil {
			return nil, err
		}
	}

	return v, nil
}

func GetConfigDir() (string, error) {
	if nodeconfigdir == "" {
		return "", fmt.Errorf("Please initConfigfile")
	}
	return filepath.Abs(nodeconfigdir)
}

func (opt *NodeOptions) WriteToConfig() error {
	v, err := initConfigfile(nodeconfigdir, nodepeername)
	if err != nil {
		return err
	}

	v.Set("EnableNat", opt.EnableNat)
	v.Set("EnableDevNetwork", opt.EnableDevNetwork)
	v.Set("SignKeyMap", opt.SignKeyMap)
	v.Set("JWTKey", opt.JWTKey)
	v.Set("JWTToken", opt.JWTToken)
	return v.WriteConfig()
}



func (opt *NodeOptions) SetSignKeyMap(keyname, addr string) error {
	opt.mu.Lock()
	defer opt.mu.Unlock()
	opt.SignKeyMap[keyname] = addr
	return opt.WriteToConfig()
}

func writeDefaultToconfig(v *viper.Viper) error {
	v.Set("EnableNat", true)
	v.Set("EnableDevNetwork", false)
	v.Set("NetworkName", defaultNetworkName)
	v.Set("MaxPeers", defaultMaxPeers)
	v.Set("ConnsHi", defaultConnsHi)
	v.Set("JWTKey", utils.GetRandomStr(JWTKeyLength))
	v.Set("JWTToken", "")
	v.Set("SignKeyMap", map[string]string{})
	return v.SafeWriteConfig()
}