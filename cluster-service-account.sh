#!/bin/bash
set -e

# Variables
NAMESPACE="aproxynate"
SERVICE_ACCOUNT="prx-user"
ROLE_NAME="prx-user-role"
ROLE_BINDING="prx-user-rolebinding"
KUBECONFIG_OUTPUT="prx-kubeconfig.yaml"

# Create namespace if it doesn't exist
# echo "Creating namespace ${NAMESPACE}..."
# kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

# Create ServiceAccount in the namespace
echo "Creating ServiceAccount ${SERVICE_ACCOUNT} in namespace ${NAMESPACE}..."
kubectl create serviceaccount ${SERVICE_ACCOUNT} -n ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

# Create a Role that allows creating secrets and ingresses
echo "Creating Role ${ROLE_NAME} in namespace ${NAMESPACE}..."
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: ${ROLE_NAME}
  namespace: ${NAMESPACE}
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["create"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["create"]
EOF

# Create a RoleBinding to bind the Role to the ServiceAccount
echo "Creating RoleBinding ${ROLE_BINDING} in namespace ${NAMESPACE}..."
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ${ROLE_BINDING}
  namespace: ${NAMESPACE}
subjects:
- kind: ServiceAccount
  name: ${SERVICE_ACCOUNT}
  namespace: ${NAMESPACE}
roleRef:
  kind: Role
  name: ${ROLE_NAME}
  apiGroup: rbac.authorization.k8s.io
EOF

# Create a token for the ServiceAccount (requires Kubernetes v1.24+)
echo "Creating token for ServiceAccount ${SERVICE_ACCOUNT}..."
TOKEN=$(kubectl create token ${SERVICE_ACCOUNT} -n ${NAMESPACE})
if [ -z "$TOKEN" ]; then
    echo "Failed to create token for ${SERVICE_ACCOUNT}" >&2
    exit 1
fi

# Retrieve the current cluster's server URL and CA data
CLUSTER_SERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
CA_DATA=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')

# Generate the kubeconfig file
echo "Generating kubeconfig file ${KUBECONFIG_OUTPUT}..."
cat <<EOF > ${KUBECONFIG_OUTPUT}
apiVersion: v1
kind: Config
clusters:
- name: my-cluster
  cluster:
    server: ${CLUSTER_SERVER}
    certificate-authority-data: ${CA_DATA}
users:
- name: ${SERVICE_ACCOUNT}
  user:
    token: ${TOKEN}
contexts:
- name: prx-context
  context:
    cluster: my-cluster
    user: ${SERVICE_ACCOUNT}
    namespace: ${NAMESPACE}
current-context: prx-context
EOF

echo "Kubeconfig file created at ${KUBECONFIG_OUTPUT}"