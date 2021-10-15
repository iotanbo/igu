// fstestutils provides utilities for testing file system items.
// Handy when testing copy, move etc.
package fstestutils

import (
	"testing"

	"github.com/iotanbo/igu/pkg/fu"
	"github.com/stretchr/testify/require"
	//"github.com/iotanbo/igu/pkg/ec"
	//. "github.com/iotanbo/igu/pkg/errs"
)

var expect = require.True

func TestCreateTestDirTree(t *testing.T) {
	rootDir := join("_testdata", "temp", "testdata")
	e := CreateSourceDirTree(rootDir)
	expect(t, e.None(), e)
	result := AssertAllSourceItemsConsistent(rootDir)
	expect(t, result, "TestDirTree not consistent")
}

// func TestValidateTestDirTreeStructure(t *testing.T) {
// 	rootDir := join("_testdata", "temp", "testdata")
// 	result := AssertTestDirTreeConsistent(rootDir)
// 	expect(t, result, "TestDirTree not valid")
// }

func TestCreateMergeDestination(t *testing.T) {
	rootDir := join("_testdata", "temp", "merge_dest")
	e := CreatePreExistingDestination(rootDir)
	expect(t, e.None(), e)
	intact := AssertAllPreExistingItemsConsistent(rootDir)
	expect(t, intact)
	// Modify file contents
	fu.CreateTextFile(join(rootDir, "test.txt"), "dummy_contents", true)
	intact = AssertAllPreExistingItemsConsistent(rootDir)
	expect(t, !intact)
	intact = AssertUniquePreExistingItemsConsistent(rootDir)
	expect(t, intact)
}

// func TestAssertMergeDestinationContentsIntact(t *testing.T) {
// 	rootDir := join("_testdata", "temp", "merge_dest")
// 	intact := AssertMergeDestinationContentsIntact(rootDir)
// 	expect(t, !intact)
// }
// func TestAssertMergeDestinationUniqueElementsIntact(t *testing.T) {
// 	rootDir := join("_testdata", "temp", "merge_dest")
// 	intact := AssertMergeDestinationUniqueElementsIntact(rootDir)
// 	expect(t, !intact)
// }
