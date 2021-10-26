/*
Copyright 2021.

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

package controllers

import (
	"context"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	jobv1alpha1 "github.com/nuka137/deadline-job-operator/api/v1alpha1"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/api/errors"
)

var logger = logf.Log.WithName("controller_deadlinejob")

// DeadlineJobReconciler reconciles a DeadlineJob object
type DeadlineJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func timeUntilSchedule(schedule string) (time.Duration, error) {
	now := time.Now().UTC()
	layout := time.RFC3339
	s, err := time.Parse(layout, schedule)
	if err != nil {
		return time.Duration(0), err
	}
	return s.Sub(now), nil
}

func newPod(cr *jobv1alpha1.DeadlineJob) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: strings.Split(cr.Spec.Command, " "),
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
	return pod
}

//+kubebuilder:rbac:groups=job.nuka137.com,resources=deadlinejobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=job.nuka137.com,resources=deadlinejobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=job.nuka137.com,resources=deadlinejobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DeadlineJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *DeadlineJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	reqLogger := logger.WithValues("namespace", req.Namespace, "DeadlineJob", req.Name)
	reqLogger.Info("=== Reconcile DeadlineJob")

	instance := &jobv1alpha1.DeadlineJob{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if instance.Status.Phase == "" {
		instance.Status.Phase = jobv1alpha1.PhasePending
	}

	switch instance.Status.Phase {
	case jobv1alpha1.PhasePending:

		reqLogger.Info("Phase: PENDING")
		d, err := timeUntilSchedule(instance.Spec.JobStart)
		if err != nil {
			reqLogger.Error(err, "Schedule parsing failure")
			return ctrl.Result{}, err
		}
		if d > 0 {
			reqLogger.Info("Not yet to time to execute the job", "rest", d)
			return ctrl.Result{RequeueAfter: d}, nil
		}
		reqLogger.Info("It's time to execute the job.")
		instance.Status.Phase = jobv1alpha1.PhaseRunning

	case jobv1alpha1.PhaseRunning:

		reqLogger.Info("Phase: RUNNING")

		pod := newPod(instance)
		if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		podInstance := &corev1.Pod{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, podInstance)
		if err != nil && errors.IsNotFound(err) {
			err = r.Create(context.TODO(), pod)
			if err != nil {
				return ctrl.Result{}, err
			}
			reqLogger.Info("Pod launched:", "name", pod.Name)
		} else if err != nil {
			return ctrl.Result{}, err
		} else if podInstance.Status.Phase == corev1.PodFailed || podInstance.Status.Phase == corev1.PodSucceeded {
			reqLogger.Info("Pod terminated", "reason", podInstance.Status.Reason, "message", podInstance.Status.Message)
			instance.Status.Phase = jobv1alpha1.PhaseDone
		} else {

			d, err := timeUntilSchedule(instance.Spec.JobEnd)
			if err != nil {
				reqLogger.Error(err, "Schedule parsing failure")
				return ctrl.Result{}, err
			}
			if d > 0 {
				reqLogger.Info("Not yet to time to finish the job", "rest", d)
				return ctrl.Result{RequeueAfter: d}, nil
			}

			err = r.Delete(context.TODO(), podInstance)
			if err != nil {
				return ctrl.Result{}, err
			}
			err = r.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, podInstance)
			if err != nil {
				return ctrl.Result{}, err
			}

			instance.Status.Phase = jobv1alpha1.PhaseExceededDeadline
		}

	case jobv1alpha1.PhaseDone:

		reqLogger.Info("Phase: DONE")
		return ctrl.Result{}, nil

	case jobv1alpha1.PhaseExceededDeadline:

		reqLogger.Info("Phase: EXCEEDED_DEADLINE")
		return ctrl.Result{}, nil

	default:

		reqLogger.Info("NOP")
		return ctrl.Result{}, nil

	}

	err = r.Status().Update(context.TODO(), instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeadlineJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jobv1alpha1.DeadlineJob{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
