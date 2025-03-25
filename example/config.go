package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	HTTP struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"http"`

	Log struct {
		Compress   bool   `yaml:"compress"`
		Debug      string `yaml:"debug"`
		Dir        string `yaml:"dir"`
		MaxAge     int    `yaml:"max_age"`
		MaxBackups int    `yaml:"max_backups"`
		MaxSize    int    `yaml:"max_size"`
	} `yaml:"log"`

	Auth struct {
		AccessTokenTTL  int    `yaml:"access_token_ttl"`
		Key             string `yaml:"key"`
		RefreshTokenTTL int    `yaml:"refresh_token_ttl"`
	} `yaml:"auth"`

	DB struct {
		DSN         string `yaml:"dsn"`
		MaxIdleConn int    `yaml:"max_idle_conn"`
		MaxOpenConn int    `yaml:"max_open_conn"`
	} `yaml:"db"`

	Redis struct {
		Address     []string `yaml:"address"`
		ClusterMode bool     `yaml:"cluster_mode"`
		DB          int      `yaml:"db"`
		MaxRetry    int      `yaml:"max_retry"`
		Password    string   `yaml:"password"`
		Username    string   `yaml:"username"`
	} `yaml:"redis"`

	Salt struct {
		Password string `yaml:"password"`
	} `yaml:"salt"`
}

func parseConfig(path string) (*config, error) {
	var cfg *config

	//Open config file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//Decode to yaml
	d := yaml.NewDecoder(file)
	if err := d.Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
