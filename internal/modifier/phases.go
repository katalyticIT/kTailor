package modifier

import (
	"ktailor/internal/config"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func ProcessModifyContainers(d *appsv1.Deployment, t *config.TemplateConfig) {
	for i := range d.Spec.Template.Spec.Containers {
		c := &d.Spec.Template.Spec.Containers[i]

		// Env Logic
		processEnv(c, t)
		
		// VolumeMounts Logic
		processVolumeMounts(c, t)
	}
}

func processEnv(c *corev1.Container, t *config.TemplateConfig) {
	// insertIfNotExists
	for _, envConf := range t.ModifyContainers.InsertIfNotExists.Env {
		exists := false
		for _, e := range c.Env {
			if e.Name == envConf.Name {
				exists = true
				break
			}
		}
		if !exists {
			c.Env = append(c.Env, corev1.EnvVar{Name: envConf.Name, Value: envConf.Value})
		}
	}
	// insertOrOverwrite
	for _, envConf := range t.ModifyContainers.InsertOrOverwrite.Env {
		exists := false
		for j := range c.Env {
			if c.Env[j].Name == envConf.Name {
				c.Env[j].Value = envConf.Value
				exists = true
				break
			}
		}
		if !exists {
			c.Env = append(c.Env, corev1.EnvVar{Name: envConf.Name, Value: envConf.Value})
		}
	}
	// setOrAppend
	for _, envConf := range t.ModifyContainers.SetOrAppend.Env {
		exists := false
		for j := range c.Env {
			if c.Env[j].Name == envConf.Name {
				if c.Env[j].Value == "" {
					c.Env[j].Value = envConf.Value
				} else {
					c.Env[j].Value = c.Env[j].Value + envConf.Separator + envConf.Value
				}
				exists = true
				break
			}
		}
		if !exists {
			c.Env = append(c.Env, corev1.EnvVar{Name: envConf.Name, Value: envConf.Value})
		}
	}
	// remove
	for _, envConf := range t.ModifyContainers.Remove.Env {
		for j := 0; j < len(c.Env); j++ {
			if c.Env[j].Name == envConf.Name {
				c.Env = append(c.Env[:j], c.Env[j+1:]...)
				j--
			}
		}
	}
}

func processVolumeMounts(c *corev1.Container, t *config.TemplateConfig) {
	for _, vmConf := range t.ModifyContainers.InsertIfNotExists.VolumeMounts {
		exists := false
		for _, vm := range c.VolumeMounts {
			if vm.MountPath == vmConf.MountPath || vm.Name == vmConf.Name {
				exists = true
				break
			}
		}
		if !exists {
			c.VolumeMounts = append(c.VolumeMounts, vmConf)
		}
	}
	for _, vmConf := range t.ModifyContainers.InsertOrOverwrite.VolumeMounts {
		exists := false
		for j := range c.VolumeMounts {
			if c.VolumeMounts[j].Name == vmConf.Name || c.VolumeMounts[j].MountPath == vmConf.MountPath {
				c.VolumeMounts[j] = vmConf
				exists = true
				break
			}
		}
		if !exists {
			c.VolumeMounts = append(c.VolumeMounts, vmConf)
		}
	}
	for _, vmConf := range t.ModifyContainers.Remove.VolumeMounts {
		for j := 0; j < len(c.VolumeMounts); j++ {
			if c.VolumeMounts[j].Name == vmConf.Name {
				c.VolumeMounts = append(c.VolumeMounts[:j], c.VolumeMounts[j+1:]...)
				j--
			}
		}
	}
}

func ProcessAddContainers(d *appsv1.Deployment, t *config.TemplateConfig) {
	for _, newC := range t.AddContainers {
		exists := false
		for _, existingC := range d.Spec.Template.Spec.Containers {
			if existingC.Name == newC.Name {
				exists = true
				break
			}
		}
		if !exists {
			d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, newC)
		}
	}
}

func ProcessAddInitContainers(d *appsv1.Deployment, t *config.TemplateConfig) {
	for _, newC := range t.AddInitContainers {
		exists := false
		for _, existingC := range d.Spec.Template.Spec.InitContainers {
			if existingC.Name == newC.Name {
				exists = true
				break
			}
		}
		if !exists {
			d.Spec.Template.Spec.InitContainers = append(d.Spec.Template.Spec.InitContainers, newC)
		}
	}
}

func ProcessAddVolumes(d *appsv1.Deployment, t *config.TemplateConfig) {
	for _, newVol := range t.AddVolumes.Volumes {
		exists := false
		for _, existingVol := range d.Spec.Template.Spec.Volumes {
			if existingVol.Name == newVol.Name {
				exists = true
				break
			}
		}
		if !exists {
			d.Spec.Template.Spec.Volumes = append(d.Spec.Template.Spec.Volumes, newVol)
		}
	}
}
