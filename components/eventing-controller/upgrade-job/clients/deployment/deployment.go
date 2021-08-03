package deployment

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Client struct {
	dynamicClient dynamic.Interface
}

func NewClient(client dynamic.Interface) Client {
	return Client{dynamicClient: client}
}

func (c Client) Get(namespace, name string) (*appsv1.Deployment, error) {
	unstructuredDeployment, err := c.dynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return toDeployment(unstructuredDeployment)
}

func (c Client) Update(namespace string, desiredDeployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	// Unmarshal from typed to unstructured
	data, err := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(desiredDeployment)
	if err != nil {
		return nil, err
	}
	unstructuredObj := &unstructured.Unstructured{
		Object: data,
	}

	unstructuredDeployment, err := c.dynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Update(context.TODO(), unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return toDeployment(unstructuredDeployment)
}

func (c Client) Delete(namespace, name string) error {
	err := c.dynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  appsv1.SchemeGroupVersion.Version,
		Group:    appsv1.SchemeGroupVersion.Group,
		Resource: "deployments",
	}
}

func toDeployment(unstructuredDeployment *unstructured.Unstructured) (*appsv1.Deployment, error) {
	deployment := new(appsv1.Deployment)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredDeployment.Object, deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}