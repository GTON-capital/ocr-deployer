package models

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type NodeInfo struct {
	P2PKey          string `yaml:"p2p-key"`
	P2PPublicKey    string `yaml:"p2p-public-key"`
	OCRKeyID        string `yaml:"ocr-key-id"`
	OCRKey          string `yaml:"ocr-key"`
	SigningAddress  string `yaml:"signing-address"`
	ConfigPublicKey string `yaml:"config-public-key"`

	OracleAddress   string `yaml:"oracle-address"`
	OperatorAddress string `yaml:"operator-address"`
}

type OCRSpec struct {
	Description string     `yaml:"description"`
	LinkToken   string     `yaml:"link_token"`
	Nodes       []NodeInfo `yaml:"nodes"`
}

func LoadSpec(fname string) (*OCRSpec, error) {
	spec := OCRSpec{}
	yamlFile, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &spec)
	if err != nil {
		return nil, err
	}
	return &spec, nil
}
