//go:build !minikube
// +build !minikube

// We don't run this test in Kubernetes because the gateway does not support start and stop services

package startosis_stop_service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	testName              = "startosis_stop_service_test"
	isPartitioningEnabled = false

	serviceName = "example-datastore-server-1"
	portId      = "grpc"

	starlarkScript = `
DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_NAME = "` + serviceName + `"
DATASTORE_PORT_ID = "` + portId + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"

def run(plan):
	plan.print("Adding service " + DATASTORE_SERVICE_NAME + ".")
	
	config = ServiceConfig(
		image = DATASTORE_IMAGE,
		ports = {
			DATASTORE_PORT_ID: PortSpec(number = DATASTORE_PORT_NUMBER, transport_protocol = DATASTORE_PORT_PROTOCOL)
		}
	)
	
	plan.add_service(name = DATASTORE_SERVICE_NAME, config = config)
	plan.print("Service " + DATASTORE_SERVICE_NAME + " deployed successfully.")
`
	// We stop the service we created through the script above with a different script
	stopScript = `
DATASTORE_SERVICE_NAME = "` + serviceName + `"
def run(plan):
	plan.stop_service(DATASTORE_SERVICE_NAME)
`
)

func TestStartosis(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		destroyErr := destroyEnclaveFunc()
		if destroyErr != nil {
			logrus.Errorf("Error destroying enclave at the end of integration test '%s'",
				testName)
		}
	}()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Starlark script to first add the datastore service...")
	logrus.Debugf("Starlark script content: \n%v", starlarkScript)

	runResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, starlarkScript)
	require.NoError(t, err, "Unexpected error executing Starlark script")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")

	expectedScriptOutput := `Adding service example-datastore-server-1.
Service 'example-datastore-server-1' added with service UUID '[a-z-0-9]+'
Service example-datastore-server-1 deployed successfully.
`
	require.Regexp(t, expectedScriptOutput, string(runResult.RunOutput))
	logrus.Infof("Successfully ran Starlark script to add datastore service")

	// Check that the service added by the script is functional
	logrus.Infof("Checking that services are all healthy")
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceName, portId),
		"Error validating datastore server '%s' is healthy",
		serviceName,
	)

	logrus.Infof("Validated that all services are healthy")

	// we run the stop script and validate that the service is unreachable.
	runResult, err = test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, stopScript)
	require.NoError(t, err, "Unexpected error executing stop script")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")

	expectedScriptOutput = `Service 'example-datastore-server-1' stopped
`
	require.Regexp(t, expectedScriptOutput, string(runResult.RunOutput))

	require.Error(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceName, portId),
		"Error validating datastore server '%s' is not healthy",
		serviceName,
	)

	logrus.Infof("Validated that the service is stopped")

	// we run the stop script one more time and validate that an error is returned since the service is already stopped.
	runResult, _ = test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, stopScript)
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.NotEmpty(t, runResult.ExecutionError, "Expected execution error coming from already stopped service")
	expectedErrorStr := fmt.Sprintf("Service '%v' is already stopped", serviceName)
	require.Contains(t, runResult.ExecutionError.ErrorMessage, expectedErrorStr)
}
