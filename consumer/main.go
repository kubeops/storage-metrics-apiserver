package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	v1beta2 "k8s.io/metrics/pkg/apis/custom_metrics/v1beta2"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	httpClient, err := rest.HTTPClientFor(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating HTTP client: %v\n", err)
		os.Exit(1)
	}

	url := cfg.Host + "/apis/custom.metrics.k8s.io/v1beta2/namespaces/default/persistentvolumeclaims/*/volume_used_percentage"

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error building request: %v\n", err)
		os.Exit(1)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error querying metric: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var list v1beta2.MetricValueList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		fmt.Fprintf(os.Stderr, "error decoding response: %v\n", err)
		os.Exit(1)
	}

	if len(list.Items) == 0 {
		fmt.Println("no metrics found")
		return
	}

	fmt.Printf("%-30s %s\n", "PVC", "volume_used_percentage")
	fmt.Println("----------------------------------------------")
	for _, item := range list.Items {
		fmt.Printf("%-30s %s\n", item.DescribedObject.Name, item.Value.String())
	}
}

func loadConfig() (*rest.Config, error) {
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
