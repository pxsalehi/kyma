package subscription

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

func (c Client) List(namespace string) (*eventingv1alpha1.SubscriptionList, error) {

	subscriptionsUnstructured, err := c.dynamicClient.Resource(GroupVersionResource()).Namespace(namespace).List(
		context.TODO(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}
	return toSubscriptionList(subscriptionsUnstructured)
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
}

func toSubscriptionList(unstructuredList *unstructured.UnstructuredList) (*eventingv1alpha1.SubscriptionList, error) {
	triggerList := new(eventingv1alpha1.SubscriptionList)
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
