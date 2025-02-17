/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package partition_topology

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/partition"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/partition_topology_db/partition_connection_overrides"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/partition_topology_db/partition_services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/partition_topology_db/service_partitions"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
)

const (
	DefaultPartitionId          = service_network_types.PartitionID("default")
	partitionNotFoundForService = ""
)

// Stores the partition topology of the network, and exposes an API for modifying it
type PartitionTopology struct {
	lock *sync.RWMutex

	defaultConnection PartitionConnection

	servicePartitions *service_partitions.ServicePartitionsBucket

	// By default, connection between 2 partitions is set to defaultConnection. This map contains overrides
	partitionConnectionOverrides *partition_connection_overrides.PartitionConnectionOverridesBucket

	// A service can be a part of exactly one partition at a time
	partitionServices *partition_services.PartitionServicesBucket
}

func NewPartitionTopology(defaultPartition service_network_types.PartitionID, defaultConnection PartitionConnection, enclaveDb *enclave_db.EnclaveDB) (*PartitionTopology, error) {
	partitionServicesBucket, err := partition_services.GetOrCreatePartitionServicesBucket(enclaveDb, partition.PartitionID(defaultPartition))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting the partition services bucket")
	}

	servicePartitionsBucket, err := service_partitions.GetOrCreateServicePartitionsBucket(enclaveDb)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the service partitions bucket")
	}

	partitionConnectionOverridesBucket, err := partition_connection_overrides.GetOrCreatePartitionConnectionOverrideBucket(enclaveDb)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the partition connection overrides bucket")
	}

	return &PartitionTopology{
		lock:                         &sync.RWMutex{},
		servicePartitions:            servicePartitionsBucket,
		partitionServices:            partitionServicesBucket,
		partitionConnectionOverrides: partitionConnectionOverridesBucket,
		defaultConnection:            defaultConnection,
	}, nil
}

func newPartitionTopologyForTesting(
	enclaveDb *enclave_db.EnclaveDB,
	defaultPartition service_network_types.PartitionID,
	defaultConnection PartitionConnection,
	partitionServices map[service_network_types.PartitionID][]service.ServiceName,
	partitionConnectionOverrides map[service_network_types.PartitionConnectionID]PartitionConnection,
) (*PartitionTopology, error) {
	partitionServicesBucket, err := partition_services.GetOrCreatePartitionServicesBucket(enclaveDb, partition.PartitionID(defaultPartition))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting the partition services bucket")
	}

	servicePartitionsBucket, err := service_partitions.GetOrCreateServicePartitionsBucket(enclaveDb)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the service partitions bucket")
	}

	for partitionId, serviceNames := range partitionServices {
		if err := partitionServicesBucket.AddServicesToPartition(partition.PartitionID(partitionId), map[service.ServiceName]bool{}); err != nil {
			return nil, stacktrace.Propagate(err, "Unable to create partition '%s'", partitionId)
		}
		for _, serviceName := range serviceNames {
			if err = servicePartitionsBucket.AddPartitionToService(serviceName, partition.PartitionID(partitionId)); err != nil {
				return nil, stacktrace.Propagate(err, "Unable to initialize partition to service bucket")
			}
			if err = partitionServicesBucket.AddServiceToPartition(partition.PartitionID(partitionId), serviceName); err != nil {
				return nil, stacktrace.Propagate(err, "Unable to initialize service to partition bucket")
			}
		}
	}

	partitionConnectionOverridesBucket, err := partition_connection_overrides.GetOrCreatePartitionConnectionOverrideBucket(enclaveDb)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the partition connection overrides bucket")
	}

	for partitionConnectionId, partitionConnection := range partitionConnectionOverrides {
		if err := partitionConnectionOverridesBucket.AddPartitionConnectionOverride(
			partitionConnectionIdDbTypeFromPartitionConnectionId(partitionConnectionId),
			partitionConnectionDbTypeFromPartitionConnection(partitionConnection)); err != nil {
			return nil, stacktrace.Propagate(err, "Unable to initialize partition connection overrides")
		}
	}

	return &PartitionTopology{
		lock:                         &sync.RWMutex{},
		servicePartitions:            servicePartitionsBucket,
		partitionServices:            partitionServicesBucket,
		partitionConnectionOverrides: partitionConnectionOverridesBucket,
		defaultConnection:            defaultConnection,
	}, nil
}

// ParsePartitionId returns the partition ID form the provided strings.
// As partition ID is optional in most places, it falls back to DefaultPartitionID is the argument is nil or empty
func ParsePartitionId(partitionIdMaybe *string) service_network_types.PartitionID {
	if partitionIdMaybe == nil || *partitionIdMaybe == "" {
		return DefaultPartitionId
	}
	return service_network_types.PartitionID(*partitionIdMaybe)
}

// ================================================================================================
//
//	Public Methods
//
// ================================================================================================

// CreateEmptyPartitionWithDefaultConnection creates an empty connection with no connection overrides (i.e. all
// connections to this partition will inherit the defaultConnection)
// It returns an error if the partition ID already exists
func (topology *PartitionTopology) CreateEmptyPartitionWithDefaultConnection(newPartitionId service_network_types.PartitionID) error {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	exists, err := topology.partitionServices.DoesPartitionExist(partition.PartitionID(newPartitionId))
	if err != nil {
		return stacktrace.Propagate(err, "Attempted to check whether partition with ID '%v' exists but failed", newPartitionId)
	}
	if exists {
		return stacktrace.NewError("Partition with ID '%v' can't be created empty because it already exists in the topology", newPartitionId)
	}
	// servicePartitions remains unchanged as the new partition is empty
	// partitionConnections remains unchanged as default connection is being used for this new partition

	// update partitionServices. As the new partition is empty, it is mapped to an empty set
	err = topology.partitionServices.AddServicesToPartition(partition.PartitionID(newPartitionId), map[service.ServiceName]bool{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding empty partition '%v' to store", newPartitionId)
	}
	return nil
}

// RemovePartition removes the partition from the topology if it is present and empty.
// If it is not present, it returns successfully and does nothing
// If the partition is present and not empty, it throws an error as the partition cannot be removed from the topology
// Note that the default partition cannot be removed. It will throw an error is an attempt is being made to remove the
// default partition
func (topology *PartitionTopology) RemovePartition(partitionId service_network_types.PartitionID) error {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	if partitionId == DefaultPartitionId {
		return stacktrace.NewError("Default partition cannot be removed")
	}

	servicesInPartition, err := topology.partitionServices.GetServicesForPartition(partition.PartitionID(partitionId))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while fetching services for partition '%v'", partitionId)
	}

	numServicesInPartition := len(servicesInPartition)
	if numServicesInPartition > 0 {
		// partition is not empty. No-op
		return stacktrace.NewError("Partition '%s' cannot be removed as it currently contains '%d' services", partitionId, numServicesInPartition)
	}

	// delete the entry in partitionServices
	err = topology.partitionServices.DeletePartition(partition.PartitionID(partitionId))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while removing the partition '%v'", partitionId)
	}

	// update partition connections dropping all potential entries referencing the deleted partition
	allPartitionConnections, err := topology.partitionConnectionOverrides.GetAllPartitionConnectionOverrides()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting all partition connection overrides")
	}
	for partitionConnectionId := range allPartitionConnections {
		if partitionConnectionId.LexicalFirst == partition.PartitionID(partitionId) || partitionConnectionId.LexicalSecond == partition.PartitionID(partitionId) {
			if err = topology.partitionConnectionOverrides.RemovePartitionConnectionOverride(partitionConnectionId); err != nil {
				return stacktrace.Propagate(err, "An error occurred while removing partition connection with ID '%v'", partitionConnectionId)
			}
		}
	}
	return nil
}

// SetDefaultConnection sets the default connection by updating its value.
// Note that all connections between 2 partitions inheriting from defaultConnection will be affected
func (topology *PartitionTopology) SetDefaultConnection(connection PartitionConnection) {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	topology.defaultConnection = connection
}

// GetDefaultConnection returns a safe-copy of the current defaultConnection
// Use SetDefaultConnection to update the default connection of this topology
func (topology *PartitionTopology) GetDefaultConnection() PartitionConnection {
	topology.lock.RLock()
	defer topology.lock.RUnlock()
	return topology.defaultConnection
}

// SetConnection overrides the connection between partition1 and partition2.
// It throws an error if either of the two partitions does not exist yet
func (topology *PartitionTopology) SetConnection(partition1 service_network_types.PartitionID, partition2 service_network_types.PartitionID, connection PartitionConnection) error {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	exists, err := topology.partitionServices.DoesPartitionExist(partition.PartitionID(partition1))
	if err != nil {
		return stacktrace.Propagate(err, "Attempted to check whether partition with ID '%v' exists but failed", partition1)
	}
	if !exists {
		return stacktrace.NewError("About to set a connection between '%s' and '%s' but '%s' does not exist", partition1, partition2, partition1)
	}

	exists, err = topology.partitionServices.DoesPartitionExist(partition.PartitionID(partition2))
	if err != nil {
		return stacktrace.Propagate(err, "Attempted to check whether partition with ID '%v' exists but failed", partition2)
	}
	if !exists {
		return stacktrace.NewError("About to set a connection between '%s' and '%s' but '%s' does not exist", partition1, partition2, partition2)
	}

	partitionConnectionIdDbType := partitionConnectionIdDbTypeFromPartitionIds(partition1, partition2)
	partitionConnectionDbType := partitionConnectionDbTypeFromPartitionConnection(connection)
	if err = topology.partitionConnectionOverrides.AddPartitionConnectionOverride(partitionConnectionIdDbType, partitionConnectionDbType); err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding partition with id '%v' to bucket", partitionConnectionIdDbType)
	}
	return nil
}

// UnsetConnection unsets the connection override between partition1 and partition2. It will therefore fallback to
// defaultConnection
// It throws an error if either of the two partitions does not exist yet
// It no-ops if there was no override for this partition connection yet
func (topology *PartitionTopology) UnsetConnection(partition1 service_network_types.PartitionID, partition2 service_network_types.PartitionID) error {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	exists, err := topology.partitionServices.DoesPartitionExist(partition.PartitionID(partition1))
	if err != nil {
		return stacktrace.Propagate(err, "Attempted to check whether partition with ID '%v' exists but failed", partition1)
	}
	if !exists {
		return stacktrace.NewError("About to unset a connection between '%s' and '%s' but '%s' does not exist", partition1, partition2, partition1)
	}

	exists, err = topology.partitionServices.DoesPartitionExist(partition.PartitionID(partition2))
	if err != nil {
		return stacktrace.Propagate(err, "Attempted to check whether partition with ID '%v' exists but failed", partition2)
	}
	if !exists {
		return stacktrace.NewError("About to unset a connection between '%s' and '%s' but '%s' does not exist", partition1, partition2, partition2)
	}
	partitionConnectionIdDbType := partitionConnectionIdDbTypeFromPartitionIds(partition1, partition2)
	if err = topology.partitionConnectionOverrides.RemovePartitionConnectionOverride(partitionConnectionIdDbType); err != nil {
		return stacktrace.Propagate(err, "An error occurred while removing partition connection with id '%v'", partitionConnectionIdDbType)
	}

	return nil
}

func (topology *PartitionTopology) AddService(serviceName service.ServiceName, partitionId service_network_types.PartitionID) error {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	exists, err := topology.servicePartitions.DoesServiceExist(serviceName)
	if err != nil {
		return stacktrace.NewError("Cannot assign service to '%v' to partition '%v'; as we couldn't verify whether the service already exists in some partition", serviceName, partitionId)
	}
	if exists {
		existingPartition, err := topology.servicePartitions.GetPartitionForService(serviceName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while fetching partition for service '%v'", serviceName)
		}
		return stacktrace.NewError(
			"Cannot add service '%v' to partition '%v' because the service is already assigned to partition '%v'",
			serviceName,
			partitionId,
			existingPartition)
	}

	exists, err = topology.partitionServices.DoesPartitionExist(partition.PartitionID(partitionId))
	if err != nil {
		return stacktrace.NewError(
			"Cannot assign service '%v' to partition '%v'; the partition doesn't exist",
			serviceName,
			partitionId)
	}

	if !exists {
		return stacktrace.NewError(
			"Cannot assign service '%v' to partition '%v'; the partition doesn't exist",
			serviceName,
			partitionId)
	}
	err = topology.servicePartitions.AddPartitionToService(serviceName, partition.PartitionID(partitionId))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding partition '%v' to service '%v'", partitionId, serviceName)
	}
	err = topology.partitionServices.AddServiceToPartition(partition.PartitionID(partitionId), serviceName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding service '%v' to partition '%v'", serviceName, partitionId)
	}
	return nil
}

// RemoveService removes the given service from the topology, if it exists. If it doesn't exist, this is a no-op.
// Note that RemoveService leaves the partition in the topology even if it is empty after the service has been removed
func (topology *PartitionTopology) RemoveService(serviceName service.ServiceName) error {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	partitionId, err := topology.servicePartitions.GetPartitionForService(serviceName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while fetching the partition for service '%v'", serviceName)
	}
	if partitionId == partitionNotFoundForService {
		return nil
	}

	if err = topology.servicePartitions.RemoveService(serviceName); err != nil {
		return stacktrace.Propagate(err, "An error occurred while removing service '%v' from underlying service partition store", serviceName)
	}

	services, err := topology.partitionServices.GetServicesForPartition(partitionId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting services for partition '%v'", partitionId)
	}
	if len(services) == 0 {
		return nil
	}
	if err = topology.partitionServices.RemoveServiceFromPartition(serviceName, partitionId); err != nil {
		return stacktrace.Propagate(err, "An error occurred while removing service '%v' from partition '%v'", serviceName, partitionId)
	}
	return nil
}

func (topology *PartitionTopology) GetPartitionServices() (map[service_network_types.PartitionID]map[service.ServiceName]bool, error) {
	topology.lock.RLock()
	defer topology.lock.RUnlock()
	allPartitions, err := topology.partitionServices.GetAllPartitions()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while retrieving all partitions")
	}
	allPartitionsWithServiceNetworkPartitionIdType := map[service_network_types.PartitionID]map[service.ServiceName]bool{}
	for partitionId, services := range allPartitions {
		allPartitionsWithServiceNetworkPartitionIdType[service_network_types.PartitionID(partitionId)] = services
	}
	return allPartitionsWithServiceNetworkPartitionIdType, nil
}

// GetPartitionConnection returns a clone of the partition connection between the 2 partitions
// It also returns a boolean indicating whether the connection was the default connection or not
// It throws an error if the one of the partition does not exist.
func (topology *PartitionTopology) GetPartitionConnection(partition1 service_network_types.PartitionID, partition2 service_network_types.PartitionID) (bool, PartitionConnection, error) {
	topology.lock.RLock()
	defer topology.lock.RUnlock()
	exists, err := topology.partitionServices.DoesPartitionExist(partition.PartitionID(partition1))
	if err != nil {
		return false, ConnectionAllowed, stacktrace.Propagate(err, "Attempted to check whether partition with ID '%v' exists but failed", partition1)
	}
	if !exists {
		return false, ConnectionAllowed, stacktrace.NewError("About to get a connection between '%s' and '%s' but '%s' does not exist", partition1, partition2, partition1)
	}

	exists, err = topology.partitionServices.DoesPartitionExist(partition.PartitionID(partition2))
	if err != nil {
		return false, ConnectionAllowed, stacktrace.Propagate(err, "Attempted to check whether partition with ID '%v' exists but failed", partition2)
	}
	if !exists {
		return false, ConnectionAllowed, stacktrace.NewError("About to get a connection between '%s' and '%s' but '%s' does not exist", partition1, partition2, partition2)
	}

	partitionConnectionIdDbType := partitionConnectionIdDbTypeFromPartitionIds(partition1, partition2)
	exists, err = topology.partitionConnectionOverrides.DoesPartitionConnectionOverrideExist(partitionConnectionIdDbType)
	if err != nil {
		return false, ConnectionAllowed, stacktrace.Propagate(err, "An error occurred while verifying whether partition connection override with id '%v' exists", partitionConnectionIdDbType)
	}
	if !exists {
		return true, topology.GetDefaultConnection(), nil
	}

	currentPartitionConnectionDbType, err := topology.partitionConnectionOverrides.GetPartitionConnectionOverride(partitionConnectionIdDbType)
	if err != nil {
		return false, ConnectionAllowed, stacktrace.Propagate(err, "An error occurred while getting the partition connection with id '%v'", partitionConnectionIdDbType)
	}

	partitionConnection := newPartitionConnectionFromDbType(currentPartitionConnectionDbType)
	return false, partitionConnection, nil
}

func (topology *PartitionTopology) GetServicePartitions() (map[service.ServiceName]service_network_types.PartitionID, error) {
	topology.lock.RLock()
	defer topology.lock.RUnlock()
	allServicePartitions, err := topology.servicePartitions.GetAllServicePartitions()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching service partition mappings")
	}
	result := map[service.ServiceName]service_network_types.PartitionID{}
	for serviceName, partitionId := range allServicePartitions {
		result[serviceName] = service_network_types.PartitionID(partitionId)
	}
	return result, nil
}

// GetServicePartitionConnectionConfigByServiceName this method returns a partition config map
// containing information a structure similar to adjacency graph hashmap data structure between services
// where nodes are services, and edges are partition connection object
func (topology *PartitionTopology) GetServicePartitionConnectionConfigByServiceName() (map[service.ServiceName]map[service.ServiceName]*PartitionConnection, error) {
	topology.lock.RLock()
	defer topology.lock.RUnlock()
	allPartitions, err := topology.partitionServices.GetAllPartitions()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while reading all partitions")
	}
	result := map[service.ServiceName]map[service.ServiceName]*PartitionConnection{}
	for partitionId, servicesInPartition := range allPartitions {
		for serviceName := range servicesInPartition {
			partitionConnectionConfigBetweenServices := map[service.ServiceName]*PartitionConnection{}
			for otherPartitionId, servicesInOtherPartition := range allPartitions {
				if partitionId == otherPartitionId {
					// Two services in the same partition will never block each other
					continue
				}
				connection, err := topology.getPartitionConnectionUnlocked(service_network_types.PartitionID(partitionId), service_network_types.PartitionID(otherPartitionId))
				if err != nil {
					return nil, stacktrace.NewError("Couldn't get connection between partitions '%v' and '%v'", partitionId, otherPartitionId)
				}
				for otherServiceId := range servicesInOtherPartition {
					partitionConnectionConfigBetweenServices[otherServiceId] = &connection
				}
			}
			result[serviceName] = partitionConnectionConfigBetweenServices
		}
	}
	return result, nil
}

// ================================================================================================
//
//	Private Helper Methods
//
// ================================================================================================
func (topology *PartitionTopology) getPartitionConnectionUnlocked(
	a service_network_types.PartitionID,
	b service_network_types.PartitionID) (PartitionConnection, error) {

	exists, err := topology.partitionServices.DoesPartitionExist(partition.PartitionID(a))
	if err != nil {
		return ConnectionAllowed, stacktrace.Propagate(err, "Attempted to check whether partition with ID '%v' exists but failed", a)
	}
	if !exists {
		return ConnectionAllowed, stacktrace.NewError("Unrecognized partition '%v'", a)
	}

	exists, err = topology.partitionServices.DoesPartitionExist(partition.PartitionID(b))
	if err != nil {
		return ConnectionAllowed, stacktrace.Propagate(err, "Attempted to check whether partition with ID '%v' exists but failed", b)
	}
	if !exists {
		return ConnectionAllowed, stacktrace.NewError("Unrecognized partition '%v'", b)
	}

	partitionConnectionIdDbType := partitionConnectionIdDbTypeFromPartitionIds(a, b)

	exists, err = topology.partitionConnectionOverrides.DoesPartitionConnectionOverrideExist(partitionConnectionIdDbType)
	if err != nil {
		return ConnectionAllowed, stacktrace.Propagate(err, "An error occurred while verifying whether partition connection override with id '%v' exists", partitionConnectionIdDbType)
	}
	if !exists {
		return topology.GetDefaultConnection(), nil
	}

	currentPartitionConnectionDbType, err := topology.partitionConnectionOverrides.GetPartitionConnectionOverride(partitionConnectionIdDbType)
	if err != nil {
		return ConnectionAllowed, stacktrace.Propagate(err, "An error occurred while getting the partition connection with id '%v'", partitionConnectionIdDbType)
	}
	partitionConnection := newPartitionConnectionFromDbType(currentPartitionConnectionDbType)
	return partitionConnection, nil

}

func partitionConnectionDbTypeFromPartitionConnection(connection PartitionConnection) partition_connection_overrides.PartitionConnection {
	return partition_connection_overrides.PartitionConnection{
		PacketLoss: connection.packetLoss.packetLossPercentage,
		PacketDelayDistribution: partition_connection_overrides.DelayDistribution{
			AvgDelayMs:  connection.packetDelayDistribution.avgDelayMs,
			Jitter:      connection.packetDelayDistribution.jitter,
			Correlation: connection.packetDelayDistribution.correlation,
		},
	}
}

func partitionConnectionIdDbTypeFromPartitionConnectionId(connectionId service_network_types.PartitionConnectionID) partition_connection_overrides.PartitionConnectionID {
	return partition_connection_overrides.PartitionConnectionID{
		LexicalFirst:  partition.PartitionID(connectionId.GetFirst()),
		LexicalSecond: partition.PartitionID(connectionId.GetSecond()),
	}
}

func partitionConnectionIdDbTypeFromPartitionIds(partitionId1, partitionId2 service_network_types.PartitionID) partition_connection_overrides.PartitionConnectionID {
	return partitionConnectionIdDbTypeFromPartitionConnectionId(*service_network_types.NewPartitionConnectionID(partitionId1, partitionId2))
}
