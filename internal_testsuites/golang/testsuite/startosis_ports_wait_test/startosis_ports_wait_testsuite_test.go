package startosis_ports_wait_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

const (
	name                  = "startosis-ports-wait"
	isPartitioningEnabled = false
)

type StartosisPortsWaitTestSuite struct {
	suite.Suite
	enclaveCtx         *enclaves.EnclaveContext
	destroyEnclaveFunc func() error
}

func TestStartosisPortsWaitTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisPortsWaitTestSuite))
}

func (suite *StartosisPortsWaitTestSuite) SetupSuite() {
	ctx := context.Background()
	t := suite.T()
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, name, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	suite.enclaveCtx = enclaveCtx
	suite.destroyEnclaveFunc = destroyEnclaveFunc
}

func (suite *StartosisPortsWaitTestSuite) TearDownSuite() {
	err := suite.destroyEnclaveFunc()
	require.NoError(suite.T(), err, "Destroying the test suite's enclave process has failed, you will have to remove it manually")
}

func (suite *StartosisPortsWaitTestSuite) RunScript(ctx context.Context, script string) (*enclaves.StarlarkRunResult, error) {
	logrus.Infof("Executing Startosis script...")
	logrus.Debugf("Startosis script content: \n%v", script)

	return test_helpers.RunScriptWithDefaultConfig(ctx, suite.enclaveCtx, script)
}
