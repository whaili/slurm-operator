// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	slurmclient "github.com/SlinkyProject/slurm-client/pkg/client"

	slinkyv1alpha1 "github.com/SlinkyProject/slurm-operator/api/v1alpha1"
	"github.com/SlinkyProject/slurm-operator/internal/builder"
	"github.com/SlinkyProject/slurm-operator/internal/controller/controller/slurmjwt"
	nodesetcontroller "github.com/SlinkyProject/slurm-operator/internal/controller/nodeset"
	"github.com/SlinkyProject/slurm-operator/internal/utils/objects"
)

type SyncStep struct {
	Name string
	Sync func(ctx context.Context, controller *slinkyv1alpha1.Controller) error
}

// Sync implements control logic for synchronizing a Controller.
func (r *ControllerReconciler) Sync(ctx context.Context, req reconcile.Request) error {
	logger := log.FromContext(ctx)

	controller := &slinkyv1alpha1.Controller{}
	if err := r.Get(ctx, req.NamespacedName, controller); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Controller has been deleted", "request", req)
			logger.V(1).Info("removed slurm client", "controller", req)
			_ = r.ClientMap.Remove(req.NamespacedName)
			return nil
		}
		return err
	}

	syncSteps := []SyncStep{
		{
			Name: "Service",
			Sync: func(ctx context.Context, controller *slinkyv1alpha1.Controller) error {
				object, err := r.builder.BuildControllerService(controller)
				if err != nil {
					return fmt.Errorf("failed to build: %w", err)
				}
				if err := objects.SyncObject(r.Client, ctx, object, false); err != nil {
					return fmt.Errorf("failed to sync object (%s): %w", klog.KObj(object), err)
				}
				return nil
			},
		},
		{
			Name: "Config",
			Sync: func(ctx context.Context, controller *slinkyv1alpha1.Controller) error {
				object, err := r.builder.BuildControllerConfig(controller)
				if err != nil {
					return fmt.Errorf("failed to build: %w", err)
				}
				if err := objects.SyncObject(r.Client, ctx, object, true); err != nil {
					return fmt.Errorf("failed to sync object (%s): %w", klog.KObj(object), err)
				}
				return nil
			},
		},
		{
			Name: "Scripts",
			Sync: func(ctx context.Context, controller *slinkyv1alpha1.Controller) error {
				object, err := r.builder.BuildControllerScripts(controller)
				if err != nil {
					return fmt.Errorf("failed to build: %w", err)
				}
				if err := objects.SyncObject(r.Client, ctx, object, true); err != nil {
					return fmt.Errorf("failed to sync object (%s): %w", klog.KObj(object), err)
				}
				return nil
			},
		},
		{
			Name: "StatefulSet",
			Sync: func(ctx context.Context, controller *slinkyv1alpha1.Controller) error {
				object, err := r.builder.BuildController(controller)
				if err != nil {
					return fmt.Errorf("failed to build: %w", err)
				}
				if err := objects.SyncObject(r.Client, ctx, object, true); err != nil {
					return fmt.Errorf("failed to sync object (%s): %w", klog.KObj(object), err)
				}
				return nil
			},
		},
		{
			Name: "Client",
			Sync: func(ctx context.Context, controller *slinkyv1alpha1.Controller) error {
				logger := log.FromContext(ctx)
				controllerKey := client.ObjectKeyFromObject(controller)

				server, err := r.getRestApiServer(ctx, controller)
				if err != nil {
					if apierrors.IsNotFound(err) {
						_ = r.ClientMap.Remove(controllerKey)
						return nil
					}
					return err
				}

				slurmClientOld := r.ClientMap.Get(controllerKey)
				if (slurmClientOld != nil) &&
					(slurmClientOld.GetServer() == server) {
					logger.V(1).Info("slurm client exists. Skipping...", "cluster", controllerKey.String())
					return nil
				}

				_ = r.ClientMap.Remove(controllerKey)

				signingKey, err := r.refResolver.GetSecretKeyRef(ctx, controller.AuthJwtHs256Ref(), controller.Namespace)
				if err != nil {
					return err
				}

				authToken, err := slurmjwt.NewToken(signingKey).NewSignedToken()
				if err != nil {
					return fmt.Errorf("failed to create Slurm auth token: %w", err)
				}

				config := &slurmclient.Config{
					Server:    server,
					AuthToken: authToken,
				}
				slurmClient, err := slurmclient.NewClient(config)
				if err != nil {
					return fmt.Errorf("failed to create slurm client: %w", err)
				}
				nodesetcontroller.SetEventHandler(slurmClient, r.EventCh)

				if r.ClientMap.Add(controllerKey, slurmClient) {
					logger.V(1).Info("added slurm client", "cluster", controllerKey.String())
				}

				return nil
			},
		},
	}

	for _, s := range syncSteps {
		if err := s.Sync(ctx, controller); err != nil {
			e := fmt.Errorf("[%s]: %w", s.Name, err)
			errors := []error{e}
			if err := r.syncStatus(ctx, controller); err != nil {
				e := fmt.Errorf("[%s]: %w", s.Name, err)
				errors = append(errors, e)
			}
			return utilerrors.NewAggregate(errors)
		}
	}

	return r.syncStatus(ctx, controller)
}

func (r *ControllerReconciler) getRestApiServer(ctx context.Context, controller *slinkyv1alpha1.Controller) (string, error) {
	logger := log.FromContext(ctx)

	restapiList, err := r.refResolver.GetRestapisForController(ctx, controller)
	if err != nil {
		return "", err
	}
	if len(restapiList.Items) == 0 {
		return "", errors.New(http.StatusText(http.StatusNotFound))
	}

	server := fmt.Sprintf("http://%s:%d", restapiList.Items[0].ServiceFQDNShort(), builder.SlurmrestdPort)
	if val := os.Getenv("DEBUG"); val == "1" {
		logger.Info("overriding restapi URL with localhost")
		server = fmt.Sprintf("http://localhost:%d", builder.SlurmrestdPort)
	}

	return server, nil
}
