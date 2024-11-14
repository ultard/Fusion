APP_NAME=fusion-backend
DOCKER_IMAGE_NAME=fusion-backend
DOCKER_TAG=latest
K8S_NAMESPACE=default
HELM_CHART_DIR=./k8s
DOCKERFILE_PATH=./Dockerfile
K8S_DEPLOYMENT_FILE=./k8s/templates/deploy.yaml
K8S_DEPLOYMENT_NAME=fusion-backend-deployment

docker:
	docker compose up -d

docker-build:
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) -f $(DOCKERFILE_PATH) .

# Helm команды для Kubernetes
helm-create:
	helm install $(APP_NAME) $(HELM_CHART_DIR) --namespace $(K8S_NAMESPACE)

helm-install:
	helm upgrade --install $(APP_NAME) $(HELM_CHART_DIR) --namespace $(K8S_NAMESPACE)

helm-upgrade:
	helm upgrade $(APP_NAME) $(HELM_CHART_DIR) --namespace $(K8S_NAMESPACE)

helm-uninstall:
	helm uninstall $(APP_NAME) --namespace $(K8S_NAMESPACE)

# Kubernetes команды
kubectl-deploy:
	kubectl apply -f $(K8S_DEPLOYMENT_FILE) --namespace $(K8S_NAMESPACE)

kubectl-delete:
	kubectl delete -f $(K8S_DEPLOYMENT_FILE) --namespace $(K8S_NAMESPACE)

kubectl-logs:
	kubectl logs -f $(K8S_DEPLOYMENT_NAME) --namespace $(K8S_NAMESPACE)