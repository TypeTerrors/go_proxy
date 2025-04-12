package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"prx/internal/models"
	
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Kube struct {
	client *kubernetes.Clientset
}

type ProxyMapping struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

func NewKubeClient() (Kube, error) {
	kubeconfigData := os.Getenv("PRX_KUBE_CONFIG")
	var config *rest.Config
	var err error

	if kubeconfigData != "" {
		config, err = clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigData))
		if err != nil {
			return Kube{}, fmt.Errorf("failed to build config from env PRX_KUBE_CONFIG: %v", err)
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

func (k Kube) AddNewProxy(body models.AddNewProxy, namespace string) error {

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

	_, err := k.client.CoreV1().Secrets("aproxynate").Create(context.Background(), secret, metav1.CreateOptions{})
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
											Name: "aproxynate",
											Port: networkingv1.ServiceBackendPort{
												Number: 443, // TODO: Replace with the actual service port.
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

	_, err = k.client.NetworkingV1().Ingresses("aproxynate").Create(context.Background(), ingress, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (k Kube) DeleteProxy(namespace, ingressName, secret string) error {
	ingress, err := k.client.NetworkingV1().Ingresses(namespace).Get(context.Background(), ingressName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ingress: %v", err)
	}

	// Check if the "managed-by" label is present and equals "prx"
	if ingress.Labels == nil || ingress.Labels["managed-by"] != "prx" {
		return fmt.Errorf("ingress '%s' is not managed by prx and cannot be deleted", ingressName)
	}

	secrets, err := k.client.CoreV1().Secrets(namespace).Get(context.Background(), "secretname", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ingress: %v", err)
	}

	// Check if the "managed-by" label is present and equals "prx"
	if secrets.Labels == nil || secrets.Labels["managed-by"] != "prx" {
		return fmt.Errorf("secret '%s' is not managed by prx and cannot be deleted", ingressName)
	}

	if err := k.client.CoreV1().Secrets(namespace).Delete(context.Background(), "secretname", metav1.DeleteOptions{}); err != nil {
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
	// If data is empty, initialize an empty slice.
	if data != "" {
		if err := yaml.Unmarshal([]byte(data), &mappings); err != nil {
			return fmt.Errorf("failed to unmarshal proxy mappings: %v", err)
		}
	}

	// Append the new mapping.
	mappings = append(mappings, newMapping)

	// Marshal back to YAML.
	updatedData, err := yaml.Marshal(mappings)
	if err != nil {
		return fmt.Errorf("failed to marshal updated mappings: %v", err)
	}

	// Update the ConfigMap.
	cm.Data["proxies.yaml"] = string(updatedData)
	_, err = k.client.CoreV1().ConfigMaps(namespace).Update(context.Background(), cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update configmap: %v", err)
	}

	return nil
}
