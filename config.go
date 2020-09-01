package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	yaml "gopkg.in/yaml.v2"
)

type Configuration struct {
	Dryrun             bool
	Verbose            bool
	Organization       string
	HardenizeRoot      string
	HardenizeUser      string
	HardenizePasswd    string
	HardenizeWebUser   string
	HardenizeWebPasswd string
	HardenizeWebRoot   string
}

func getConfig() *Configuration {
	var config Configuration
	var conffilename string

	// define and parse command line arguments
	flag.StringVar(&conffilename, "conf", "", "Filename to read configuration from")
	flag.BoolVar(&config.Dryrun, "dryrun", false, "Print results instead of writing to InfluxDB")
	flag.BoolVar(&config.Verbose, "verbose", false, "print more information while running")
	flag.StringVar(&config.Organization, "org", "", "Organisation ID")
	flag.StringVar(&config.HardenizeRoot, "hardenizeRoot", "", "Hardenize API root url")
	flag.StringVar(&config.HardenizeUser, "hardenizeUser", "", "Hardenize API user")
	flag.StringVar(&config.HardenizePasswd, "hardenizePasswd", "", "Hardenize API password")
	flag.StringVar(&config.HardenizeWebUser, "hardenizeWebUser", "", "Hardenize web user")
	flag.StringVar(&config.HardenizeWebPasswd, "hardenizeWebPasswd", "", "Hardenize web password")
	flag.StringVar(&config.HardenizeWebRoot, "hardenizeWebRoot", "", "Hardenize web root url")
	flag.Parse()

	var confFromFile *Configuration
	if conffilename != "" {
		var err error
		confFromFile, err = readConfigFile(conffilename)
		if err != nil {
			panic(err)
		}
	}

	defaultConfig := readDefaultConfigFiles()
	return checkConfiguration(joinConfig(defaultConfig, joinConfig(confFromFile, &config)))
}

func readConfigFile(filename string) (*Configuration, error) {
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	config := &Configuration{}
	err = yaml.Unmarshal(source, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func readDefaultConfigFiles() (config *Configuration) {

	// config in user home directory
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	fileconfig, err := readConfigFile(path.Join(usr.HomeDir, ".hardblame"))
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	config = joinConfig(config, fileconfig)

	// done
	return
}

func joinConfig(oldConf *Configuration, newConf *Configuration) (config *Configuration) {
	if oldConf == nil && newConf == nil {
		return nil
	}
	if oldConf != nil && newConf == nil {
		return oldConf
	}
	if oldConf == nil && newConf != nil {
		return newConf
	}

	// we have two configs, join them
	config = &Configuration{}
	config.Dryrun = newConf.Dryrun
	config.Verbose = newConf.Verbose
	if newConf.Organization != "" {
		config.Organization = newConf.Organization
	} else {
		config.Organization = oldConf.Organization
	}
	if newConf.HardenizeRoot != "" {
		config.HardenizeRoot = newConf.HardenizeRoot
	} else {
		config.HardenizeRoot = oldConf.HardenizeRoot
	}
	if newConf.HardenizeUser != "" {
		config.HardenizeUser = newConf.HardenizeUser
	} else {
		config.HardenizeUser = oldConf.HardenizeUser
	}
	if newConf.HardenizePasswd != "" {
		config.HardenizePasswd = newConf.HardenizePasswd
	} else {
		config.HardenizePasswd = oldConf.HardenizePasswd
	}
	if newConf.HardenizeWebUser != "" {
		config.HardenizeWebUser = newConf.HardenizeWebUser
	} else {
		config.HardenizeWebUser = oldConf.HardenizeWebUser
	}
	if newConf.HardenizeWebPasswd != "" {
		config.HardenizeWebPasswd = newConf.HardenizeWebPasswd
	} else {
		config.HardenizeWebPasswd = oldConf.HardenizeWebPasswd
	}
	if newConf.HardenizeWebRoot != "" {
		config.HardenizeWebRoot = newConf.HardenizeWebRoot
	} else {
		config.HardenizeWebRoot = oldConf.HardenizeWebRoot
	}

	// Done
	return config
}

func checkConfiguration(config *Configuration) *Configuration {
	if len(config.Organization) == 0 {
		panic(errors.New("Organization must be given."))
	}

	// Hardenize Config
	if len(config.HardenizeRoot) == 0 {
		panic(errors.New("Hardenize root url must be given."))
	}
	if len(config.HardenizeUser) == 0 {
		panic(errors.New("Hardenize API user must be given."))
	}
	if len(config.HardenizePasswd) == 0 {
		panic(errors.New("Hardenize API password must be given."))
	}
	if len(config.HardenizeWebUser) == 0 {
		panic(errors.New("Hardenize web user must be given."))
	}
	if len(config.HardenizeWebPasswd) == 0 {
		panic(errors.New("Hardenize web password must be given."))
	}
	if len(config.HardenizeWebRoot) == 0 {
		panic(errors.New("Hardenize web root url must be given."))
	}

	// Done
	return config
}
