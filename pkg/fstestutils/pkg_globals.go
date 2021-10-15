package fstestutils

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/iotanbo/igu/pkg/fu"
)

//const testDirName string = "_testdir"

var join = filepath.Join
var printf = fmt.Printf
var errorf = fmt.Errorf

// Suffixes to be added to file contents when creating merge destination,
// allow to verify whether the file was overwritten.
var SourceItemSetBinSuffix = []byte{}
var SourceItemSetTextSuffix = ""
var PreExistingDestBinSuffix = []byte{0x75, 0x76, 0x77, 0x78}
var PreExistingDestTextSuffix = "-merge"

type ItemDescriptor struct {
	// path relative to the root
	Path string
	Type fu.FsItemType
	// Only for files
	IsBinary    bool
	Text        string
	BinContents []byte
	// Only for links
	LinkTarget string
}

// Only path equality matters because it serves as an ID.
func (a ItemDescriptor) Equal(b ItemDescriptor) bool {
	return strings.Compare(a.Path, b.Path) == 0
}

// True if a is present in slice b.
func (a ItemDescriptor) In(b []ItemDescriptor) bool {
	for _, el := range b {
		if a.Equal(el) {
			return true
		}
	}
	return false
}

// Returns intersection (common elements) of a and b.
func ItemDescriptorIntersection(a, b []ItemDescriptor) (result []ItemDescriptor) {
	for _, aDesc := range a {
		for _, bDesc := range b {
			if aDesc.Equal(bDesc) {
				result = append(result, aDesc)
			}
		}
	}
	return result
}

// Returns slice of elements of that are unique to a.
func ItemDescriptorUniqueElements(a, b []ItemDescriptor) (result []ItemDescriptor) {
	intersection := ItemDescriptorIntersection(a, b)
	for _, element := range a {
		if !element.In(intersection) {
			result = append(result, element)
		}
	}
	return result
}

/*
SOURCE DIRECTORY TREE STRUCTURE

	dir_a
		bin
			a.bin
		a.txt
		symlink_to_b.txt
	dir_b
		bin
			b.bin
		b.txt
	.hidden_dir
		.hidden_file.txt
	test.txt
*/
var SourceItemSet = []ItemDescriptor{
	{
		Path: "dir_a",
		Type: fu.TYPE_DIR,
	},
	{
		Path: join("dir_a", "bin"),
		Type: fu.TYPE_DIR,
	},
	{
		Path:        join("dir_a", "bin", "a.bin"),
		Type:        fu.TYPE_FILE,
		IsBinary:    true,
		BinContents: []byte{0x00, 0x01, 0x02, 0x03, '\n'},
	},
	{
		Path: join("dir_a", "a.txt"),
		Type: fu.TYPE_FILE,
		Text: "a.txt",
	},
	{
		Path:       join("dir_a", "symlink_to_b.txt"),
		Type:       fu.TYPE_SYMLINK,
		LinkTarget: "../dir_b/b.txt",
	},
	{
		Path: "dir_b",
		Type: fu.TYPE_DIR,
	},
	{
		Path: join("dir_b", "b.txt"),
		Type: fu.TYPE_FILE,
		Text: "b.txt",
	},
	{
		Path: join("dir_b", "bin"),
		Type: fu.TYPE_DIR,
	},
	{
		Path:        join("dir_b", "bin", "b.bin"),
		Type:        fu.TYPE_FILE,
		IsBinary:    true,
		BinContents: []byte{0x04, 0x05, 0x06, 0x07, '\n'},
	},
	// .hidden_dir
	{
		Path: ".hidden_dir",
		Type: fu.TYPE_DIR,
	},
	{
		Path: join(".hidden_dir", ".hidden_file.txt"),
		Type: fu.TYPE_FILE,
		Text: ".hidden_file.txt",
	},
	{
		Path: "test.txt",
		Type: fu.TYPE_FILE,
		Text: "test.txt",
	},
}

/*
DESTINATION DIRECTORY TREE STRUCTURE

	dir_a
		a.txt
	dir_c
		c.txt
	.hidden_dir
		.hidden_file.txt
	test.txt
	other.txt
*/
var DestPreExistingItemSet = []ItemDescriptor{
	{
		Path: "dir_a",
		Type: fu.TYPE_DIR,
	},
	{
		Path: join("dir_a", "a.txt"),
		Type: fu.TYPE_FILE,
		Text: "a.txt",
	},
	{
		Path: "dir_c",
		Type: fu.TYPE_DIR,
	},
	{
		Path: join("dir_c", "c.txt"),
		Type: fu.TYPE_FILE,
		Text: "c.txt",
	},
	// .hidden_dir
	{
		Path: ".hidden_dir",
		Type: fu.TYPE_DIR,
	},
	{
		Path: join(".hidden_dir", ".hidden_file.txt"),
		Type: fu.TYPE_FILE,
		Text: ".hidden_file.txt",
	},
	{
		Path: "test.txt",
		Type: fu.TYPE_FILE,
		Text: "test.txt",
	},
	{
		Path: "other.txt",
		Type: fu.TYPE_FILE,
		Text: "other.txt",
	},
}
