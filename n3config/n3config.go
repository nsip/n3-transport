// n3config.go

package n3config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/nsip/n3-transport/common"
	"github.com/nsip/n3-transport/n3crypto"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

//
// Configs are required for all tools, node, dispatcher etc.
// they all follow a common base set of requirements
// to provide configurable addresses for the various servers
// and to create a pub/priv keypair identity for the
// given tool.
//
// This module provides routines to ensure a base config
// is created consistently
//

var cfgFileName = "n3config"
var cfgFileType = "toml"

var ConfigNotFoundError = errors.New("no config file found")

//
// reads the config from the expected path into the
// default viper instance for the calling application
//
func ReadConfig() error {

	if configExists() {
		configDir, err := getConfigDir()
		if err != nil {
			return err
		}

		// set up viper
		viper.SetConfigName(cfgFileName)
		viper.SetConfigType(cfgFileType)
		viper.AddConfigPath(configDir)

		// attempt a viper read
		if err := viper.ReadInConfig(); err == nil {
			// config exists
			// log.Println("Using config file:", viper.ConfigFileUsed())
			return nil
		} else {
			return errors.Wrap(err, "cannot read config file:")
		}
	} else {
		return ConfigNotFoundError
	}

	return nil

}

//
// generates a base config with default server ids and a
// keypair identity
//
func CreateBaseConfig() error {

	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// ensure required folder exists
	err = common.EnsureDir(configDir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "cannot create config directory: ")
	}
	// create an empty config file
	configFile := fmt.Sprintf("%s.%s", cfgFileName, cfgFileType)
	fullFileName := fmt.Sprintf("%s/%s", configDir, configFile)
	err = touch(fullFileName)
	if err != nil {
		return errors.Wrap(err, "cannot create config file: ")
	}

	// set up viper
	viper.SetConfigName(cfgFileName)
	viper.SetConfigType(cfgFileType)
	viper.AddConfigPath(configDir)

	// set base config
	// set server params
	viper.Set("nats_addr", "nats://localhost:4222")
	viper.Set("lb_addr", "localhost:9292")
	viper.Set("rpc_port", 5777)
	viper.Set("influx_addr", "http://localhost:8086")

	// generate identity keys
	pubkey, privkey, err := n3crypto.GenerateKeys()
	if err != nil {
		return errors.Wrap(err, "create config failed, error generating keys: ")
	}
	viper.Set("pubkey", pubkey)
	viper.Set("privkey", privkey)

	// write the new config file
	err = viper.WriteConfig()
	if err != nil {
		return errors.Wrap(err, "createBaseConfig failed, cannot save config file: ")
	} else {
		log.Println("new config written.", viper.ConfigFileUsed())
	}

	// re-read to confirm
	if err := viper.ReadInConfig(); err == nil {
		// config exists
		// log.Println("Using config file:", viper.ConfigFileUsed())
		return nil
	} else {
		return errors.Wrap(err, "new base config create failed:")
	}

	return nil
}

//
// writes the current state of the config to file
//
func SaveConfig() error {

	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// set up viper
	viper.SetConfigName(cfgFileName)
	viper.SetConfigType(cfgFileType)
	viper.AddConfigPath(configDir)

	// write the new config file
	err = viper.WriteConfig()
	if err != nil {
		return errors.Wrap(err, "createBaseConfig failed, cannot save config file: ")
	} else {
		log.Println("new config written.", viper.ConfigFileUsed())
	}

	return nil
}

//
// gets the expected config directory
//
func getConfigDir() (string, error) {

	// get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "cannot establish cwd:")
	}
	// log.Println("cwd:", cwd)
	configDir := fmt.Sprintf("%s/config", cwd)

	return configDir, nil
}

//
// checks if a config file can be found in the expected path
//
func configExists() bool {

	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	// log.Println("cwd:", cwd)
	configDir := fmt.Sprintf("%s/config", cwd)
	configFile := fmt.Sprintf("%s.%s", cfgFileName, cfgFileType)
	fullFileName := fmt.Sprintf("%s/%s", configDir, configFile)

	return common.FileExists(fullFileName)

}

//
// creates a blank file at the given location
//
func touch(filename string) error {

	d1 := []byte("\n\n\n")
	err := ioutil.WriteFile(filename, d1, 0644)

	return err
}
