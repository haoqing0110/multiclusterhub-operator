package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	operatorsv1alpha1 "github.com/open-cluster-management/multicloudhub-operator/pkg/apis/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// WebhookServiceName ...
	WebhookServiceName = "multiclusterhub-operator-webhook"
	// APIServerSecretName ...
	APIServerSecretName = "mcm-apiserver-self-signed-secrets" // #nosec G101 (no confidential credentials)
	// KlusterletSecretName ...
	KlusterletSecretName = "mcm-klusterlet-self-signed-secrets" // #nosec G101 (no confidential credentials)

	// CertManagerNamespace ...
	CertManagerNamespace = "cert-manager"

	// MongoEndpoints ...
	MongoEndpoints = "multicluster-mongodb"
	// MongoReplicaSet ...
	MongoReplicaSet = "rs0"
	// MongoTLSSecret ...
	MongoTLSSecret = "multicluster-mongodb-client-cert"
	// MongoCaSecret ...
	MongoCaSecret = "multicloud-ca-cert" // #nosec G101 (no confidential credentials)

	podNamespaceEnvVar = "POD_NAMESPACE"
	apiserviceName     = "mcm-apiserver"
	rsaKeySize         = 2048
	duration365d       = time.Hour * 24 * 365

	// DefaultRepository ...
	DefaultRepository = "quay.io/open-cluster-management"

	// directory housing image manifests (also specified in Dockerfile)
	ImageManifestsDir = "image-manifests"
)

// CacheSpec ...
type CacheSpec struct {
	IngressDomain   string
	ImageShaDigests map[string]string
	ISDVersion      string
}

// CertManagerNS returns the namespace to deploy cert manager objects
func CertManagerNS(m *operatorsv1alpha1.MultiClusterHub) string {
	if m.Spec.CloudPakCompatibility {
		return CertManagerNamespace
	}
	return m.Namespace
}

// ContainsPullSecret returns whether a list of pullSecrets contains a given pull secret
func ContainsPullSecret(pullSecrets []corev1.LocalObjectReference, ps corev1.LocalObjectReference) bool {
	for _, v := range pullSecrets {
		if v == ps {
			return true
		}
	}
	return false
}

// ContainsMap returns whether the expected map entries are included in the map
func ContainsMap(all map[string]string, expected map[string]string) bool {
	for key, exval := range expected {
		allval, ok := all[key]
		if !ok || allval != exval {
			return false
		}

	}
	return true
}

// AddInstallerLabel adds Installer Labels ...
func AddInstallerLabel(u *unstructured.Unstructured, name string, ns string) {
	labels := make(map[string]string)
	for key, value := range u.GetLabels() {
		labels[key] = value
	}
	labels["installer.name"] = name
	labels["installer.namespace"] = ns

	u.SetLabels(labels)
}

// CoreToUnstructured converts a Core Kube resource to unstructured
func CoreToUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	content, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	u := &unstructured.Unstructured{}
	err = u.UnmarshalJSON(content)
	return u, err
}

// MchIsValid Checks if the optional default parameters need to be set
func MchIsValid(m *operatorsv1alpha1.MultiClusterHub) bool {
	invalid := m.Spec.Version == "" ||
		m.Spec.ImageRepository == "" ||
		m.Spec.ImagePullPolicy == "" ||
		m.Spec.Mongo.Storage == "" ||
		m.Spec.Mongo.StorageClass == "" ||
		m.Spec.Etcd.Storage == "" ||
		m.Spec.Etcd.StorageClass == "" ||
		m.Spec.ReplicaCount <= 0 ||
		m.Spec.Mongo.ReplicaCount <= 0

	return !invalid
}

//GetImageReference returns full image reference for images referenced directly within operator
func GetImageReference(mch *operatorsv1alpha1.MultiClusterHub, imageName, imageVersion string, cache CacheSpec) string {

	imageShaDigest := cache.ImageShaDigests[strings.ReplaceAll(imageName, "-", "_")]

	if imageShaDigest != "" {
		return fmt.Sprintf("%s/%s@%s", mch.Spec.ImageRepository, imageName, imageShaDigest)
	} else {
		imageRef := fmt.Sprintf("%s/%s:%s", mch.Spec.ImageRepository, imageName, imageVersion)
		if mch.Spec.ImageTagSuffix == "" {
			return imageRef
		}
		return imageRef + "-" + mch.Spec.ImageTagSuffix
	}
}
