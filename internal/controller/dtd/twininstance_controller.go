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

package dtd

import (
	"context"
	"fmt"

	twinevent "ktwin/operator/pkg/event"
	eventStore "ktwin/operator/pkg/event-store"
	twinservice "ktwin/operator/pkg/service"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dtdv0 "ktwin/operator/api/dtd/v0"
)

// TwinInstanceReconciler reconciles a TwinInstance object
type TwinInstanceReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	TwinService twinservice.TwinService
	TwinEvent   twinevent.TwinEvent
	EventStore  eventStore.EventStore
}

//+kubebuilder:rbac:groups=dtd.ktwin,resources=twininstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dtd.ktwin,resources=twininstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dtd.ktwin,resources=twininstances/finalizers,verbs=update

func (r *TwinInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	twinInstance := &dtdv0.TwinInstance{}
	err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, twinInstance)

	// Delete scenario
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, fmt.Sprintf("Unexpected error while deleting TwinInstance %s", req.Name))
		return ctrl.Result{}, err
	}

	// Get parent TwinInterface
	twinInterface := &dtdv0.TwinInterface{}
	twinInterfaceName := twinInstance.Spec.Interface
	err = r.Get(ctx, types.NamespacedName{Name: twinInterfaceName, Namespace: twinInstance.Namespace}, twinInterface)

	if err != nil {
		logger.Error(err, fmt.Sprintf("Unexpected error while getting TwinInterface %s", twinInterfaceName))
		return ctrl.Result{}, err
	}

	return r.createUpdateTwinInstance(ctx, req, twinInstance, twinInterface)
}

func (r *TwinInstanceReconciler) createUpdateTwinInstance(ctx context.Context, req ctrl.Request, twinInstance *dtdv0.TwinInstance, twinInterface *dtdv0.TwinInterface) (ctrl.Result, error) {
	twinInterfaceName := twinInstance.ObjectMeta.Name

	var resultErrors []error
	logger := log.FromContext(ctx)

	bindings := r.TwinEvent.GetMQQTDispatcherBindings(twinInstance)

	for _, binding := range bindings {
		err := r.Create(ctx, &binding, &client.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			logger.Error(err, fmt.Sprintf("Error while creating TwinInterface Binding %s", binding.Name))
			resultErrors = append(resultErrors, err)
		}
	}

	if len(resultErrors) > 0 {
		twinInstance.Status.Status = dtdv0.TwinInstancePhaseFailed
		return ctrl.Result{}, resultErrors[0]
	} else {
		twinInstance.Status.Status = dtdv0.TwinInstancePhaseRunning
	}

	twinInstance.Labels = map[string]string{
		"ktwin/twin-interface": twinInterfaceName,
	}

	// Update Status for Running or Failed
	_, err := r.updateTwinInstance(ctx, req, twinInstance)

	if err != nil {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *TwinInstanceReconciler) updateTwinInstance(ctx context.Context, req ctrl.Request, twinInstance *dtdv0.TwinInstance) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	err := r.Update(ctx, twinInstance, &client.UpdateOptions{})

	if err != nil {
		logger.Error(err, fmt.Sprintf("Error while updating TwinInstance %s", twinInstance.ObjectMeta.Name))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TwinInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dtdv0.TwinInstance{}).
		Complete(r)
}
