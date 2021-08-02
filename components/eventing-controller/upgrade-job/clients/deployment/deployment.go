package deployment

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Client struct {
	clientset *kubernetes.Clientset
}

func NewClient(clientset *kubernetes.Clientset) Client {
	return Client{clientset: clientset}
}

func (c Client) Get(namespace, name string) (*appsv1.Deployment, error) {
	deploymentsClient := c.clientset.AppsV1().Deployments(namespace)
	deployment, err := deploymentsClient.Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func (c Client) Update(namespace string, desiredDeployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	deploymentsClient := c.clientset.AppsV1().Deployments(namespace)
	result, err := deploymentsClient.Update(context.TODO(), desiredDeployment, metav1.UpdateOptions{})

	if err != nil {
		return nil, err
	}
	return result, nil
}