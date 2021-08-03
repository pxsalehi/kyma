package secret

import (
	"context"
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Client struct {
	dynamicClient dynamic.Interface
}

func NewClient(client dynamic.Interface) Client {
	return Client{dynamicClient: client}
}

func (c Client) ListByMatchingLabels(namespace string, labelSelector string) (*corev1.SecretList, error) {

	subscriptionsUnstructured, err := c.dynamicClient.Resource(GroupVersionResource()).Namespace(namespace).List(
		context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})

	if err != nil {
		return nil, err
	}
	return toSecretList(subscriptionsUnstructured)
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  corev1.SchemeGroupVersion.Version,
		Group:    corev1.SchemeGroupVersion.Group,
		Resource: "secrets",
	}
}

func toSecretList(unstructuredList *unstructured.UnstructuredList) (*corev1.SecretList, error) {
	triggerList := new(corev1.SecretList)
	triggerListBytes, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(triggerListBytes, triggerList)
	if err != nil {
		return nil, err
	}
	return triggerList, nil
}
