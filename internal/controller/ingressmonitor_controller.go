/*
Copyright 2023.

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
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	betterstackv1beta1 "everadaptive/betterstack/api/v1beta1"

	bsapi "github.com/everadaptive/betteruptime-go/api"
	netv1 "k8s.io/api/networking/v1"
)

// IngressMonitorReconciler reconciles a IngressMonitor object
type IngressMonitorReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	bsClient *bsapi.Client
}

//+kubebuilder:rbac:groups=betterstack.everadaptive.tech,resources=ingressmonitors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=betterstack.everadaptive.tech,resources=ingressmonitors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=betterstack.everadaptive.tech,resources=ingressmonitors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the IngressMonitor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *IngressMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	ingressMonitor := &betterstackv1beta1.IngressMonitor{}
	err := r.Get(ctx, req.NamespacedName, ingressMonitor)
	if err != nil {
		return ctrl.Result{}, nil
	}

	ingress := &netv1.Ingress{}
	err = r.Get(ctx, client.ObjectKey{
		Name:      ingressMonitor.Spec.IngressName,
		Namespace: req.Namespace,
	}, ingress)
	if err != nil {
		return ctrl.Result{}, err
	}

	status := make([]betterstackv1beta1.MonitorStatus, 0)
	updateResource := false

	myFinalizerName := "ingressmonitor.betterstack.everadaptive.tech"
	if ingressMonitor.ObjectMeta.DeletionTimestamp.IsZero() {
		monitorGroupName := fmt.Sprintf("🚀 %s/%s", ingress.Namespace, ingress.Name)
		if ingressMonitor.Status.MonitorGroup.ID != "" {
			resp, err := bsapi.MonitorGroupUpdate(ctx, r.bsClient, ingressMonitor.Status.MonitorGroup.ID, bsapi.MonitorGroup{
				Name: monitorGroupName,
			})

			if err != nil {
				log.Log.Error(err, "Error updating Better Stack monitor group", "name", monitorGroupName, "ID", ingressMonitor.Status.MonitorGroup.ID)
			}

			ingressMonitor.Status.MonitorGroup.ID = resp.Data.ID
		} else {
			resp, err := bsapi.MonitorGroupCreate(ctx, r.bsClient, bsapi.MonitorGroup{
				Name: monitorGroupName,
			})

			if err != nil {
				log.Log.Error(err, "Error creating Better Stack monitor group", "name", monitorGroupName)
			}

			ingressMonitor.Status.MonitorGroup.ID = resp.Data.ID
		}

		for _, rule := range ingress.Spec.Rules {
			monitorName := rule.Host

			monitorConfig := ingressMonitor.Spec.MonitorConfig
			monitorConfig.PronounceableName = monitorName
			monitorConfig.URL = rule.Host
			monitorConfig.MonitorGroupID, err = strconv.Atoi(ingressMonitor.Status.MonitorGroup.ID)
			if err != nil {
				log.Log.Error(err, "Error converting monnitor group id", "input", ingressMonitor.Status.MonitorGroup.ID)
				monitorConfig.MonitorGroupID = 0
			}

			monitor := bsapi.Monitor(monitorConfig)

			monitors, err := bsapi.MonitorGet(ctx, r.bsClient, fmt.Sprintf("pronounceable_name=%s", monitorName))
			if err != nil {
				log.Log.Error(err, "Error fetching monitors from BetterStack", "name", monitorName)
			}

			if len(monitors.Data) > 1 {
				log.Log.Error(nil, "Too many monitors retrieved from BetterStack", "name", monitorName)

			} else if len(monitors.Data) == 1 {
				resp, err := bsapi.MonitorUpdate(ctx, r.bsClient, monitors.Data[0].ID, monitor)

				if err != nil {
					log.Log.Error(err, "Error updating Better Stack monitor", "name", monitorName)
				}

				status = append(status, betterstackv1beta1.MonitorStatus{
					ID:          resp.Data.ID,
					Name:        resp.Data.Attributes.PronounceableName,
					MonitorType: resp.Data.Attributes.MonitorType,
					Paused:      resp.Data.Attributes.Paused,
				})
			} else if len(monitors.Data) == 0 {
				resp, err := bsapi.MonitorCreate(ctx, r.bsClient, monitor)

				if err != nil {
					log.Log.Error(err, "Error creating Better Stack monitor", "name", monitorName)
				}

				status = append(status, betterstackv1beta1.MonitorStatus{
					ID:          resp.Data.ID,
					Name:        resp.Data.Attributes.PronounceableName,
					MonitorType: resp.Data.Attributes.MonitorType,
					Paused:      resp.Data.Attributes.Paused,
				})
			}
		}

		if len(status) > 0 {
			ingressMonitor.Status = betterstackv1beta1.IngressMonitorStatus{
				Monitors: status,
			}
			r.Status().Update(context.Background(), ingressMonitor)
		}

		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !containsString(ingressMonitor.ObjectMeta.Finalizers, myFinalizerName) {
			ingressMonitor.ObjectMeta.Finalizers = append(ingressMonitor.ObjectMeta.Finalizers, myFinalizerName)
			updateResource = true
		}
	} else {
		// The object is being deleted
		if containsString(ingressMonitor.ObjectMeta.Finalizers, myFinalizerName) {
			// our finalizer is present, so lets handle our external dependency
			if err := r.deleteExternalDependency(ctx, ingressMonitor); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			ingressMonitor.ObjectMeta.Finalizers = removeString(ingressMonitor.ObjectMeta.Finalizers, myFinalizerName)
			updateResource = true
		}
	}

	if updateResource {
		if err := r.Update(context.Background(), ingressMonitor); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressMonitorReconciler) SetupWithManager(mgr ctrl.Manager, token string) error {
	client, err := bsapi.NewClient(token)
	if err != nil {
		return err
	}

	r.bsClient = client

	return ctrl.NewControllerManagedBy(mgr).
		For(&betterstackv1beta1.IngressMonitor{}).
		Complete(r)
}

func (r *IngressMonitorReconciler) deleteExternalDependency(ctx context.Context, instance *betterstackv1beta1.IngressMonitor) error {
	var err error

	for _, m := range instance.Status.Monitors {
		log.Log.Info("Deleting Better Stack monitor", "ID", m.ID)

		err = bsapi.MonitorDelete(ctx, r.bsClient, m.ID)
		if err != nil {
			log.Log.Error(err, "Error deleting Better Stack monitor")
		}
	}

	return err
}
