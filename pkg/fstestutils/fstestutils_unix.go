// fstestutils provides utilities for testing file system items.
// Handy when testing copy, move etc.
package fstestutils

import (
	//"github.com/iotanbo/igu/pkg/fu"
	//"github.com/iotanbo/igu/pkg/ec"

	"bytes"
	"os"
	"strings"

	//lint:ignore ST1001 - for concise error handling.

	. "github.com/iotanbo/igu/pkg/errs"
	"github.com/iotanbo/igu/pkg/fu"
)

func createDirTree(rootDir string,
	itemList []ItemDescriptor,
	textFileSuffix string,
	binFileSuffix []byte) Err {
	for _, desc := range itemList {
		//printf("* %v\n", desc)
		p := join(rootDir, desc.Path)
		switch desc.Type {
		case fu.TYPE_DIR:
			if err := os.MkdirAll(p, 0755); err != nil {
				panic(errorf("can't create directory " + p))
			}
		case fu.TYPE_FILE:
			if desc.IsBinary {
				contents := desc.BinContents
				if len(binFileSuffix) > 0 {
					contents = append(contents, binFileSuffix...)
				}
				e := fu.CreateBinFile(p, contents, true)
				if e.Some() {
					return e
				}
			} else {
				contents := desc.Text
				if len(textFileSuffix) > 0 {
					contents += textFileSuffix
				}
				e := fu.CreateTextFile(p, contents, true)
				if e.Some() {
					return e
				}
			}
		case fu.TYPE_SYMLINK:
			if err := os.Symlink(desc.LinkTarget, p); err != nil {
				panic(errorf("can't create symlink '%s' to '%s'",
					desc.Path, desc.LinkTarget))
			}
		default:
			panic(errorf("* FsItemType not supported (yet): %v", desc.Type))
		}
	}

	return NoError
}

// Creates directory that contains files and other directories
// with contents that can be verified.
// Note: if rootDir already exists,
// it will be removed and then re-created.
func CreateSourceDirTree(rootDir string) Err {
	// rootDir := filepath.Join(parentDir, testDirName)
	exists, _, _ := fu.PathExists(rootDir)
	if exists {
		os.RemoveAll(rootDir)
	}
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return FromError(err)
	}
	return createDirTree(
		rootDir,
		SourceItemSet,
		SourceItemSetTextSuffix,
		SourceItemSetBinSuffix)
}

// Returns true if rootDir contains all items from itemSet
// and their contents is intact.
func AssertItemsConsistent(rootDir string,
	itemSet []ItemDescriptor,
	mergeTextFileSuffix string,
	mergeBinFileSuffix []byte,
) bool {
	for _, desc := range itemSet {
		path := join(rootDir, desc.Path)
		switch desc.Type {
		case fu.TYPE_DIR:
			// Check if directory exists
			exists, e := fu.DirExists(path)
			if e.Some() {
				panic(errorf("* can't check if dir exists '%s', %v\n", path, e))
			}
			if !exists {
				printf("* directory '%s' expected to exist but it does not.\n",
					path)
				return false
			}
		case fu.TYPE_FILE:
			if desc.IsBinary {
				expectedContents := append(desc.BinContents, mergeBinFileSuffix...)
				actualContents, e := fu.ReadBinFile(path)
				if e.Some() {
					panic(errorf("* can't read bin file '%s', %v\n", path, e))
				}
				if !bytes.Equal(expectedContents, actualContents) {
					printf("* bin file '%s' contents error, expected '%v', got '%v'\n",
						path, expectedContents, actualContents)
					return false
				}
			} else { // text file
				expectedContents := desc.Text + mergeTextFileSuffix
				actualContents, e := fu.ReadTextFile(path)
				if e.Some() {
					panic(errorf("* can't read text file '%s', %v\n", path, e))
				}
				if strings.Compare(expectedContents, actualContents) != 0 {
					printf("* text file '%s' contents error, expected '%v', got '%v'\n",
						path, expectedContents, actualContents)
					return false
				}
			}
		default:
			continue
		}
	}
	return true
}

// Checks if all items listed in SourceItemSet are consistent inside rootDir.
// Presence of other files inside rootDir does not affect the result.
func AssertAllSourceItemsConsistent(rootDir string) bool {
	return AssertItemsConsistent(
		rootDir,
		SourceItemSet,
		SourceItemSetTextSuffix,
		SourceItemSetBinSuffix,
	)
}

// Creates directory that contains some files
// from the test dir tree, and some extra files.
// This allows to check if existing files are overwritten
// and if extra files left intact.
func CreatePreExistingDestination(mergeDestRoot string) Err {
	exists, _, _ := fu.PathExists(mergeDestRoot)
	if exists {
		os.RemoveAll(mergeDestRoot)
	}
	if err := os.MkdirAll(mergeDestRoot, 0755); err != nil {
		return FromError(err)
	}
	return createDirTree(
		mergeDestRoot,
		DestPreExistingItemSet,
		PreExistingDestTextSuffix,
		PreExistingDestBinSuffix)
}

// Returns true if all files and directories originally present
// in the merge destination have their contents intact.
// This is useful to check overwrite modes.
func AssertAllPreExistingItemsConsistent(mergeDestRoot string) bool {
	return AssertItemsConsistent(
		mergeDestRoot,
		DestPreExistingItemSet,
		PreExistingDestTextSuffix,
		PreExistingDestBinSuffix,
	)
}

// Returns true if all files and directories that are
// unique to merge destination are intact.
// This is useful to check overwrite modes.
func AssertUniquePreExistingItemsConsistent(mergeDestRoot string) bool {
	uniqueToDest := ItemDescriptorUniqueElements(DestPreExistingItemSet, SourceItemSet)
	printf("* uniqueElements: %v\n", uniqueToDest)
	return AssertItemsConsistent(
		mergeDestRoot,
		uniqueToDest,
		PreExistingDestTextSuffix,
		PreExistingDestBinSuffix,
	)
}

// Returns true if all files and directories that are
// common to source and destination are overwritten.
// This is useful to check overwrite modes.
func AssertIntersectionOverwrittenInDest(mergeDestRoot string) bool {
	intersection := ItemDescriptorIntersection(DestPreExistingItemSet, SourceItemSet)
	printf("* intersection: %v\n", intersection)
	return AssertItemsConsistent(
		mergeDestRoot,
		intersection,
		SourceItemSetTextSuffix,
		SourceItemSetBinSuffix,
	)
}

// Returns true if all files and directories that are
// unique to source are consistent.
func AssertUniqueToSourceItemsConsistent(mergeDestRoot string) bool {
	uniqueToSrc := ItemDescriptorUniqueElements(SourceItemSet, DestPreExistingItemSet)
	printf("* uniqueToSrc: %v\n", uniqueToSrc)
	return AssertItemsConsistent(
		mergeDestRoot,
		uniqueToSrc,
		SourceItemSetTextSuffix,
		SourceItemSetBinSuffix,
	)
}
