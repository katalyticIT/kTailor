package modifier

import (
	"encoding/json"

	"ktailor/internal/config"
	"ktailor/internal/logger"

	"github.com/mattbaird/jsonpatch"
	appsv1 "k8s.io/api/apps/v1"
)

// CreatePatch Signatur hat keinen 'debug' bool mehr
func CreatePatch(d *appsv1.Deployment, t *config.TemplateConfig, originalJSON []byte) ([]byte, error) {
	mod := d.DeepCopy()

	logger.Logf("DEBUG", "Phase modifyContainers gestartet")
	ProcessModifyContainers(mod, t)
	logPhaseEnd("modifyContainers", originalJSON, mod)

	logger.Logf("DEBUG", "Phase addContainers gestartet")
	ProcessAddContainers(mod, t)
	logPhaseEnd("addContainers", originalJSON, mod)

	logger.Logf("DEBUG", "Phase addInitContainers gestartet")
	ProcessAddInitContainers(mod, t)
	logPhaseEnd("addInitContainers", originalJSON, mod)

	logger.Logf("DEBUG", "Phase addVolumes gestartet")
	ProcessAddVolumes(mod, t)
	logPhaseEnd("addVolumes", originalJSON, mod)

	modJSON, err := json.Marshal(mod)
	if err != nil {
		return nil, err
	}

	patch, err := jsonpatch.CreatePatch(originalJSON, modJSON)
	if err != nil {
		return nil, err
	}
	
	return json.Marshal(patch)
}

func logPhaseEnd(phase string, orig []byte, mod *appsv1.Deployment) {
	// Teure JSON-Operationen nur ausführen, wenn DEBUG aktiv ist!
	if logger.IsDebugEnabled() {
		modJSON, _ := json.Marshal(mod)
		patch, _ := jsonpatch.CreatePatch(orig, modJSON)
		patchJSON, _ := json.Marshal(patch)
		logger.Logf("DEBUG", "Ergebnis %s: %s", phase, string(patchJSON))
	}
}
