// Code generated by MockGen. DO NOT EDIT.
// Source: invest_handler.go

// Package part_rebalance is a generated GoMock package.
package part_rebalance

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockEventChecker is a mock of EventChecker interface
type MockEventChecker struct {
	ctrl     *gomock.Controller
	recorder *MockEventCheckerMockRecorder
}

// MockEventCheckerMockRecorder is the mock recorder for MockEventChecker
type MockEventCheckerMockRecorder struct {
	mock *MockEventChecker
}

// NewMockEventChecker creates a new mock instance
func NewMockEventChecker(ctrl *gomock.Controller) *MockEventChecker {
	mock := &MockEventChecker{ctrl: ctrl}
	mock.recorder = &MockEventCheckerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEventChecker) EXPECT() *MockEventCheckerMockRecorder {
	return m.recorder
}

// checkEventHandled mocks base method
func (m *MockEventChecker) checkEventHandled(arg0 *checkEventParam) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "checkEventHandled", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// checkEventHandled indicates an expected call of checkEventHandled
func (mr *MockEventCheckerMockRecorder) checkEventHandled(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "checkEventHandled", reflect.TypeOf((*MockEventChecker)(nil).checkEventHandled), arg0)
}
