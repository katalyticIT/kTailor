package config

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// --- Main Configuration (ktailor.yaml) ---

type TransportConfig struct {
	Port   int    `json:"port"`
	Secure bool   `json:"secure"`
	Cert   string `json:"cert"`
	Key    string `json:"key"`
}

type LoggingConfig struct {
	Level string `json:"level"`
}

type TemplatesConfig struct {
	AllowCustomTemplates bool `json:"allowCustomTemplates"`
}

type MatchConfig struct {
	Exact      []string `json:"exact"`
	StartsWith []string `json:"startsWith"`
	EndsWith   []string `json:"endsWith"`
}

type NamespaceConfig struct {
	Mode  string      `json:"mode"` // "blocklist" or "allowlist"
	Match MatchConfig `json:"match"`
}

type MainConfig struct {
	Transport  TransportConfig `json:"transport"`
	Logging    LoggingConfig   `json:"logging"`
	Templates  TemplatesConfig `json:"templates"`
	Namespaces NamespaceConfig `json:"namespaces"`
}

// --- Template Configuration (ConfigMaps) ---

type ModifyContainerAction struct {
	Env          []corev1.EnvVar      `json:"env,omitempty"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

// EnvVarAppend ist eine Spezial-Struktur für setOrAppend, da das K8s-Original keinen Separator hat
type EnvVarAppend struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Separator string `json:"separator,omitempty"`
}

type SetOrAppendAction struct {
	Env []EnvVarAppend `json:"env,omitempty"`
}

// RemoveAction erwartet Objekte mit einem Name-Feld, passend zur K8s-Syntax
type RemoveAction struct {
	Env          []corev1.EnvVar      `json:"env,omitempty"`
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

type ModifyContainers struct {
	InsertIfNotExists ModifyContainerAction `json:"insertIfNotExists,omitempty"`
	InsertOrOverwrite ModifyContainerAction `json:"insertOrOverwrite,omitempty"`
	SetOrAppend       SetOrAppendAction     `json:"setOrAppend,omitempty"`
	Remove            RemoveAction          `json:"remove,omitempty"` // <-- nutzt jetzt corev1.EnvVar und corev1.VolumeMount
}

type TemplateConfig struct {
	Kind              string             `json:"kind"`
	ModifyContainers  ModifyContainers   `json:"modifyContainers,omitempty"`
	AddContainers     []corev1.Container `json:"addContainers,omitempty"`
	AddInitContainers []corev1.Container `json:"addInitContainers,omitempty"`
	AddVolumes        AddVolumes         `json:"addVolumes,omitempty"`
}

type AddVolumes struct {
	Volumes []corev1.Volume `json:"volumes,omitempty"`
}

//-- Functions

// LoadMainConfig reads the central configuration file
func LoadMainConfig(path string) (*MainConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg MainConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	// Set default mode if empty
	if cfg.Namespaces.Mode == "" {
		cfg.Namespaces.Mode = "blocklist"
	}

	return &cfg, nil
}
