package scenario

import (
	threatestergithubiov1alpha1 "github.com/mrtc0/threatester/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type ScenarioBuilder struct {
	namespace string
	podSpec   corev1.PodSpec
}

func NewScenarioJobBuilder() *ScenarioBuilder {
	return &ScenarioBuilder{
		podSpec: corev1.PodSpec{Containers: []corev1.Container{}},
	}
}

func (b *ScenarioBuilder) WithScenarioJobs(templates []threatestergithubiov1alpha1.Template) *ScenarioBuilder {
	for _, template := range templates {
		b.podSpec.Containers = append(b.podSpec.Containers, *template.Container)
	}

	b.podSpec.RestartPolicy = corev1.RestartPolicyNever
	return b
}

func (b *ScenarioBuilder) WithNamespace(namespace string) *ScenarioBuilder {
	b.namespace = namespace
	return b
}

func (b *ScenarioBuilder) Build() (*batchv1.Job, error) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "threatester-scenario",
			Namespace: b.namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: pointer.Int32(0),
			Completions:  pointer.Int32(1),
			Template: corev1.PodTemplateSpec{
				Spec: b.podSpec,
			},
		},
	}

	return job, nil
}
