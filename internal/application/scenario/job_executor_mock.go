// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package scenario

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	"sync"
)

// Ensure, that ScenarioJobExecutorMock does implement ScenarioJobExecutor.
// If this is not the case, regenerate this file with moq.
var _ ScenarioJobExecutor = &ScenarioJobExecutorMock{}

// ScenarioJobExecutorMock is a mock implementation of ScenarioJobExecutor.
//
//	func TestSomethingThatUsesScenarioJobExecutor(t *testing.T) {
//
//		// make and configure a mocked ScenarioJobExecutor
//		mockedScenarioJobExecutor := &ScenarioJobExecutorMock{
//			ExecuteFunc: func(ctx context.Context, namespacedName types.NamespacedName) error {
//				panic("mock out the Execute method")
//			},
//		}
//
//		// use mockedScenarioJobExecutor in code that requires ScenarioJobExecutor
//		// and then make assertions.
//
//	}
type ScenarioJobExecutorMock struct {
	// ExecuteFunc mocks the Execute method.
	ExecuteFunc func(ctx context.Context, namespacedName types.NamespacedName) error

	// calls tracks calls to the methods.
	calls struct {
		// Execute holds details about calls to the Execute method.
		Execute []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// NamespacedName is the namespacedName argument value.
			NamespacedName types.NamespacedName
		}
	}
	lockExecute sync.RWMutex
}

// Execute calls ExecuteFunc.
func (mock *ScenarioJobExecutorMock) Execute(ctx context.Context, namespacedName types.NamespacedName) error {
	if mock.ExecuteFunc == nil {
		panic("ScenarioJobExecutorMock.ExecuteFunc: method is nil but ScenarioJobExecutor.Execute was just called")
	}
	callInfo := struct {
		Ctx            context.Context
		NamespacedName types.NamespacedName
	}{
		Ctx:            ctx,
		NamespacedName: namespacedName,
	}
	mock.lockExecute.Lock()
	mock.calls.Execute = append(mock.calls.Execute, callInfo)
	mock.lockExecute.Unlock()
	return mock.ExecuteFunc(ctx, namespacedName)
}

// ExecuteCalls gets all the calls that were made to Execute.
// Check the length with:
//
//	len(mockedScenarioJobExecutor.ExecuteCalls())
func (mock *ScenarioJobExecutorMock) ExecuteCalls() []struct {
	Ctx            context.Context
	NamespacedName types.NamespacedName
} {
	var calls []struct {
		Ctx            context.Context
		NamespacedName types.NamespacedName
	}
	mock.lockExecute.RLock()
	calls = mock.calls.Execute
	mock.lockExecute.RUnlock()
	return calls
}