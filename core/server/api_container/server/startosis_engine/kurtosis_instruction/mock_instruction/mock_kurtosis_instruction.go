// Code generated by mockery v2.20.2. DO NOT EDIT.

package mock_instruction

import (
	context "context"

	kurtosis_core_rpc_api_bindings "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	kurtosis_starlark_framework "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"

	mock "github.com/stretchr/testify/mock"

	startosis_validator "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
)

// MockKurtosisInstruction is an autogenerated mock type for the KurtosisInstruction type
type MockKurtosisInstruction struct {
	mock.Mock
}

type MockKurtosisInstruction_Expecter struct {
	mock *mock.Mock
}

func (_m *MockKurtosisInstruction) EXPECT() *MockKurtosisInstruction_Expecter {
	return &MockKurtosisInstruction_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: ctx
func (_m *MockKurtosisInstruction) Execute(ctx context.Context) (*string, error) {
	ret := _m.Called(ctx)

	var r0 *string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*string, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *string); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockKurtosisInstruction_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type MockKurtosisInstruction_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockKurtosisInstruction_Expecter) Execute(ctx interface{}) *MockKurtosisInstruction_Execute_Call {
	return &MockKurtosisInstruction_Execute_Call{Call: _e.mock.On("Execute", ctx)}
}

func (_c *MockKurtosisInstruction_Execute_Call) Run(run func(ctx context.Context)) *MockKurtosisInstruction_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockKurtosisInstruction_Execute_Call) Return(_a0 *string, _a1 error) *MockKurtosisInstruction_Execute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockKurtosisInstruction_Execute_Call) RunAndReturn(run func(context.Context) (*string, error)) *MockKurtosisInstruction_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// GetCanonicalInstruction provides a mock function with given fields:
func (_m *MockKurtosisInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	ret := _m.Called()

	var r0 *kurtosis_core_rpc_api_bindings.StarlarkInstruction
	if rf, ok := ret.Get(0).(func() *kurtosis_core_rpc_api_bindings.StarlarkInstruction); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*kurtosis_core_rpc_api_bindings.StarlarkInstruction)
		}
	}

	return r0
}

// MockKurtosisInstruction_GetCanonicalInstruction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCanonicalInstruction'
type MockKurtosisInstruction_GetCanonicalInstruction_Call struct {
	*mock.Call
}

// GetCanonicalInstruction is a helper method to define mock.On call
func (_e *MockKurtosisInstruction_Expecter) GetCanonicalInstruction() *MockKurtosisInstruction_GetCanonicalInstruction_Call {
	return &MockKurtosisInstruction_GetCanonicalInstruction_Call{Call: _e.mock.On("GetCanonicalInstruction")}
}

func (_c *MockKurtosisInstruction_GetCanonicalInstruction_Call) Run(run func()) *MockKurtosisInstruction_GetCanonicalInstruction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockKurtosisInstruction_GetCanonicalInstruction_Call) Return(_a0 *kurtosis_core_rpc_api_bindings.StarlarkInstruction) *MockKurtosisInstruction_GetCanonicalInstruction_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockKurtosisInstruction_GetCanonicalInstruction_Call) RunAndReturn(run func() *kurtosis_core_rpc_api_bindings.StarlarkInstruction) *MockKurtosisInstruction_GetCanonicalInstruction_Call {
	_c.Call.Return(run)
	return _c
}

// GetPositionInOriginalScript provides a mock function with given fields:
func (_m *MockKurtosisInstruction) GetPositionInOriginalScript() *kurtosis_starlark_framework.KurtosisBuiltinPosition {
	ret := _m.Called()

	var r0 *kurtosis_starlark_framework.KurtosisBuiltinPosition
	if rf, ok := ret.Get(0).(func() *kurtosis_starlark_framework.KurtosisBuiltinPosition); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*kurtosis_starlark_framework.KurtosisBuiltinPosition)
		}
	}

	return r0
}

// MockKurtosisInstruction_GetPositionInOriginalScript_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetPositionInOriginalScript'
type MockKurtosisInstruction_GetPositionInOriginalScript_Call struct {
	*mock.Call
}

// GetPositionInOriginalScript is a helper method to define mock.On call
func (_e *MockKurtosisInstruction_Expecter) GetPositionInOriginalScript() *MockKurtosisInstruction_GetPositionInOriginalScript_Call {
	return &MockKurtosisInstruction_GetPositionInOriginalScript_Call{Call: _e.mock.On("GetPositionInOriginalScript")}
}

func (_c *MockKurtosisInstruction_GetPositionInOriginalScript_Call) Run(run func()) *MockKurtosisInstruction_GetPositionInOriginalScript_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockKurtosisInstruction_GetPositionInOriginalScript_Call) Return(_a0 *kurtosis_starlark_framework.KurtosisBuiltinPosition) *MockKurtosisInstruction_GetPositionInOriginalScript_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockKurtosisInstruction_GetPositionInOriginalScript_Call) RunAndReturn(run func() *kurtosis_starlark_framework.KurtosisBuiltinPosition) *MockKurtosisInstruction_GetPositionInOriginalScript_Call {
	_c.Call.Return(run)
	return _c
}

// String provides a mock function with given fields:
func (_m *MockKurtosisInstruction) String() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockKurtosisInstruction_String_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'String'
type MockKurtosisInstruction_String_Call struct {
	*mock.Call
}

// String is a helper method to define mock.On call
func (_e *MockKurtosisInstruction_Expecter) String() *MockKurtosisInstruction_String_Call {
	return &MockKurtosisInstruction_String_Call{Call: _e.mock.On("String")}
}

func (_c *MockKurtosisInstruction_String_Call) Run(run func()) *MockKurtosisInstruction_String_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockKurtosisInstruction_String_Call) Return(_a0 string) *MockKurtosisInstruction_String_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockKurtosisInstruction_String_Call) RunAndReturn(run func() string) *MockKurtosisInstruction_String_Call {
	_c.Call.Return(run)
	return _c
}

// ValidateAndUpdateEnvironment provides a mock function with given fields: environment
func (_m *MockKurtosisInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	ret := _m.Called(environment)

	var r0 error
	if rf, ok := ret.Get(0).(func(*startosis_validator.ValidatorEnvironment) error); ok {
		r0 = rf(environment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockKurtosisInstruction_ValidateAndUpdateEnvironment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ValidateAndUpdateEnvironment'
type MockKurtosisInstruction_ValidateAndUpdateEnvironment_Call struct {
	*mock.Call
}

// ValidateAndUpdateEnvironment is a helper method to define mock.On call
//   - environment *startosis_validator.ValidatorEnvironment
func (_e *MockKurtosisInstruction_Expecter) ValidateAndUpdateEnvironment(environment interface{}) *MockKurtosisInstruction_ValidateAndUpdateEnvironment_Call {
	return &MockKurtosisInstruction_ValidateAndUpdateEnvironment_Call{Call: _e.mock.On("ValidateAndUpdateEnvironment", environment)}
}

func (_c *MockKurtosisInstruction_ValidateAndUpdateEnvironment_Call) Run(run func(environment *startosis_validator.ValidatorEnvironment)) *MockKurtosisInstruction_ValidateAndUpdateEnvironment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*startosis_validator.ValidatorEnvironment))
	})
	return _c
}

func (_c *MockKurtosisInstruction_ValidateAndUpdateEnvironment_Call) Return(_a0 error) *MockKurtosisInstruction_ValidateAndUpdateEnvironment_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockKurtosisInstruction_ValidateAndUpdateEnvironment_Call) RunAndReturn(run func(*startosis_validator.ValidatorEnvironment) error) *MockKurtosisInstruction_ValidateAndUpdateEnvironment_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewMockKurtosisInstruction interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockKurtosisInstruction creates a new instance of MockKurtosisInstruction. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockKurtosisInstruction(t mockConstructorTestingTNewMockKurtosisInstruction) *MockKurtosisInstruction {
	mock := &MockKurtosisInstruction{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
