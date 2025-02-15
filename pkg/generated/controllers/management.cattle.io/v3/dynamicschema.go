/*
Copyright 2023 Rancher Labs, Inc.

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

// Code generated by main. DO NOT EDIT.

package v3

import (
	"context"
	"time"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DynamicSchemaController interface for managing DynamicSchema resources.
type DynamicSchemaController interface {
	generic.NonNamespacedControllerInterface[*v3.DynamicSchema, *v3.DynamicSchemaList]
}

// DynamicSchemaClient interface for managing DynamicSchema resources in Kubernetes.
type DynamicSchemaClient interface {
	generic.NonNamespacedClientInterface[*v3.DynamicSchema, *v3.DynamicSchemaList]
}

// DynamicSchemaCache interface for retrieving DynamicSchema resources in memory.
type DynamicSchemaCache interface {
	generic.NonNamespacedCacheInterface[*v3.DynamicSchema]
}

type DynamicSchemaStatusHandler func(obj *v3.DynamicSchema, status v3.DynamicSchemaStatus) (v3.DynamicSchemaStatus, error)

type DynamicSchemaGeneratingHandler func(obj *v3.DynamicSchema, status v3.DynamicSchemaStatus) ([]runtime.Object, v3.DynamicSchemaStatus, error)

func RegisterDynamicSchemaStatusHandler(ctx context.Context, controller DynamicSchemaController, condition condition.Cond, name string, handler DynamicSchemaStatusHandler) {
	statusHandler := &dynamicSchemaStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, generic.FromObjectHandlerToHandler(statusHandler.sync))
}

func RegisterDynamicSchemaGeneratingHandler(ctx context.Context, controller DynamicSchemaController, apply apply.Apply,
	condition condition.Cond, name string, handler DynamicSchemaGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &dynamicSchemaGeneratingHandler{
		DynamicSchemaGeneratingHandler: handler,
		apply:                          apply,
		name:                           name,
		gvk:                            controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterDynamicSchemaStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type dynamicSchemaStatusHandler struct {
	client    DynamicSchemaClient
	condition condition.Cond
	handler   DynamicSchemaStatusHandler
}

func (a *dynamicSchemaStatusHandler) sync(key string, obj *v3.DynamicSchema) (*v3.DynamicSchema, error) {
	if obj == nil {
		return obj, nil
	}

	origStatus := obj.Status.DeepCopy()
	obj = obj.DeepCopy()
	newStatus, err := a.handler(obj, obj.Status)
	if err != nil {
		// Revert to old status on error
		newStatus = *origStatus.DeepCopy()
	}

	if a.condition != "" {
		if errors.IsConflict(err) {
			a.condition.SetError(&newStatus, "", nil)
		} else {
			a.condition.SetError(&newStatus, "", err)
		}
	}
	if !equality.Semantic.DeepEqual(origStatus, &newStatus) {
		if a.condition != "" {
			// Since status has changed, update the lastUpdatedTime
			a.condition.LastUpdated(&newStatus, time.Now().UTC().Format(time.RFC3339))
		}

		var newErr error
		obj.Status = newStatus
		newObj, newErr := a.client.UpdateStatus(obj)
		if err == nil {
			err = newErr
		}
		if newErr == nil {
			obj = newObj
		}
	}
	return obj, err
}

type dynamicSchemaGeneratingHandler struct {
	DynamicSchemaGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
}

func (a *dynamicSchemaGeneratingHandler) Remove(key string, obj *v3.DynamicSchema) (*v3.DynamicSchema, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v3.DynamicSchema{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

func (a *dynamicSchemaGeneratingHandler) Handle(obj *v3.DynamicSchema, status v3.DynamicSchemaStatus) (v3.DynamicSchemaStatus, error) {
	if !obj.DeletionTimestamp.IsZero() {
		return status, nil
	}

	objs, newStatus, err := a.DynamicSchemaGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}

	return newStatus, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
}
