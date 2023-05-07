package controller

import (
	"context"
	"fmt"
	"time"

	threatestergithubiov1alpha1 "github.com/mrtc0/threatester/api/v1alpha1"
	"github.com/mrtc0/threatester/internal/application/expectation"
	scenarioApplication "github.com/mrtc0/threatester/internal/application/scenario"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Scenario controller", func() {
	Context("Scenario controller test", func() {
		ctx := context.Background()
		const scenarioName = "test-scenario"
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "threatester-test",
			},
		}

		BeforeEach(func() {
			By("Creating the namespace for tests")
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			By("Deleting the namespace for tests")
			_ = k8sClient.Delete(ctx, namespace)
		})

		It("Should successfully reconcile a custom resource for threatester scenario", func() {
			scenario := &threatestergithubiov1alpha1.Scenario{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: scenarioName, Namespace: namespace.Name}, scenario)
			if err != nil && errors.IsNotFound(err) {
				scenario := &threatestergithubiov1alpha1.Scenario{
					ObjectMeta: metav1.ObjectMeta{
						Name:      scenarioName,
						Namespace: namespace.Name,
					},
					Spec: threatestergithubiov1alpha1.ScenarioSpec{
						Templates: []threatestergithubiov1alpha1.Template{
							{
								Name: "test",
								Container: &corev1.Container{
									Name:    "test",
									Image:   "alpine",
									Command: []string{"echo", "hello"},
								},
							},
						},
						Expectations: []threatestergithubiov1alpha1.Expectation{
							{
								Timeout: "10s",
								Datadog: &threatestergithubiov1alpha1.DatadogExpectation{
									Monitor: &threatestergithubiov1alpha1.DatadogMonitor{
										ID:     "123456",
										Status: "Alert",
									},
								},
							},
						},
					},
				}

				err = k8sClient.Create(ctx, scenario)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the scenario was created")
			Eventually(func() error {
				found := &threatestergithubiov1alpha1.Scenario{}
				return k8sClient.Get(ctx, types.NamespacedName{Name: scenarioName, Namespace: namespace.Name}, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			scenarioReconciler := &ScenarioReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				ExpectationService: &expectation.ExpectationServiceMock{
					RunExpectationFunc: func(ctx context.Context) (bool, error) {
						return true, nil
					},
					SetExpectationsFunc: func(expectations []threatestergithubiov1alpha1.Expectation) {
					},
				},
				ScenarioJobExecutor: &scenarioApplication.ScenarioJobExecutorMock{
					ExecuteFunc: func(ctx context.Context, scenarioJob batchv1.Job) error {
						return nil
					},
					DeleteScenarioJobFunc: func(ctx context.Context, scenarioJob batchv1.Job) error {
						return nil
					},
				},
			}

			_, err = scenarioReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: scenarioName, Namespace: namespace.Name},
			})
			Expect(err).To(Not(HaveOccurred()))

			Eventually(func() error {
				found := &threatestergithubiov1alpha1.Scenario{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: scenarioName, Namespace: namespace.Name}, found)

				Expect(err).To(Not(HaveOccurred()))

				latestStatusCondition := found.Status.Conditions[len(found.Status.Conditions)-1]
				expectLatestStatusCondition := metav1.Condition{
					Type:    typeSucceededScenario,
					Status:  metav1.ConditionTrue,
					Reason:  "Success",
					Message: "Successfully run scenario expectations",
				}

				if latestStatusCondition.Status != expectLatestStatusCondition.Status {
					return fmt.Errorf("expected status %#v but got %#v", expectLatestStatusCondition.Status, latestStatusCondition.Status)
				}

				return nil
			}, time.Minute, time.Second).Should(Succeed())
		})
	})
})
