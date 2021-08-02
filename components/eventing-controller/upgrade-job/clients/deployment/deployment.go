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

//func (c Client) Patch(namespace, name string, data []byte) (*appsv1.Deployment, error) {
//	unstructuredDeployment, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Patch(name, types.MergePatchType, data, metav1.PatchOptions{})
//	if err != nil {
//		return nil, err
//	}
//	return toDeployment(unstructuredDeployment)
//}