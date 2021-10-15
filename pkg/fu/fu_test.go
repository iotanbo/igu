package fu_test

import (
	//"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	//"github.com/iotanbo/igu/pkg/fstestutils"
	"github.com/iotanbo/igu/pkg/ec"
	"github.com/iotanbo/igu/pkg/fstestutils"
	"github.com/iotanbo/igu/pkg/fu"
	"github.com/stretchr/testify/require"
	//. "github.com/iotanbo/igu/pkg/errs"
)

var globalTmpDir string

// Setting this flag to false allows to preserve
// globalTmpDir after tests are complete for manual analysis.
var _removeTmpDirAfterTests = false

// Type aliases to improve readability
var join = filepath.Join
var mkdir = os.Mkdir
var printf = fmt.Printf
var expect = require.True

// var expectNot = require.False

// Initialized in TestMain
var testDirTreeRoot string
var existingFile string
var existingDir string
var nonExistingPath string

// Platform-specific variables
var symlinkToFile string

//var existingHardlink string
//var existingNamedPipe string

// anyContents is used while verifying file contents
// to specify that any file contents will do.
//var anyContents = "**any**"

// containsBinData is used while verifying file contents
// to specify that binary data must be verified.
//var containsBinData = "**bindata**"

// createTestDir creates a directory with specified name
// inside globalTmpDir and returns path to it.
// Default permissions are 0711.
func createTestDir(name string, perms ...os.FileMode) string {
	var resolvedPerms os.FileMode = 0711
	if len(perms) > 0 {
		resolvedPerms = perms[0]
	}
	newDir := join(globalTmpDir, name)
	err := mkdir(newDir, resolvedPerms)
	if err != nil {
		panic(
			fmt.Sprintf("createTestDir(): failed to create temp dir, '%v'\n", err))
	}
	return newDir
}

func TestMain(m *testing.M) {
	fmt.Println("--- Preparing to run FU package tests.")
	// Create a temporary dir for all tests and copy _testdata into that directory
	tmpDir, err := os.MkdirTemp("", "IGU_FU_tests")
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create temporary directory for FU tests: %v\n", err)
		os.Exit(1)
	}
	globalTmpDir = tmpDir
	printf("--- created temporary dir for FU tests: '%s'\n", tmpDir)

	// Create test data
	testDirTreeRoot = join(tmpDir, "temp", "_testdata")
	e := fstestutils.CreateSourceDirTree(testDirTreeRoot)
	if e.Some() {
		panic(fmt.Sprintf("TestMain(): failed to create test data: %v", e))
	}
	// Assign global variables
	existingFile = join(testDirTreeRoot, "test.txt")
	existingDir = testDirTreeRoot
	nonExistingPath = join(testDirTreeRoot, "not_exists")

	// Create symlinks
	symlinkToFile = join(testDirTreeRoot, "dir_a", "symlink_to_b.txt")

	// if runtime.GOOS != "windows" {
	// 	symlinksToBeCreated := []struct {
	// 		Src  string
	// 		Dest string
	// 	}{
	// 		{
	// 			Src:  existingFile,
	// 			Dest: symlinkToFile,
	// 		},
	// 		{
	// 			Src:  join(testDirTreeRoot, "dir_a"),
	// 			Dest: join(testDirTreeRoot, "dir_b", "symlink_to_a"),
	// 		},
	// 		{
	// 			Src:  join(testDirTreeRoot, "dir_b"),
	// 			Dest: join(testDirTreeRoot, "dir_a", "symlink_to_b"),
	// 		},
	// 	}

	// 	for _, s := range symlinksToBeCreated {
	// 		err := os.Symlink(s.Src, s.Dest)
	// 		if err != nil {
	// 			panic(fmt.Sprintf("TestMain(): failed to create symlink '%s': %v",
	// 				symlinkToFile, err))
	// 		}
	// 	}
	// }

	exitVal := m.Run()

	if _removeTmpDirAfterTests {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"can't remove temporary directory for FU tests: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("--- temporary directory removed, FU package tests complete.")
	} else {
		fmt.Printf("--- FU package tests complete. Please manually remove directory '%s'\n", globalTmpDir)
	}
	os.Exit(exitVal)
}

func TestGetItemType(t *testing.T) {
	//printf("* TestGetItemType(): using temp dir '%s'\n", globalTmpDir)
	// Create dedicated directory for this test
	tmpDir := createTestDir("test_get_item_type")
	printf("* TestGetItemType(): using temp dir '%s'\n", tmpDir)

	// When passing empty string, should return (TYPE_UNKNOWN, ec.NotFound).
	r, e := fu.GetItemType("")
	expect(t, e.Eq(ec.NotFound),
		`GetItemType(''): expected ec.NotFound, got '%v'`, e)
	expect(t, r == fu.TYPE_UNKNOWN,
		`GetItemType(''): expected TYPE_UNKNOWN, got '%v'`, r)

	// When passing existing file, should return (TYPE_FILE, NoError).
	r, e = fu.GetItemType(existingFile)
	expect(t, e.None(),
		`GetItemType(existingFile): returned error '%v'`, e)
	expect(t, r == fu.TYPE_FILE,
		`GetItemType(existingFile): expected TYPE_FILE, got '%v'`, r)

	// When passing non-existing file, should return (TYPE_UNKNOWN, ec.NotFound).
	r, e = fu.GetItemType(nonExistingPath)
	expect(t, e.Eq(ec.NotFound),
		`GetItemType(nonExistingPath): expected ec.NotFound, got '%v'`, e)
	expect(t, r == fu.TYPE_UNKNOWN,
		`GetItemType(nonExistingPath): expected TYPE_UNKNOWN, got '%v'`, r)

	// When passing directory, should return (TYPE_DIR, NoError).
	r, e = fu.GetItemType(testDirTreeRoot)
	expect(t, e.None(),
		`GetItemType(existingDir): expected ec.NoError, got '%v'`, e)
	expect(t, r == fu.TYPE_DIR,
		`GetItemType(existingDir): expected TYPE_DIR, got '%v'`, r)

	// UNIX-ONLY
	// When passing symlink, should return (TYPE_SYMLINK, NoError).
	r, e = fu.GetItemType(symlinkToFile)
	expect(t, e.None(),
		`GetItemType(symlinkToFile): expected ec.NoError, got '%v'`, e)
	expect(t, r == fu.TYPE_SYMLINK,
		`GetItemType(symlinkToFile): expected TYPE_SYMLINK, got '%v'`, r)

	// TODO: test TYPE_NAMED_PIPE
}

func TestFileExists(t *testing.T) {
	printf("* TestFileExists(): using temp dir '%s'\n", globalTmpDir)
	// When passing existing file, should return (true, no error)
	r, e := fu.FileExists(existingFile)
	expect(t, e.None(),
		`fu.FileExists(existingFile): returned error '%v'`, e)
	expect(t, r, "fu.FileExists(existingFile): returned false")

	// When passing non-existing file, should return (false, no error)
	r, e = fu.FileExists(nonExistingPath)
	expect(t, e.None(),
		`fu.FileExists(nonExistingPath): returned error '%v'`, e)
	expect(t, !r, `fu.FileExists(nonExistingPath): returned true`)

	// When passing empty string, should return (false, no error)
	r, e = fu.FileExists("")
	expect(t, e.None(),
		`fu.FileExists(''): returned unexpected error '%v'`, e)
	expect(t, !r, `fu.FileExists(''): returned true`)

	// When passing existing directory, should return (false, ec.Type)
	r, e = fu.FileExists(testDirTreeRoot)
	expect(t, e.Eq(ec.Type),
		`fu.FileExists(existingDir): didn't return ec.Type error, got '%v'`, e)
	expect(t, !r, "fu.FileExists(existingDir): returned true")

	// UNIX-ONLY
	// When passing symlink, should return (true, no error)
	r, e = fu.FileExists(symlinkToFile)
	expect(t, e.None(),
		`fu.FileExists(symlinkToFile): returned error '%v'`, e)
	expect(t, r, "fu.FileExists(symlinkToFile): returned false")

}

func TestCreateTextFile(t *testing.T) {
	// Create dedicated directory for this test
	tmpDir := createTestDir("test_create_text_file")
	printf("* TestCreateTextFile(): using temp dir '%s'\n", tmpDir)

	// Normally should create a file with given contents and return NoError
	path := join(tmpDir, "text_file.txt")
	e := fu.CreateTextFile(path, "test", false)
	expect(t, e.None(),
		`fu.CreateTextFile("text_file.txt", ..., false): `+
			`NOT returned ec.NoError, got '%v'`, e)
	exists, e := fu.FileExists(path)
	expect(t, e.None())
	expect(t, exists) // TODO: verify file contents

	// When insufficient permissions should return ec.PermissionDenied
	e = fu.CreateTextFile("/dummy.txt", "test", false)
	expect(t, e.Eq(ec.PermissionDenied),
		`fu.CreateTextFile("/dummy.txt", ..., false): `+
			`NOT returned ec.PermissionDenied, got '%v'`, e)

	// When file already exists should return ec.AlreadyExists
	e = fu.CreateTextFile(path, "test", false)
	expect(t, e.Eq(ec.AlreadyExists),
		`fu.CreateTextFile(path, ..., false): `+
			`NOT returned ec.AlreadyExists, got '%v'`, e)

	// When file already exists and overwrite=true should return ec.NoError
	e = fu.CreateTextFile(path, "test2", true)
	expect(t, e.None(),
		`fu.CreateTextFile(path, ..., true): `+
			`NOT returned ec.NoError, got '%v'`, e)

	// When destination already exists but is a directory
	// and overwrite=true, should return ec.Type
	e = fu.CreateTextFile(tmpDir, "test3", true)
	expect(t, e.Eq(ec.Type),
		`fu.CreateTextFile(tmpDir, ..., true): `+
			`NOT returned ec.Type, got '%v'`, e)
}

func TestCopy(t *testing.T) {
	// Create dedicated directory for this test
	localTmpDir := createTestDir("copy_test")
	printf("* TestCopy(): using temp dir '%s'\n", localTmpDir)

	// When src does not exist,
	// should return ec.NotFound.
	shouldNotExist := join(localTmpDir, "should_not_exist.txt")
	e := fu.Copy(nonExistingPath, shouldNotExist)
	expect(t, e.Eq(ec.NotFound),
		`Copy(nonExistingPath, shouldNotExist): `+
			`expected ec.NotFound, got '%v'`, e)
	dest1Exists, e := fu.FileExists(shouldNotExist)
	expect(t, !dest1Exists && e.None())

	// When src is a dir and dest does not exist,
	// should return NoError.
	testdataCopy1 := join(localTmpDir, "testdata_copy_1")
	e = fu.Copy(testDirTreeRoot, testdataCopy1)
	expect(t, e.None(),
		`Copy(testdataSrcDir, testdataDestDir): `+
			`expected ec.NoError, got '%v'`, e)
	testdataDestDirExists, e := fu.DirExists(testdataCopy1)
	expect(t, testdataDestDirExists && e.None())
	intact := fstestutils.AssertAllSourceItemsConsistent(testdataCopy1)
	expect(t, intact, `testdataDestDir consistency check failed`)

	// Create a text file to be used in subsequent tests
	anotherExistingFile := join(localTmpDir, "another_existing_file.txt")
	e = fu.CreateTextFile(anotherExistingFile, "another_existing_file", false)
	expect(t, e.None(),
		`Copy(): fu.CreateTextFile(anotherExistingFile, ...) failed '%v'`, e)

	// When src is a file, dest is an existing file and using default options,
	// should return ec.AlreadyExists.
	e = fu.Copy(existingFile, anotherExistingFile)
	expect(t, e.Eq(ec.AlreadyExists),
		`Copy(existingFile, anotherExistingFile): `+
			`expected ec.AlreadyExists, got '%v'`, e)

	// When src is a file, dest is an existing dir and using default options,
	// should return ec.Type.
	e = fu.Copy(existingFile, testdataCopy1)
	expect(t, e.Eq(ec.Type),
		`Copy(existingFile, testdataDestDir): `+
			`expected ec.Type, got '%v'`, e)

	// Test custom Skip function.
	// When src is a dir, dest not exists and skip function specified,
	// should skip specified items and return NoError.
	testdataWithSkipDir := join(localTmpDir, "testdata_with_skip")
	e = fu.Copy(testDirTreeRoot, testdataWithSkipDir, fu.CopyOptions{
		Skip: func(src string) (bool, error) {
			if strings.Contains(src, "dir_a") {
				return true, nil
			}
			return false, nil
		},
	})
	expect(t, e.None(),
		`Copy(testdataSrcDir, ..., Skip dir_a): `+
			`expected NoError, got '%v'`, e)
	testdataWithSkipDirExists, e := fu.DirExists(testdataWithSkipDir)
	expect(t, testdataWithSkipDirExists && e.None())
	dir_a_path := join(testdataWithSkipDir, "dir_a")
	exists, _, e := fu.PathExists(dir_a_path)
	expect(t, !exists, `Copy(testdataSrcDir, ..., Skip dir_a): `+
		`testdataDestDir still contains "dir_a"`)

	// Test other overwrite modes

	// When fu.MERGE specified, all files common to source and dest
	// should NOT be overwritten, items unique to
	// merge dest should be intact.
	mdrMerge := join(localTmpDir, "mdrMerge")
	e = fstestutils.CreatePreExistingDestination(mdrMerge)
	expect(t, e.None())
	overwriteOptions := fu.CopyOptions{
		OverwriteMode: fu.MERGE,
	}
	e = fu.Copy(testDirTreeRoot, mdrMerge, overwriteOptions)
	expect(t, e.None())
	// All items unique to source must be copied to the dest.
	consistent := fstestutils.AssertUniqueToSourceItemsConsistent(mdrMerge)
	expect(t, consistent)
	// All items that already exist in dest must be intact.
	allPreExistingItemsIntact := fstestutils.AssertAllPreExistingItemsConsistent(mdrMerge)
	expect(t, allPreExistingItemsIntact)

	// When fu.OVERWRITE_INTERSECTION specified, all common files in
	// mdrSoftOverwrite should be overwritten, items unique to
	// merge dest should be intact.
	mdrOverwriteIntersection := join(localTmpDir, "mdrOverwriteIntersection")
	e = fstestutils.CreatePreExistingDestination(mdrOverwriteIntersection)
	expect(t, e.None())

	overwriteOptions = fu.CopyOptions{
		OverwriteMode: fu.OVERWRITE_INTERSECTION,
	}
	e = fu.Copy(testDirTreeRoot, mdrOverwriteIntersection, overwriteOptions)
	expect(t, e.None())
	// The whole test dir tree must be consistent
	consistent = fstestutils.AssertAllSourceItemsConsistent(mdrOverwriteIntersection)
	expect(t, consistent)
	// Items unique to merge dest should be intact.
	uniqueItemsIntact := fstestutils.AssertUniquePreExistingItemsConsistent(mdrOverwriteIntersection)
	expect(t, uniqueItemsIntact)

	// When fu.OVERWRITE_FULL specified, all common files in
	// mdrSoftOverwrite should be overwritten, items unique to
	// merge dest should no longer exist.
	mdrOverwriteFull := join(localTmpDir, "mdrOverwriteFull")
	e = fstestutils.CreatePreExistingDestination(mdrOverwriteFull)
	expect(t, e.None())
	overwriteOptions = fu.CopyOptions{
		OverwriteMode: fu.OVERWRITE_FULL,
	}
	e = fu.Copy(testDirTreeRoot, mdrOverwriteFull, overwriteOptions)
	expect(t, e.None())
	// The whole test dir tree must be consistent
	consistent = fstestutils.AssertAllSourceItemsConsistent(mdrOverwriteFull)
	expect(t, consistent)
	// Items unique to merge dest should no longer exist.
	uniqueItemsIntact = fstestutils.AssertUniquePreExistingItemsConsistent(mdrOverwriteFull)
	expect(t, !uniqueItemsIntact)
}

// Returns true when a==b
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, s := range a {
		if s != b[i] {
			return false
		}
	}
	return true
}

func TestReadLines(t *testing.T) {
	tmpDir := createTestDir("test_read_into_lines")
	printf("* TestReadLines(): using temp dir '%s'\n", tmpDir)

	// Create a few text files with different new line separators
	testData := []struct {
		Path   string
		Data   string
		Verify []string
	}{
		{
			Path:   join(tmpDir, "no_newline.txt"),
			Data:   "line1",
			Verify: []string{"line1"},
		},
		{
			Path:   join(tmpDir, "unix.txt"),
			Data:   "line1\nline2\nline3\n",
			Verify: []string{"line1", "line2", "line3"},
		},
		{
			Path:   join(tmpDir, "windows.txt"),
			Data:   "line1\r\nline2\r\nline3\r\n",
			Verify: []string{"line1", "line2", "line3"},
		},
		{
			Path:   join(tmpDir, "mixed.txt"),
			Data:   "line1\nline2\r\nline3\n",
			Verify: []string{"line1", "line2", "line3"},
		},
	}
	for _, td := range testData {
		e := fu.CreateTextFile(td.Path, td.Data, false)
		expect(t, e.None(),
			`fu.CreateTextFile(): `+
				`failed to create file '%s', '%v'`, td.Path, e)
	}
	// Read the files and verify data
	for _, td := range testData {
		c, e := fu.ReadLines(td.Path)
		expect(t, e.None(),
			`ReadLines(): `+
				`failed to read file '%s', '%v'`, td.Path, e)
		expect(t, stringSlicesEqual(c, td.Verify),
			`ReadLines(): contents verification failed for file '%s': `+
				`expected '%v', got '%v'`, td.Path, td.Verify, c)
	}

	// UNIX-ONLY
	// When reading symlink.txt, should contain single line 'test'
	c, e := fu.ReadLines(symlinkToFile)
	expect(t, e.None(),
		`ReadLines(): `+
			`failed to read symlink to file '%s', '%v'`, symlinkToFile, e)
	expectedResult := []string{"b.txt"}
	expect(t, stringSlicesEqual(c, expectedResult),
		`ReadLines(): contents verification failed for file '%s': `+
			`expected '%v', got '%v'`, symlinkToFile, expectedResult, c)
}

func TestSymlinks(t *testing.T) {
	// UNIX-ONLY
	tmpDir := createTestDir("test_symlinks")
	printf("* TestSymlinks(): using temp dir '%s'\n", tmpDir)
	// Create a symlink
	newSymlink := join(tmpDir, "symlink.txt")
	err := os.Symlink(existingFile, newSymlink)
	expect(t, err == nil, `TestSymlinks(): `+
		`failed to create symlink '%s', '%v'`, newSymlink, err)
	// Verify that new symlink exists
	newSymlinkExists, e := fu.SymlinkExists(newSymlink)
	expect(t, (newSymlinkExists && e.None()),
		`SymlinkExists(): `+
			`returned negative result '%s', '%v'`, newSymlink, e)

	// When removing symlink, the original file must not be removed
	err = os.Remove(newSymlink)
	expect(t, err == nil, `TestSymlinks(): `+
		`failed to remove symlink '%s', '%v'`, newSymlink, err)
	targetFileExists, e := fu.FileExists(existingFile)
	expect(t, (targetFileExists && e.None()), `TestSymlinks(): `+
		`original file removed with symlink '%s', '%v'`, existingFile, e)
	// Symlink must not exist
	newSymlinkExists, e = fu.SymlinkExists(newSymlink)
	expect(t, (!newSymlinkExists && e.None()),
		`SymlinkExists(): `+
			`failed to remove symlink '%s', '%v'`, newSymlink, e)
}

func TestHardlinks(t *testing.T) {
	// UNIX-ONLY
	tmpDir := createTestDir("test_hardlinks")
	printf("* TestHardlinks(): using temp dir '%s'\n", tmpDir)
	// Create a hardlink fo file
	newHardlink := join(tmpDir, "hardlink.txt")
	err := os.Link(existingFile, newHardlink)
	expect(t, err == nil, `TestHardlinks(): `+
		`failed to create hardlink '%s', '%v'`, newHardlink, err)
	// Verify that new hardlink exists
	newHardlinkExists, e := fu.HardlinkExists(newHardlink)
	expect(t, (newHardlinkExists && e.None()),
		`HardlinkExists(): `+
			`returned negative result '%s', '%v'`, newHardlink, e)
	// When removing hardlink, the original file must not be removed
	err = os.Remove(newHardlink)
	expect(t, err == nil, `TestHardlinks(): `+
		`failed to remove hardlink '%s', '%v'`, newHardlink, err)
	targetFileExists, e := fu.FileExists(existingFile)
	expect(t, (targetFileExists && e.None()), `TestHardlinks(): `+
		`original file removed with hardlink '%s', '%v'`, existingFile, e)
	// Hardlink must not exist
	newHardlinkExists, e = fu.HardlinkExists(newHardlink)
	expect(t, (!newHardlinkExists && e.None()),
		`HardlinkExists(): `+
			`failed to remove hardlink '%s', '%v'`, newHardlink, e)

	// IT SHOULD NOT BE POSSIBLE TO CREATE A HARDLINK TO DIRECTORY
	newHardlink = join(tmpDir, "hardlink_a")
	err = os.Link(join(testDirTreeRoot, "dir_a"), newHardlink)
	expect(t, err != nil, `TestHardlinks(): `+
		`create hardlink to directory expected to fail: '%s', '%v'`,
		newHardlink, err)
	// Verify that new hardlink not exists
	newHardlinkExists, e = fu.HardlinkExists(newHardlink)
	expect(t, (!newHardlinkExists && e.None()),
		`HardlinkExists(): `+
			`hardlink to directory exists but it should NOT '%s', '%v'`, newHardlink, e)
}

func TestNamedPipes(t *testing.T) {
	// UNIX-ONLY
	// TODO:

}

func TestFsItemType(t *testing.T) {
	testData := []fu.FsItemType{fu.TYPE_UNKNOWN, fu.TYPE_FILE, fu.TYPE_DIR,
		fu.TYPE_SYMLINK, fu.TYPE_HARDLINK, fu.TYPE_NAMED_PIPE}
	for _, st := range testData {
		expect(t, len(st.String()) != 0,
			"string representation expected to have non-zero length")
	}
}
