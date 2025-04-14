# go_proxy

## Overview

The proxy application offers a reverse proxy solution built for Kubernetes clusters. The proxy application redirects incoming HTTP requests based on user-defined proxy mappings. The application maintains proxy mapping records in memory and uses a Kubernetes ConfigMap as a persistent backup. The application creates Kubernetes Ingress resources and TLS Secrets to secure HTTPS traffic. In addition, the application employs JWT-based authentication to secure API endpoints.

## Key Features

- **Reverse Proxy Routing:**  
  The application accepts incoming HTTP requests and uses the Host header to determine the appropriate backend target from a list of proxy mapping records.

- **Proxy Mapping Management:**  
  The API endpoints support adding, updating, deleting, and retrieving proxy mapping records. Each mapping includes host information, target URL, TLS certificate, and private key. Records appear in the application’s in-memory store and are backed up in a Kubernetes ConfigMap.

- **TLS and Ingress Management:**  
  A new TLS Secret stores the certificate and key for each added proxy mapping. A corresponding Ingress resource uses the stored TLS Secret to secure incoming HTTPS connections on a defined host.

- **JWT-based Authentication:**  
  The application generates and validates JSON Web Tokens (JWTs). A JWT token secures API requests by validating the authorization header using the configured secret.

- **CI/CD Pipeline Integration:**  
  The GitHub Actions pipeline named **BuildPushDeploy** automates the building, tagging, pushing, and deployment of the proxy application container image to the Kubernetes cluster.

## Installation and Setup

### Prerequisites

- A Kubernetes cluster accessible via a kubeconfig file or through the environment variable `PRX_KUBE_CONFIG`.
- A working Go environment to build the application.
- Environment variables configured before running the application:
  - **NAMESPACE:** Kubernetes namespace for all application resources.
  - **JWT_SECRET:** Secret key used for JWT operations.
  - **PRX_KUBE_CONFIG:** (Optional) kubeconfig contents for cluster communication.

### Building the Application

1. Clone the Git repository containing the proxy application source code.
2. Build the application using the Go compiler:
   ```bash
   go build -o proxy-application
   ```
3. Set the required environment variables prior to running the application.

### Running in a Kubernetes Cluster

1. Package the application as a container image.
2. Deploy the container image in a Kubernetes Pod with proper environment variable injection.
3. Monitor the application log output for settings confirmation, including the generated JWT token and version information.

## API Endpoints and Their Functions

### Proxy Routing (Root Endpoint)

- **Purpose:**  
  Routes incoming HTTP requests based on the host header. The application retrieves target URLs from the in-memory proxy mapping store and the Kubernetes ConfigMap backup.

- **Behavior:**  
  The reverse proxy redirects requests to corresponding backend services after matching the incoming host with the stored records. Logs record details on incoming methods, hosts, and target URLs.

### GET `/api/prx`

- **Purpose:**  
  Retrieves all proxy mapping records from in-memory storage and the cluster-backed ConfigMap.

- **Behavior:**  
  A JSON array lists the proxy mapping records. Each record displays an originating host and its corresponding target URL. A status code of 200 (OK) appears upon successful retrieval.

### POST `/api/prx`

- **Purpose:**  
  Adds a new proxy mapping record, creates a new Kubernetes Ingress resource, and stores a TLS Secret.

- **Behavior:**  
  A JSON payload containing host, target URL, TLS certificate, and TLS key triggers the following actions:
  - The application creates a TLS Secret with the certificate and key.
  - An Ingress resource associates the TLS Secret with the host for secure HTTPS connections.
  - The proxy mapping record appears in the in-memory store and is included in the ConfigMap backup.
  - A successful creation returns a 201 (Created) status code.

### PATCH `/api/prx`

- **Purpose:**  
  Updates an existing proxy mapping record. The operation replaces the previous values, updating both the Ingress resource and its associated TLS Secret.

- **Behavior:**  
  A JSON payload similar to that of the POST endpoint instructs the application to:
  - Delete the existing Ingress resource and TLS Secret.
  - Create updated resources based on new values.
  - Update the in-memory record and the Kubernetes ConfigMap backup.
  - A successful update returns a 201 (Created) status code.

### DELETE `/api/prx`

- **Purpose:**  
  Removes an existing proxy mapping record along with its Kubernetes Ingress resource and TLS Secret.

- **Behavior:**  
  The DELETE operation accepts a JSON payload specifying the host for removal. The application deletes:
  - The proxy mapping record from memory.
  - The Kubernetes Ingress resource.
  - The TLS Secret from the cluster.
  - The ConfigMap backup reflects the removal, with a 201 (Created) status code confirming successful deletion.

### Status Endpoint

- **Purpose:**  
  Provides information on application health, current timestamp, and version data.

- **Behavior:**  
  A JSON object shows a status message (OK), the current time in RFC3339 format, and the version number. This endpoint supports monitoring and health checking.

## Storage and Backup

- **In-memory Storage:**  
  Proxy mapping records use a Go map structure for fast access during routing operations.

- **Kubernetes ConfigMap Backup:**  
  Proxy mapping records back up in a ConfigMap that stores a YAML file named `proxies.yaml`. This persistent storage allows records to survive application restarts and supports recovery processes.

## Ingress and TLS Secret Management

- **TLS Secret Creation:**  
  A new TLS Secret stores certificate data and the corresponding key. The secret uses the naming pattern `HOST-tls`, where HOST appears as part of the secret name.

- **Ingress Resource Creation:**  
  An Ingress resource applies the TLS Secret for HTTPS traffic management. The host in a proxy mapping record becomes the Ingress host. For each mapping, Kubernetes Ingress directs traffic from the specified host to the defined backend service on port 443.

### Pipeline Process

- **Trigger Conditions:**  
  The pipeline triggers when pull requests to the `main` or `dev` branches are closed.
  
- **Version Tagging:**  
  A short SHA and the pull request number compose the version tag used for naming the Docker image (formatted as `v_<SHA>.<PR_NUMBER>`).

- **Container Image Management:**  
  Steps include logging into the GitHub Container Registry, building the Docker image with build arguments, pushing the image, tagging an image as `latest`, and cleaning up local images.

- **Release Creation:**  
  A new GitHub release appears using the generated tag upon successful build and push operations.

- **Cluster Configuration and Deployment:**  
  Kubectl and Helm set up the deployment environment. The pipeline creates or updates a TLS Secret using the provided certificate and key from repository secrets. Finally, Helm deploys or upgrades the application in the designated namespace using custom values, including the image tag, JWT secret, and kubeconfig.

## Running the Application

### Environment Variables

Operators must set and verify the following environment variables before running the proxy application:

- **NAMESPACE:**  
  Specifies the Kubernetes namespace for Ingress resources, TLS Secrets, and ConfigMap backups.
  
- **JWT_SECRET:**  
  Used by the JWT service to generate and validate tokens for API authentication.
  
- **PRX_KUBE_CONFIG:**  
  Provides the Kubernetes configuration data needed for connecting to the cluster.

### Deployment Process

1. Build the application using the Go build process.
2. Package and deploy the application as a Docker container.
3. Use the provided GitHub Actions pipeline to automate image building, release creation, and deployment with Helm.
4. Verify logging output for version confirmation and JWT token details.
5. Monitor API endpoints and check Kubernetes resources (Ingresses, Secrets, ConfigMap) to ensure proper operation.

## Conclusion

The proxy application provides secure redirection of HTTP requests within Kubernetes using reverse proxy techniques. The application stores proxy mappings in both a fast in-memory store and a persistent ConfigMap. Kubernetes Ingress resources and TLS Secrets secure the traffic directed by the proxy mappings. Each API endpoint explicitly manages proxy records, and the GitHub Actions pipeline automates the build, container image push, and deployment process. Detailed logging and JWT-based authentication further enhance operations and security for developers and operators alike.