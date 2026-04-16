package config

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type MainConfig struct {
	Transport struct {
		Port   int    `json:"port"`
		Secure bool   `json:"secure"`
		Cert   string `json:"cert"`
		Key    string `json:"key"`
	} `json:"transport"`
	Logging struct {
		Level string `json:"level"`
	} `json:"logging"`
	Templates struct {
		AllowCustomTemplates bool `json:"allowCustomTemplates"`
	} `json:"templates"`
}

type TemplateConfig struct {
	Meta struct {
		Author  string `json:"autor"`
		Version string `json:"version"`
		Date    string `json:"date"`
	} `json:"meta"`
	ModifyContainers struct {
		InsertIfNotExists ElementBlock `json:"insertIfNotExists"`
		InsertOrOverwrite ElementBlock `json:"insertOrOverwrite"`
		SetOrAppend       ElementBlock `json:"setOrAppend"`
		Remove            ElementBlock `json:"remove"`
	} `json:"modifyContainers"`
	AddContainers     []corev1.Container `json:"addContainers"`
	AddInitContainers []corev1.Container `json:"addInitContainers"`
	AddVolumes        struct {
		Volumes []corev1.Volume `json:"volumes"`
	} `json:"addVolumes"`
}

type ElementBlock struct {
	Env []struct {
		Name      string `json:"name"`
		Value     string `json:"value"`
		Separator string `json:"separator,omitempty"`
	} `json:"env"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts"`
}

func LoadMainConfig(path string) (*MainConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg MainConfig
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}

