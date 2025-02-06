package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/template"

	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicinformer "k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	var config *rest.Config
	var err error

	if *kubeconfig == "" {
		log.Println("kubeconfig is empty, using in-cluster config")
		config, err = rest.InClusterConfig()
	} else {
		log.Printf("using kubeconfig: %s", *kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}

	if err != nil {
		log.Println("failed to get kubeconfig", err)
		return
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating dynamic client: %s\n", err.Error())
		os.Exit(1)
	}

	startInformer(dynamicClient)
}

func startInformer(dynamicClient dynamic.Interface) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gvr := schema.GroupVersionResource{
		Group:    "akri.sh",
		Version:  "v0",
		Resource: "instances",
	}

	// Create DynamicSharedInformerFactory
	factory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 0)
	informer := factory.ForResource(gvr).Informer()

	// Add event handlers
	if _, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			instance := obj.(*unstructured.Unstructured)
			log.Printf("New Instance Added: %s\n", instance.GetName())
			processInstance(dynamicClient, *instance, Create)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			instance := newObj.(*unstructured.Unstructured)
			log.Printf("Instance Updated: %s\n", instance.GetName())
			processInstance(dynamicClient, *instance, Update)
		},
		DeleteFunc: func(obj interface{}) {
			instance := obj.(*unstructured.Unstructured)
			log.Printf("Instance Deleted: %s\n", instance.GetName())
			processInstance(dynamicClient, *instance, Delete)
		},
	}); err != nil {
		log.Printf("Error adding event handler: %v", err)
		return
	}

	// Start Informer
	go informer.Run(ctx.Done())

	// Capture system signals to gracefully shut down Informer
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	signal.Stop(sigCh) // stop further signal notifications
	cancel()
	log.Println("Informer stopped")
}

type Action uint8

const (
	Create Action = iota
	Update
	Delete
)

func processInstance(dynamicClient dynamic.Interface, instance unstructured.Unstructured, action Action) {
	data := extractInstanceDetails(instance)
	if err := validateInstanceDetails(data); err != nil {
		log.Printf("Invalid instance details: %v", err)
		return
	}

	templatesDir := getTemplatesDir(data.Protocol)
	if templatesDir == "" {
		log.Printf("Unsupported protocol: %s", data.Protocol)
		return
	}

	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		log.Printf("Templates directory not found: %s", templatesDir)
		return
	}

	files := []string{
		filepath.Join(templatesDir, "edgedevice.yaml.tpl"),
		filepath.Join(templatesDir, "deviceshifu-deployment.yaml.tpl"),
		filepath.Join("templates", "deviceshifu-service.yaml.tpl"),
	}

	for _, file := range files {
		if err := renderAndApplyTemplate(dynamicClient, file, *data, action); err != nil {
			log.Printf("Failed to process template %s: %v", file, err)
		}
	}
}

func validateInstanceDetails(data *InstanceDetail) error {
	if data == nil {
		return fmt.Errorf("instance details is nil")
	}
	if data.Name == "" || data.Namespace == "" {
		return fmt.Errorf("name or namespace is empty")
	}
	if data.ConfigmapName == "" {
		return fmt.Errorf("configmap name is empty")
	}
	return nil
}

func getTemplatesDir(protocol string) string {
	protocolDirs := map[string]string{
		"opcua": "templates/opcua",
		"RTSP":  "templates/rtsp",
		"S7":    "templates/s7",
	}
	dir := protocolDirs[protocol]
	if dir == "" {
		log.Printf("No templates path found for protocol: %s", protocol)
	}
	return dir
}

type InstanceDetail struct {
	Name          string
	Namespace     string
	Address       string
	Protocol      string
	SKU           string
	ConfigmapName string
	Properties    map[string]interface{}
}

func extractInstanceDetails(instance unstructured.Unstructured) *InstanceDetail {
	instanceName := instance.GetName()
	namespace := instance.GetNamespace()

	spec, found, err := unstructured.NestedMap(instance.Object, "spec")
	if err != nil || !found {
		log.Printf("Error extracting spec from instance %s: %v", instanceName, err)
		return nil
	}

	brokerProperties, found, err := unstructured.NestedMap(spec, "brokerProperties")
	if err != nil || !found {
		log.Printf("Error extracting brokerProperties from spec in instance %s: %v", instanceName, err)
		return nil
	}

	protocol, address, found := detectProtocol(brokerProperties)
	if !found {
		log.Printf("Error detecting protocol and address for instance %s", instanceName)
		protocol = "UNKNOWN"
		address = ""
	}

	configmapName, ok := brokerProperties["DEVICESHIFU_CONFIG"].(string)
	if !ok {
		log.Printf("Error extracting configmapName from brokerProperties in instance %s", instanceName)
		return nil
	}

	return &InstanceDetail{
		Name:          instanceName,
		Namespace:     namespace,
		Address:       address,
		Protocol:      protocol,
		SKU:           instanceName,
		ConfigmapName: configmapName,
		Properties:    brokerProperties,
	}
}

func detectProtocol(brokerProperties map[string]interface{}) (string, string, bool) {
	// First check explicit protocol and endpoint
	if protocol, ok := brokerProperties["Protocol"].(string); ok {
		if endpoint, ok := brokerProperties["Endpoint"].(string); ok {
			return protocol, endpoint, true
		}
	}

	// Then check known protocol patterns
	protocolPatterns := []struct {
		addressKey string
		protocol   string
	}{
		{"OPCUA_DISCOVERY_URL", "opcua"},
		{"HTTP_URL", "http"},
		{"RTSP_URL", "RTSP"},
		{"S7_URL", "S7"},
	}

	for _, pattern := range protocolPatterns {
		if address, ok := brokerProperties[pattern.addressKey].(string); ok {
			return pattern.protocol, address, true
		}
	}

	return "", "", false
}

func renderAndApplyTemplate(dynamicClient dynamic.Interface, templateFile string, data InstanceDetail, action Action) error {
	t, err := template.ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("Error parsing template file %s: %v", templateFile, err)
	}

	var renderedTemplate bytes.Buffer
	err = t.Execute(&renderedTemplate, data)
	if err != nil {
		return fmt.Errorf("Error rendering template: %v", err)
	}

	switch action {
	case Create, Update:
		return applyYAML(dynamicClient, renderedTemplate.Bytes())
	case Delete:
		return deleteYaml(dynamicClient, renderedTemplate.Bytes())
	default:
		return fmt.Errorf("Unsupported action: %d", action)
	}
}

func applyYAML(dynamicClient dynamic.Interface, yamlContent []byte) error {
	var obj unstructured.Unstructured
	if err := yaml.Unmarshal(yamlContent, &obj); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    obj.GetObjectKind().GroupVersionKind().Group,
		Version:  obj.GetObjectKind().GroupVersionKind().Version,
		Resource: getResourcePlural(&obj),
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, err := dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Create(
			context.Background(),
			&obj,
			v1.CreateOptions{},
		)
		return err
	})
}

func deleteYaml(dynamicClient dynamic.Interface, yamlContent []byte) error {
	var obj unstructured.Unstructured
	if err := yaml.Unmarshal(yamlContent, &obj); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    obj.GetObjectKind().GroupVersionKind().Group,
		Version:  obj.GetObjectKind().GroupVersionKind().Version,
		Resource: getResourcePlural(&obj),
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Delete(
			context.Background(),
			obj.GetName(),
			v1.DeleteOptions{},
		)
	})
}

func getResourcePlural(obj *unstructured.Unstructured) string {
	resourcePlural, _ := meta.UnsafeGuessKindToResource(obj.GroupVersionKind())
	return resourcePlural.Resource
}
