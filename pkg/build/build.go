package build

import (
	"os"
)

func GetBuildMethod() string {
	kubernetesServiceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	kubernetesServicePort := os.Getenv("KUBERNETES_SERVICE_PORT")
	if kubernetesServiceHost == "" && kubernetesServicePort == "" {
		return "container"
	} else {
		return "kubernetes"
	}
}
