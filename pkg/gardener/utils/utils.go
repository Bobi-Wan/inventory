// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/inventory/pkg/clients/db"
	"github.com/gardener/inventory/pkg/gardener/models"
)

// GetSeedsFromDB fetches the [models.Seed] items from the database.
func GetSeedsFromDB(ctx context.Context) ([]models.Seed, error) {
	items := make([]models.Seed, 0)
	err := db.DB.NewSelect().Model(&items).Scan(ctx)

	return items, err
}

// GetProjectsFromDB fetches the [models.Project] items from the database.
func GetProjectsFromDB(ctx context.Context) ([]models.Project, error) {
	items := make([]models.Project, 0)
	err := db.DB.NewSelect().Model(&items).Scan(ctx)

	return items, err
}

// ErrCannotInferShoot is an error which is returned when a shoot cannot be
// inferred from the specified instance name.
var ErrCannotInferShoot = errors.New("cannot infer shoot")

// InferShootFromInstanceName infers the shoot from a Virtual Machine instance
// name.
//
// The GCP, AWS and Azure extension providers follow the same naming convention
// when creating a new Virtual Machine, which is:
//
// Convention: <shoot-namespace>-<worker-pool>-z<zone-index>-<pool-hash>-<vm-hash>
//
// Example: shoot--myproject--myshoot-mypool-z1-abcde-12345
//
// The <pool-hash> and <vm-hash> represent the first 5 bytes from a
// SHA-256 digest.
//
// Use this utility function to infer shoot details for Virtual Machines
// provisioned by the GCP, AWS, Azure or OpenStack extensions only.
func InferShootFromInstanceName(ctx context.Context, name string) (*models.Shoot, error) {
	pattern := regexp.MustCompile("^shoot--(?P<project>.*)--(?P<shoot_and_workerpool>.*)-z(?P<zone_index>.)-(?P<pool_hash>.{5})-(?P<vm_hash>.{5})$")
	matches := pattern.FindStringSubmatch(name)
	if len(matches) == 0 {
		return nil, ErrCannotInferShoot
	}

	// 5 groups + 1 for the whole instance name
	if len(matches) != 6 {
		return nil, ErrCannotInferShoot
	}

	// Lookup the shoot by using the project and worker prefix
	project := matches[pattern.SubexpIndex("project")]
	shootAndWorkerPool := matches[pattern.SubexpIndex("shoot_and_workerpool")]
	workerPrefix := fmt.Sprintf("shoot--%s--%s", project, shootAndWorkerPool)

	items := make([]models.Shoot, 0)
	err := db.DB.NewSelect().
		Model(&items).
		Where("project_name = ? AND array_position(worker_prefixes, ?) > 0", project, workerPrefix).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	switch {
	case len(items) == 0:
		return nil, ErrCannotInferShoot
	case len(items) > 1:
		return nil, fmt.Errorf("%w: multiple shoots match", ErrCannotInferShoot)
	default:
		return &items[0], nil
	}
}

// Decode takes a `decoder` and decodes the provided `data` into the provided object.
// The underlying `into` address is used to assign the decoded object.
func Decode(decoder runtime.Decoder, data []byte, into runtime.Object) error {
	// By not providing an `into` it is necessary that the serialized `data` is configured with
	// a proper `apiVersion` and `kind` field. This also makes sure that the conversion logic to
	// the internal version is called.
	output, _, err := decoder.Decode(data, nil, nil)
	if err != nil {
		return err
	}

	intoType := reflect.TypeOf(into)

	if reflect.TypeOf(output) == intoType {
		reflect.ValueOf(into).Elem().Set(reflect.ValueOf(output).Elem())

		return nil
	}

	return fmt.Errorf("is not of type %s", intoType)
}
