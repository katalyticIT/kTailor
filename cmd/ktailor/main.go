package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"ktailor/internal/config"
	"ktailor/internal/logger"
	"ktailor/internal/webhook"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func main() {
	cfg, err := config.LoadMainConfig("/etc/ktailor/ktailor.yaml")
	if err != nil {
		fmt.Printf("FATAL Failed to load config: %v\n", err)
		return
	}

	logger.Init(cfg.Logging.Level)

	// Fetch the namespace ktailor is running in
	ktailorNamespace := os.Getenv("POD_NAMESPACE")
	if ktailorNamespace == "" {
		ktailorNamespace = "ktailor" // Fallback
		logger.Logf("WARN", "POD_NAMESPACE env var not set, falling back to '%s'", ktailorNamespace)
	}

	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Logf("FATAL", "Failed to load in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		logger.Logf("FATAL", "Failed to create K8s clientset: %v", err)
	}

	tweakListOptions := func(options *metav1.ListOptions) {
		options.LabelSelector = "ktailor.io/template=true"
	}
	factory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		30*time.Second,
		informers.WithTweakListOptions(tweakListOptions),
	)

	cmInformer := factory.Core().V1().ConfigMaps()
	cmLister := cmInformer.Lister()

	cmInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if cm, ok := obj.(*corev1.ConfigMap); ok {
				logger.Logf("INFO", "Template loaded into cache: %s/%s", cm.Namespace, cm.Name)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if newCm, ok := newObj.(*corev1.ConfigMap); ok {
				oldCm, _ := oldObj.(*corev1.ConfigMap)
				if newCm.ResourceVersion != oldCm.ResourceVersion {
					logger.Logf("INFO", "Template updated in cache: %s/%s", newCm.Namespace, newCm.Name)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if cm, ok := obj.(*corev1.ConfigMap); ok {
				logger.Logf("INFO", "Template deleted from cache: %s/%s", cm.Namespace, cm.Name)
			}
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)
	factory.Start(stopCh)

	logger.Logf("INFO", "Waiting for informer cache to sync...")
	if !cache.WaitForCacheSync(stopCh, cmInformer.Informer().HasSynced) {
		logger.Logf("FATAL", "Failed to sync cache")
	}

	cachedTemplates, err := cmLister.List(labels.Everything())
	if err == nil {
		logger.Logf("INFO", "Cache synced successfully. %d templates in RAM.", len(cachedTemplates))
	}

	logger.Logf("INFO", "kTailor started on port %d, Secure: %t", cfg.Transport.Port, cfg.Transport.Secure)

	mux := http.NewServeMux()
	// Pass the ktailorNamespace to the webhook handler
	mux.HandleFunc("/mutate", webhook.Serve(cfg, cmLister, ktailorNamespace))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Transport.Port),
		Handler: mux,
	}

	if cfg.Transport.Secure {
		err = server.ListenAndServeTLS(cfg.Transport.Cert, cfg.Transport.Key)
	} else {
		err = server.ListenAndServe()
	}

	if err != nil {
		logger.Logf("FATAL", "Server stopped: %v", err)
	}
}
