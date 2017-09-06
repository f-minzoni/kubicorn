// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/cutil/defaults"
	"github.com/kris-nova/kubicorn/cutil/logger"
)

var _ cloud.Resource = &StorageAccount{}

type StorageAccount struct {
	Shared
}

func (r *StorageAccount) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("storageAccount.Actual")

	newResource := &LoadBalancer{
		Shared: Shared{
			Tags: r.Tags,
		},
	}

	account, err := Sdk.StorageAccount.GetProperties(immutable.Name, immutable.Name)
	if err != nil {
		logger.Debug("Unable to lookup storage account: %v", err)
	} else {
		logger.Debug("Found storage account [%s]", *account.ID)
		newResource.Identifier = *account.ID
		newResource.Name = *account.Name
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *StorageAccount) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("storageAccount.Expected")
	newResource := &LoadBalancer{
		Shared: Shared{
			Tags:       r.Tags,
			Identifier: immutable.StorageIdentifier,
			Name:       immutable.Name,
		},
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *StorageAccount) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("storageAccount.Apply")
	applyResource := expected.(*StorageAccount)
	isEqual, err := compare.IsEqual(actual.(*StorageAccount), expected.(*StorageAccount))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	parameters := &storage.Account{}
	accountch, errch := Sdk.StorageAccount.Create(immutable.Name, immutable.Name, parameters, make(chan struct{}))
	account := <-accountch
	err = <-errch
	if err != nil {
		return nil, nil, err
	}
	logger.Info("Created storage account [%s]", immutable.Name)
	newResource := &StorageAccount{
		Shared: Shared{
			Tags: r.Tags,
			Identifier: *account.ID,
			Name: *account.Name,
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}
func (r *StorageAccount) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("storageAccount.Delete")
	deleteResource := actual.(*StorageAccount)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete VPC resource without ID [%s]", deleteResource.Name)
	}

	_, err  := Sdk.StorageAccount.Delete(immutable.Name, immutable.Name)
	if err != nil {
		return nil, nil, err
	}
	logger.Info("Deleted storage account [%s]", immutable.Name)
	newResource := &StorageAccount{
		Shared: Shared{
			Tags: r.Tags,
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *StorageAccount) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("storageAccount.Render")
	newCluster := defaults.NewClusterDefaults(inaccurateCluster)
	newCluster.StorageIdentifier = newResource.(*StorageAccount).Identifier
	return newCluster
}
