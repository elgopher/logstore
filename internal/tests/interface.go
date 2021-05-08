// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package tests

type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
	Helper()
	Cleanup(func())
}
