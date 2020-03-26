package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"gitlab.com/vocdoni/go-dvote/config"
	"gitlab.com/vocdoni/go-dvote/log"
	"gitlab.com/vocdoni/go-dvote/swarm"
)

func newConfig() (config.PssMetaCfg, error) {
	var globalCfg config.PssMetaCfg
	// setup flags
	home, err := os.UserHomeDir()
	if err != nil {
		return globalCfg, err
	}
	defaultDirPath := home + "/.dvote/pss"
	path := flag.String("cfgpath", defaultDirPath+"/config.yaml", "cfgpath. Specify filepath for pss config")

	flag.String("listenHost", "0.0.0.0", "address host to listen")
	flag.Int16("listenPort", 4171, "address port to listen")
	flag.String("datadir", "", "datadir directory for swarm/pss files")
	flag.String("logLevel", "info", "Log level. Valid values are: debug, info, warn, error, fatal.")

	flag.Parse()

	viper := viper.New()
	viper.SetDefault("listenHost", "0.0.0.0")
	viper.SetDefault("listenPort", 4171)
	viper.SetDefault("datadir", "")
	viper.SetDefault("logLevel", "info")

	viper.SetConfigType("yaml")
	if *path == defaultDirPath+"/config.yaml" { // if path left default, write new cfg file if empty or if file doesn't exist.
		if err := viper.SafeWriteConfigAs(*path); err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(defaultDirPath, os.ModePerm)
				if err != nil {
					return globalCfg, err
				}
				err = viper.WriteConfigAs(*path)
				if err != nil {
					return globalCfg, err
				}
			}
		}
	}

	viper.BindPFlag("listenHost", flag.Lookup("listenHost"))
	viper.BindPFlag("listenPort", flag.Lookup("listenPort"))
	viper.BindPFlag("datadir", flag.Lookup("datadir"))
	viper.BindPFlag("logLevel", flag.Lookup("logLevel"))

	viper.SetConfigFile(*path)
	err = viper.ReadInConfig()
	if err != nil {
		return globalCfg, err
	}

	err = viper.Unmarshal(&globalCfg)
	return globalCfg, err
}

func main() {
	// setup config
	globalCfg, err := newConfig()
	// setup logger
	log.Init(globalCfg.LogLevel, "stdout")
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	sn := new(swarm.SimpleSwarm)
	sn.SetDatadir(globalCfg.Datadir)
	sn.ListenAddr = fmt.Sprintf("%s:%d", globalCfg.ListenHost, globalCfg.ListenPort)
	sp := swarm.NewSwarmPorts()
	sp.P2P = int(globalCfg.ListenPort)
	sp.Bzz = 0
	sp.HTTPRPC = 0
	sp.WebSockets = 0
	sn.Ports = sp

	err = sn.InitPSS([]string{})
	if err != nil {
		log.Errorf("%v\n", err)
		return
	}

	log.Infof("listening on %s:%d", globalCfg.ListenHost, globalCfg.ListenPort)

	// Block forever.
	select {}
}
