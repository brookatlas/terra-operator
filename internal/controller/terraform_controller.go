/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	iacv1alpha1 "github.com/brookatlas/terraoperator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// TerraformReconciler reconciles a Terraform object
type TerraformReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=iac.terraoperator.com,resources=terraforms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=iac.terraoperator.com,resources=terraforms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=iac.terraoperator.com,resources=terraforms/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Terraform object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *TerraformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	terraform := &iacv1alpha1.Terraform{}
	err := r.Get(ctx, req.NamespacedName, terraform)
	// Check object exists
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Terraform resource not found. object should be deleted")
			return ctrl.Result{}, nil
		}
	}
	// Create batch job if not exists
	terraformJob := &batchv1.Job{}
	err = r.Get(ctx, types.NamespacedName{Name: terraform.Name, Namespace: terraform.Namespace}, terraformJob)
	if err != nil && errors.IsNotFound(err) {
		job := r.JobForTerraform(terraform)
		logger.Info("Creating a new Job for terraform", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		err = r.Create(ctx, job)
		if err != nil {
			logger.Error(err, "failed to create a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TerraformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&iacv1alpha1.Terraform{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}

func (r *TerraformReconciler) JobForTerraform(terraform *iacv1alpha1.Terraform) *batchv1.Job {
	//variablesArgument := mapToString(terraform.Spec.Variables)
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      terraform.Name,
			Namespace: terraform.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "iac.terraoperator.com/v1alpha1",
					Kind:       "Terraform",
					Name:       terraform.Name,
					UID:        terraform.UID,
				},
			},
			Labels: map[string]string{
				"createdby":               "terraoperator",
				"terraoperatorObjectName": terraform.Name,
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "terra-operator-runner",
							Image: "brookatlas/terra-operator:latest", // Terraform image
							Env: []v1.EnvVar{
								{
									Name:  "MODULEPATH",
									Value: terraform.Spec.ModulePath,
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}
	return &job
}

//
//func mapToString(m map[string]string) string {
//	var pairs []string
//
//	for key, value := range m {
//		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
//	}
//
//	return strings.Join(pairs, ", ")
//}
