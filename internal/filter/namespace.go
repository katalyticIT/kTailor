package filter

import (
	"strings"

	"ktailor/internal/config"
	"ktailor/internal/logger"
)

var forbiddenNamespaces = []string{
	"kube-system",
	"kube-node-lease",
	"kube-public",
	"cert-manager",
}

// IsNamespaceAllowed checks if mutation is permitted in the given namespace.
func IsNamespaceAllowed(ns string, cfg config.NamespaceConfig, ktailorNamespace string) bool {
	// 1. Check hardcoded forbidden namespaces
	for _, forbidden := range forbiddenNamespaces {
		if ns == forbidden {
			logger.Logf("DEBUG", "Namespace '%s' is in internal forbidden list. Skipped.", ns)
			return false
		}
	}

	// 2. Prevent mutation of kTailor itself
	if ns == ktailorNamespace {
		logger.Logf("DEBUG", "Namespace '%s' is kTailor's own namespace. Skipped.", ns)
		return false
	}

	// 3. Check against configured rules
	matched := false

	// Check Exact
	for _, exact := range cfg.Match.Exact {
		if ns == exact {
			matched = true
			break
		}
	}

	// Check StartsWith
	if !matched {
		for _, prefix := range cfg.Match.StartsWith {
			if strings.HasPrefix(ns, prefix) {
				matched = true
				break
			}
		}
	}

	// Check EndsWith
	if !matched {
		for _, suffix := range cfg.Match.EndsWith {
			if strings.HasSuffix(ns, suffix) {
				matched = true
				break
			}
		}
	}

	// 4. Apply mode logic
	mode := strings.ToLower(cfg.Mode)
	if mode == "allowlist" {
		if !matched {
			logger.Logf("DEBUG", "Namespace '%s' not in allowlist. Skipped.", ns)
		}
		return matched // Only allowed if it matched a rule
	}

	// Default: blocklist
	if matched {
		logger.Logf("DEBUG", "Namespace '%s' is in blocklist. Skipped.", ns)
		return false // Blocked because it matched a rule
	}

	return true // Allowed by default in blocklist mode
}

