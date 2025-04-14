package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"prx/internal/models"
	"strings"

	"encoding/base64"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Kube struct {
	client *kubernetes.Clientset
}

type ProxyMapping struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

func NewKubeClient() (Kube, error) {
	kubeconfigDataEnc := os.Getenv("PRX_KUBE_CONFIG")

	var config *rest.Config
	var err error

	if kubeconfigDataEnc != "" {
		decodedKubeconfig, err := base64.StdEncoding.DecodeString(kubeconfigDataEnc)
		if err != nil {
			return Kube{}, fmt.Errorf("failed to decode base64 PRX_KUBE_CONFIG: %v", err)
		}

		// Trim leading/trailing quotes if present.
		kubeconfigStr := strings.Trim(string(decodedKubeconfig), "\"")
		config, err = clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigStr))
		if err != nil {
			return Kube{}, fmt.Errorf("failed to parse kubeconfig: %v", err)
		}
	} else {
		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return Kube{}, fmt.Errorf("failed to build config from file %s: %v", kubeconfigPath, err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return Kube{}, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	return Kube{client: clientset}, nil
}

// AddNewProxy
// create a new ingress and TLS certificate for the ingress controller create
// to point to the service which points to the deployment which deploys the
// pods with the containers so the application can receive the request from
// the defined FQDN from the original users request. (Note) the parameter
// anyBody is of type any so the method AddNewProxy can be used in both
// the HandleNewProxy endpoint and the HandlePatchProxy as HandlePatchProxy
// pathes an existing record and updates the values in both the cluster
// config and the in memory records.
func (k Kube) AddNewProxy(anyBody any, namespace, name string) error {

	body := anyBody.(models.AddNewProxy)

	// Create TLS secret with the provided certificate and key.
	secretName := body.From + "-tls"
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Labels: map[string]string{
				"managed-by": "prx",
			},
			Namespace: namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": []byte(body.Cert),
			"tls.key": []byte(body.Key),
		},
	}

	_, err := k.client.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	ingressName := body.From + "-ingress"
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: ingressName,
			Labels: map[string]string{
				"managed-by": "prx",
			},
			Namespace: namespace,
		},
		Spec: networkingv1.IngressSpec{
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{body.From},
					SecretName: secretName,
				},
			},
			Rules: []networkingv1.IngressRule{
				{
					Host: body.From,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path: "/",
									PathType: func() *networkingv1.PathType {
										pt := networkingv1.PathTypePrefix
										return &pt
									}(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: name,
											Port: networkingv1.ServiceBackendPort{
												Number: 443,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = k.client.NetworkingV1().Ingresses(namespace).Create(context.Background(), ingress, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (k Kube) DeleteProxy(namespace, name string) error {
	ingressName := name + "-ingress"
	secret := name + "-tls"
	ingress, err := k.client.NetworkingV1().Ingresses(namespace).Get(context.Background(), ingressName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ingress: %v", err)
	}

	// Check if the "managed-by" label is present and equals "prx"
	if ingress.Labels == nil || ingress.Labels["managed-by"] != "prx" {
		return fmt.Errorf("ingress '%s' is not managed by prx and cannot be deleted", ingressName)
	}

	secrets, err := k.client.CoreV1().Secrets(namespace).Get(context.Background(), secret, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ingress: %v", err)
	}

	// Check if the "managed-by" label is present and equals "prx"
	if secrets.Labels == nil || secrets.Labels["managed-by"] != "prx" {
		return fmt.Errorf("secret '%s' is not managed by prx and cannot be deleted", ingressName)
	}

	if err := k.client.CoreV1().Secrets(namespace).Delete(context.Background(), secret, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("failed to get ingress: %v", err)
	}

	// Delete the ingress resource
	if err := k.client.NetworkingV1().Ingresses(namespace).Delete(context.Background(), ingressName, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("failed to delete ingress: %v", err)
	}
	return nil
}

func (k Kube) GetProxyMappings(namespace, configMapName string) (map[string]string, error) {

	res := make(map[string]string)

	cm, err := k.client.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap: %v", err)
	}

	data, ok := cm.Data["proxies.yaml"]
	if !ok {
		return nil, fmt.Errorf("proxies.yaml not found in configmap")
	}

	var mappings []ProxyMapping
	if err := yaml.Unmarshal([]byte(data), &mappings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proxy mappings: %v", err)
	}

	for _, v := range mappings {
		res[v.From] = v.To
	}

	return res, nil
}

func (k Kube) AddProxyMapping(namespace, configMapName string, newMapping ProxyMapping) error {
	cm, err := k.client.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get configmap: %v", err)
	}

	data, ok := cm.Data["proxies.yaml"]
	if !ok {
		data = ""
	}

	var mappings []ProxyMapping
	if data != "" {
		if err := yaml.Unmarshal([]byte(data), &mappings); err != nil {
			return fmt.Errorf("failed to unmarshal proxy mappings: %v", err)
		}
	}

	mappings = append(mappings, newMapping)

	updatedData, err := yaml.Marshal(mappings)
	if err != nil {
		return fmt.Errorf("failed to marshal updated mappings: %v", err)
	}

	cm.Data["proxies.yaml"] = string(updatedData)
	_, err = k.client.CoreV1().ConfigMaps(namespace).Update(context.Background(), cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update configmap: %v", err)
	}

	return nil
}

// DeleteProxyMapping removes a proxy mapping from the proxies.yaml file inside the specified ConfigMap.
// It identifies the mapping to be deleted by matching the 'From' field. If a mapping with the provided 'from' value
// is not found, the method returns an error.
func (k Kube) DeleteProxyMapping(namespace, configMapName, from string) error {
	// Get the ConfigMap that contains the proxy mappings.
	cm, err := k.client.CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get configmap: %v", err)
	}

	// Retrieve the proxies.yaml data from the ConfigMap.
	data, ok := cm.Data["proxies.yaml"]
	if !ok {
		return fmt.Errorf("proxies.yaml not found in configmap")
	}

	// Unmarshal the YAML data into a slice of ProxyMapping.
	var mappings []ProxyMapping
	if err := yaml.Unmarshal([]byte(data), &mappings); err != nil {
		return fmt.Errorf("failed to unmarshal proxy mappings: %v", err)
	}

	// Filter out the mapping whose 'From' field matches the provided parameter.
	updatedMappings := []ProxyMapping{}
	found := false
	for _, mapping := range mappings {
		if mapping.From == from {
			found = true
			continue // Skip the mapping to be deleted
		}
		updatedMappings = append(updatedMappings, mapping)
	}

	if !found {
		return fmt.Errorf("mapping with from '%s' not found", from)
	}

	// Marshal the updated slice back to YAML format.
	updatedData, err := yaml.Marshal(updatedMappings)
	if err != nil {
		return fmt.Errorf("failed to marshal updated proxy mappings: %v", err)
	}

	// Update the ConfigMap with the new proxies.yaml content.
	cm.Data["proxies.yaml"] = string(updatedData)
	_, err = k.client.CoreV1().ConfigMaps(namespace).Update(context.Background(), cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update configmap: %v", err)
	}

	return nil
}
