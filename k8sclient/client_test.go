package k8sclient

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/util/retry"
)

func deploymentList(client v1.DeploymentInterface, t *testing.T) {

	_, err := client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Errorf("list error: %v", err)
	}

}
func deploymentCreate(client v1.DeploymentInterface, t *testing.T) {
	replicas := int32(2)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := client.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create error %s", err)
	}
	t.Logf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}
func deploymentUpdate(client v1.DeploymentInterface, t *testing.T) {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := client.Get(context.TODO(), "demo-deployment", metav1.GetOptions{})
		if getErr != nil {
			t.Fatalf("Failed to get latest version of Deployment: %s", getErr)

		}

		replicas := int32(1)
		result.Spec.Replicas = &replicas                             // reduce replica count
		result.Spec.Template.Spec.Containers[0].Image = "nginx:1.13" // change nginx version
		_, updateErr := client.Update(context.TODO(), result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		t.Fatalf("Update failed: %v", retryErr)
	}
	t.Log("Updated deployment...")
}

func TestNewK8SClient(t *testing.T) {

	clientset := NewK8SClient()
	deploymentClient := clientset.AppsV1().Deployments("default")
	deploymentCreate(deploymentClient, t)
	deploymentList(deploymentClient, t)
	deploymentUpdate(deploymentClient, t)

}
