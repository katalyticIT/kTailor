package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ktailor/internal/config"
	"ktailor/internal/logger"
	"ktailor/internal/modifier"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"sigs.k8s.io/yaml"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

const LabelKey = "ktailor.io/fit"

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	length     int
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

// Serve now accepts the ktailorNamespace
func Serve(cfg *config.MainConfig, lister corev1listers.ConfigMapLister, ktailorNamespace string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		defer func() {
			logger.Logf("DEBUG", "HTTP Request %s %s Status: %d Length: %d", r.Method, r.URL.Path, lw.statusCode, lw.length)
		}()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(lw, "could not read body", http.StatusBadRequest)
			return
		}

		var admissionReviewReq admissionv1.AdmissionReview
		if _, _, err := deserializer.Decode(body, nil, &admissionReviewReq); err != nil {
			http.Error(lw, "could not decode body", http.StatusBadRequest)
			return
		}

		admissionResponse := mutate(admissionReviewReq.Request, cfg, lister, ktailorNamespace)
		admissionReviewResponse := admissionv1.AdmissionReview{
			TypeMeta: metav1.TypeMeta{Kind: "AdmissionReview", APIVersion: "admission.k8s.io/v1"},
			Response: admissionResponse,
		}
		admissionReviewResponse.Response.UID = admissionReviewReq.Request.UID

		respBytes, _ := json.Marshal(admissionReviewResponse)
		lw.Write(respBytes)
	}
}

func mutate(req *admissionv1.AdmissionRequest, cfg *config.MainConfig, lister corev1listers.ConfigMapLister, ktailorNamespace string) *admissionv1.AdmissionResponse {
	if req.Resource.Resource != "deployments" {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	var deployment appsv1.Deployment
	if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
		return &admissionv1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
	}

	labels := deployment.GetLabels()
	labelValue, ok := labels[LabelKey]
	if !ok {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	parts := strings.SplitN(labelValue, ".", 2)
	if len(parts) != 2 {
		logger.Logf("WARN", "Invalid label format (expected prefix.templateName): %s", labelValue)
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	prefix := parts[0]
	templateName := parts[1]

	var template *config.TemplateConfig
	var err error

	// Changed 'ktailor' to 'central'
	switch prefix {
	case "central":
		// Load from ktailor namespace
		template, err = loadTemplateFromCache(lister, ktailorNamespace, templateName)
	case "local":
		// Load from deployment's namespace
		if !cfg.Templates.AllowCustomTemplates {
			logger.Logf("INFO", "Local template %s requested but rejected (AllowCustomTemplates=false)", templateName)
			return &admissionv1.AdmissionResponse{Allowed: true}
		}
		template, err = loadTemplateFromCache(lister, req.Namespace, templateName)
	default:
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	if err != nil {
		logger.Logf("ERROR", "Failed to load template %s: %v", templateName, err)
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	logger.Logf("INFO", "Applying %s template '%s' to deployment '%s'", prefix, templateName, deployment.Name)

	patchBytes, err := modifier.CreatePatch(&deployment, template, req.Object.Raw)
	if err != nil {
		logger.Logf("ERROR", "Failed to create JSON patch for %s: %v", deployment.Name, err)
		return &admissionv1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
	}

	pt := admissionv1.PatchTypeJSONPatch
	return &admissionv1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &pt,
	}
}

// Unified loading function for both central and local templates
func loadTemplateFromCache(lister corev1listers.ConfigMapLister, namespace, cmName string) (*config.TemplateConfig, error) {
	cm, err := lister.ConfigMaps(namespace).Get(cmName)
	if err != nil {
		return nil, fmt.Errorf("ConfigMap not found in cache: %v", err)
	}

	if len(cm.Data) == 0 {
		return nil, fmt.Errorf("ConfigMap %s is empty", cmName)
	}

	// Read the first available key in the ConfigMap data, ignoring its name
	var yamlData string
	for _, v := range cm.Data {
		yamlData = v
		break
	}

	var tmpl config.TemplateConfig
	err = yaml.Unmarshal([]byte(yamlData), &tmpl)
	return &tmpl, err
}
