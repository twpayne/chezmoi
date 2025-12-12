package chezmoiassert_test

import (
	"errors"
	"testing"

	"chezmoi.io/chezmoi/internal/chezmoiassert"
)

// mockTB is a mock testing.TB that tracks whether Fatal or Fatalf was called
// to prevent test failures.
type mockTB struct {
	testing.TB
	failed bool
}

func (m *mockTB) Helper() {}

func (m *mockTB) Fatal(args ...any) {
	m.failed = true
}

func (m *mockTB) Fatalf(format string, args ...any) {
	m.failed = true
}

func (m *mockTB) Errorf(format string, args ...any) {
	m.failed = true
}

func (m *mockTB) Error(args ...any) {
	m.failed = true
}

func (m *mockTB) FailNow() {
	m.failed = true
}

func (m *mockTB) Fail() {
	m.failed = true
}


func TestPanicsWithError(t *testing.T) {
	tests := []struct {
		name        string
		expectedErr error
		fn          func()
		shouldFail  bool
	}{
		{
			name:        "panic matches expected error",
			expectedErr: errors.New("boom"),
			fn: func() {
				panic(errors.New("boom"))
			},
			shouldFail: false,
		},
		{
			name:        "no panic should fail",
			expectedErr: errors.New("boom"),
			fn:          func() {},
			shouldFail:  true,
		},
		{
			name:        "panic with non-error should fail",
			expectedErr: errors.New("boom"),
			fn: func() {
				panic("not an error")
			},
			shouldFail: true,
		},
		{
			name:        "panic with different error should fail",
			expectedErr: errors.New("boom"),
			fn: func() {
				panic(errors.New("different error"))
			},
			shouldFail: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockTB{TB: t}
			chezmoiassert.PanicsWithError(mock, tc.expectedErr, tc.fn)

			if mock.failed && !tc.shouldFail {
				t.Errorf("assertion failed but should have passed")
			}
			if !mock.failed && tc.shouldFail {
				t.Errorf("assertion passed but should have failed")
			}
		})
	}
}

func TestPanicsWithErrorString(t *testing.T) {
	tests := []struct {
		name        string
		expectedMsg string
		fn          func()
		shouldFail  bool
	}{
		{
			name:        "panic matches expected string",
			expectedMsg: "fatal issue",
			fn: func() {
				panic(errors.New("fatal issue"))
			},
			shouldFail: false,
		},
		{
			name:        "no panic should fail",
			expectedMsg: "boom",
			fn:          func() {},
			shouldFail:  true,
		},
		{
			name:        "panic non-error should fail",
			expectedMsg: "boom",
			fn: func() {
				panic("text panic")
			},
			shouldFail: true,
		},
		{
			name:        "panic with different error message should fail",
			expectedMsg: "boom",
			fn: func() {
				panic(errors.New("different message"))
			},
			shouldFail: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockTB{TB: t}
			chezmoiassert.PanicsWithErrorString(mock, tc.expectedMsg, tc.fn)

			if mock.failed && !tc.shouldFail {
				t.Errorf("assertion failed but should have passed")
			}
			if !mock.failed && tc.shouldFail {
				t.Errorf("assertion passed but should have failed")
			}
		})
	}
}
