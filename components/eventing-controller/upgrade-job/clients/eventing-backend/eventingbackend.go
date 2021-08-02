package eventingbackend

import (
	"context"
	"encoding/json"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
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

func (c Client) Get(namespace string, name string) (*eventingv1alpha1.EventingBackend, error) {

	ebUnstructured, err := c.dynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Get(
		context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}
	return toEventingBackend(ebUnstructured)
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "eventingbackends",
	}
}

func toEventingBackend(unstructured *unstructured.Unstructured) (*eventingv1alpha1.EventingBackend, error) {
	triggerList := new(eventingv1alpha1.EventingBackend)
	triggerListBytes, err := unstructured.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(triggerListBytes, triggerList)
	if err != nil {
		return nil, err
	}
	return triggerList, nil
}
