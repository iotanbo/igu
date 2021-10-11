package fu

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/iotanbo/igu/pkg/ec"
	"github.com/stretchr/testify/require"
	//. "github.com/iotanbo/igu/pkg/errs"
)

var globalTmpDir string

// Setting this value to false allows to preserve
// globalTmpDir after tests are complete.
var _removeTmpDirAfterTests = false

// Type aliases to improve readability
var join = filepath.Join
var mkdir = os.Mkdir
var printf = fmt.Printf
var expect = require.True

// var expectNot = require.False

// Initialized in TestMain
var testdataSrcDir string
var existingFile string
var existingDir string
var nonExistingPath string

// Platform-specific variables
var symlinkToFile string
var existingHardlink string
var existingNamedPipe string

// anyContents is used while verifying file contents
// to specify that any file contents will do.
var anyContents = "**any**"

// containsBinData is used while verifying file contents
// to specify that binary data must be verified.
var containsBinData = "**bindata**"

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

	testdataSrcDir = join(tmpDir, "testdata")
	// Copy testData
	e := Copy("_testdata", testdataSrcDir)
	if e.Some() {
		panic(fmt.Sprintf("TestMain(): failed to copy FU testdata: %v", e))
	}
	existingFile = join(testdataSrcDir, "test.txt")
	existingDir = testdataSrcDir
	nonExistingPath = join(testdataSrcDir, "not_exists")

	// Create symlinks
	symlinkToFile = join(testdataSrcDir, "symlink.txt")
	if runtime.GOOS != "windows" {
		symlinksToBeCreated := []struct {
			Src  string
			Dest string
		}{
			{
				Src:  existingFile,
				Dest: symlinkToFile,
			},
			{
				Src:  join(testdataSrcDir, "dir_a"),
				Dest: join(testdataSrcDir, "dir_b", "symlink_to_a"),
			},
			{
				Src:  join(testdataSrcDir, "dir_b"),
				Dest: join(testdataSrcDir, "dir_a", "symlink_to_b"),
			},
		}

		for _, s := range symlinksToBeCreated {
			err := os.Symlink(s.Src, s.Dest)
			if err != nil {
				panic(fmt.Sprintf("TestMain(): failed to create symlink '%s': %v",
					symlinkToFile, err))
			}
		}
	}

	exitVal := m.Run()

	if _removeTmpDirAfterTests {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"can't remove temporary directory for FU tests: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("--- temporary directory removed, FU package tests complete.")
	}
	os.Exit(exitVal)
}

func TestGetItemType(t *testing.T) {
	//printf("* TestGetItemType(): using temp dir '%s'\n", globalTmpDir)
	// Create dedicated directory for this test
	tmpDir := createTestDir("test_get_item_type")
	printf("* TestGetItemType(): using temp dir '%s'\n", tmpDir)

	// When passing empty string, should return (TYPE_UNKNOWN, ec.NotFound).
	r, e := GetItemType("")
	expect(t, e.Eq(ec.NotFound),
		`GetItemType(''): expected ec.NotFound, got '%v'`, e)
	expect(t, r == TYPE_UNKNOWN,
		`GetItemType(''): expected TYPE_UNKNOWN, got '%v'`, r)

	// When passing existing file, should return (TYPE_FILE, NoError).
	r, e = GetItemType(existingFile)
	expect(t, e.None(),
		`GetItemType(existingFile): returned error '%v'`, e)
	expect(t, r == TYPE_FILE,
		`GetItemType(existingFile): expected TYPE_FILE, got '%v'`, r)

	// When passing non-existing file, should return (TYPE_UNKNOWN, ec.NotFound).
	r, e = GetItemType(nonExistingPath)
	expect(t, e.Eq(ec.NotFound),
		`GetItemType(nonExistingPath): expected ec.NotFound, got '%v'`, e)
	expect(t, r == TYPE_UNKNOWN,
		`GetItemType(nonExistingPath): expected TYPE_UNKNOWN, got '%v'`, r)

	// When passing directory, should return (TYPE_DIR, NoError).
	r, e = GetItemType(testdataSrcDir)
	expect(t, e.Eq(ec.NoError),
		`GetItemType(existingDir): expected ec.NoError, got '%v'`, e)
	expect(t, r == TYPE_DIR,
		`GetItemType(existingDir): expected TYPE_DIR, got '%v'`, r)

	// UNIX-ONLY
	// When passing symlink, should return (TYPE_SYMLINK, NoError).
	r, e = GetItemType(symlinkToFile)
	expect(t, e.Eq(ec.NoError),
		`GetItemType(symlinkToFile): expected ec.NoError, got '%v'`, e)
	expect(t, r == TYPE_SYMLINK,
		`GetItemType(symlinkToFile): expected TYPE_SYMLINK, got '%v'`, r)

	// TODO: test TYPE_NAMED_PIPE
}

func TestFileExists(t *testing.T) {
	printf("* TestFileExists(): using temp dir '%s'\n", globalTmpDir)
	// When passing existing file, should return (true, no error)
	r, e := FileExists(existingFile)
	expect(t, e.None(),
		`FileExists(existingFile): returned error '%v'`, e)
	expect(t, r, "FileExists(existingFile): returned false")

	// When passing non-existing file, should return (false, no error)
	r, e = FileExists(nonExistingPath)
	expect(t, e.None(),
		`FileExists(nonExistingPath): returned error '%v'`, e)
	expect(t, !r, `FileExists(nonExistingPath): returned true`)

	// When passing empty string, should return (false, no error)
	r, e = FileExists("")
	expect(t, e.None(),
		`FileExists(''): returned unexpected error '%v'`, e)
	expect(t, !r, `FileExists(''): returned true`)

	// When passing existing directory, should return (false, ec.Type)
	r, e = FileExists(testdataSrcDir)
	expect(t, e.Eq(ec.Type),
		`FileExists(existingDir): didn't return ec.Type error, got '%v'`, e)
	expect(t, !r, "FileExists(existingDir): returned true")

	// UNIX-ONLY
	// When passing symlink, should return (true, no error)
	r, e = FileExists(symlinkToFile)
	expect(t, e.None(),
		`FileExists(symlinkToFile): returned error '%v'`, e)
	expect(t, r, "FileExists(symlinkToFile): returned false")

}

func TestCreateTextFile(t *testing.T) {
	// Create dedicated directory for this test
	tmpDir := createTestDir("test_create_text_file")
	printf("* TestCreateTextFile(): using temp dir '%s'\n", tmpDir)

	// Normally should create a file with given contents and return NoError
	path := join(tmpDir, "text_file.txt")
	e := CreateTextFile(path, "test", false)
	expect(t, e.Eq(ec.NoError),
		`CreateTextFile("text_file.txt", ..., false): `+
			`NOT returned ec.NoError, got '%v'`, e)
	exists, e := FileExists(path)
	expect(t, e.None())
	expect(t, exists) // TODO: verify file contents

	// When insufficient permissions should return ec.PermissionDenied
	e = CreateTextFile("/dummy.txt", "test", false)
	expect(t, e.Eq(ec.PermissionDenied),
		`CreateTextFile("/dummy.txt", ..., false): `+
			`NOT returned ec.PermissionDenied, got '%v'`, e)

	// When file already exists should return ec.AlreadyExists
	e = CreateTextFile(path, "test", false)
	expect(t, e.Eq(ec.AlreadyExists),
		`CreateTextFile(path, ..., false): `+
			`NOT returned ec.AlreadyExists, got '%v'`, e)

	// When file already exists and overwrite=true should return ec.NoError
	e = CreateTextFile(path, "test2", true)
	expect(t, e.Eq(ec.NoError),
		`CreateTextFile(path, ..., true): `+
			`NOT returned ec.NoError, got '%v'`, e)

	// When destination already exists but is a directory
	// and overwrite=true, should return ec.Type
	e = CreateTextFile(tmpDir, "test3", true)
	expect(t, e.Eq(ec.Type),
		`CreateTextFile(tmpDir, ..., true): `+
			`NOT returned ec.Type, got '%v'`, e)
}

func dirExists(path string) bool {
	if exists, e := DirExists(path); !(exists && e.None()) {
		return false
	}
	return true
}

// Returns true if file with specified contents exist.
// Prints debug error message to stdout if validation failed.
func validateFileContents(path, contents string, binContents []byte) bool {
	if exists, e := FileExists(path); !(exists && e.None()) {
		printf("* validateFileContents(): file not exists '%s', %v\n",
			path, e)
		return false
	}
	if contents == anyContents {
		return true
	} else if contents == containsBinData {
		fc, e := ReadBinFile(path)
		if e.Some() {
			printf("* validateFileContents(): unexpected error "+
				"while reading bindata '%s', %v\n", path, e)
			return false
		}
		if bytes.Equal(fc, binContents) {
			return true
		}
		printf("* validateFileContents(): error '%s', expected '%v',"+
			"actual '%v'\n", path, contents, fc)
		return false
	}
	fc, e := ReadTextFile(path)
	if e.Some() {
		printf("* validateFileContents(): unexpected error "+
			"while reading '%s', %v\n", path, e)
		return false
	}
	if fc == contents {
		return true
	}

	printf("* validateFileContents(): error '%s', expected '%s', "+
		"actual '%s'\n", path, contents, fc)
	return false
}

// Returns true if testDataDir contains
// required files and subdirectories and their contents is valid.
func checkTestdataDirConsistency(testDataDir string) bool {
	dirsToVerify := []string{
		testDataDir,
		join(testDataDir, ".hidden_dir"),
		join(testDataDir, "dir_a"),
		join(testDataDir, "dir_b"),
	}
	for _, d := range dirsToVerify {
		if !dirExists(d) {
			printf("* checkTestdataDirConsistency(): "+
				"dir doesn't exist: '%s'\n", d)
			return false
		}
	}

	filesToVerify := []struct {
		Path    string
		Data    string
		BinData []byte
	}{
		{
			Path: join(testDataDir, "info.txt"),
			Data: anyContents,
		},
		{
			Path: join(testDataDir, "test.txt"),
			Data: "test",
		},
		{
			Path: join(testDataDir, "dir_a", "a.txt"),
			Data: "a.txt",
		},
		{
			Path: join(testDataDir, "dir_b", "b.txt"),
			Data: "b.txt",
		},
		{
			Path: join(testDataDir, ".hidden_dir", ".hidden_file.txt"),
			Data: ".hidden_file.txt",
		},
	}
	for _, f := range filesToVerify {
		if !validateFileContents(f.Path, f.Data, f.BinData) {
			return false
		}
	}

	return true
}

func TestCopy(t *testing.T) {
	// Create dedicated directory for this test
	tmpDir := createTestDir("copy_test")
	printf("* TestCopy(): using temp dir '%s'\n", tmpDir)

	// When src does not exist,
	// should return ec.NotFound.
	shouldNotExist := join(tmpDir, "should_not_exist.txt")
	e := Copy(nonExistingPath, shouldNotExist)
	expect(t, e.Eq(ec.NotFound),
		`Copy(nonExistingPath, shouldNotExist): `+
			`expected ec.NotFound, got '%v'`, e)
	dest1Exists, e := FileExists(shouldNotExist)
	expect(t, !dest1Exists && e.None())

	// When src is a dir and dest does not exist,
	// should return NoError.
	testdataDestDir := join(tmpDir, "testdata")
	e = Copy(testdataSrcDir, testdataDestDir)
	expect(t, e.Eq(ec.NoError),
		`Copy(testdataSrcDir, testdataDestDir): `+
			`expected ec.NoError, got '%v'`, e)
	testdataDestDirExists, e := DirExists(testdataDestDir)
	expect(t, testdataDestDirExists && e.None())
	consistent := checkTestdataDirConsistency(testdataDestDir)
	expect(t, consistent, `testdataDestDir consistency check failed`)

	// Create a text file to be used in subsequent tests
	anotherExistingFile := join(tmpDir, "another_existing_file.txt")
	e = CreateTextFile(anotherExistingFile, "another_existing_file", false)
	expect(t, e.None(),
		`Copy(): CreateTextFile(anotherExistingFile, ...) failed '%v'`, e)

	// When src is a file, dest is an existing file and using default options,
	// should return ec.AlreadyExists.
	e = Copy(existingFile, anotherExistingFile)
	expect(t, e.Eq(ec.AlreadyExists),
		`Copy(existingFile, anotherExistingFile): `+
			`expected ec.AlreadyExists, got '%v'`, e)

	// When src is a file, dest is an existing dir and using default options,
	// should return ec.Type.
	e = Copy(existingFile, testdataDestDir)
	expect(t, e.Eq(ec.Type),
		`Copy(existingFile, testdataDestDir): `+
			`expected ec.Type, got '%v'`, e)

	// Test custom Skip function.
	// When src is a dir, dest not exists and skip function specified,
	// should skip specified items and return NoError.
	testdataWithSkipDir := join(tmpDir, "testdata_with_skip")
	e = Copy(testdataSrcDir, testdataWithSkipDir, CopyOptions{
		Skip: func(src string) (bool, error) {
			if strings.Contains(src, "dir_a") {
				return true, nil
			}
			return false, nil
		},
	})
	expect(t, e.Eq(ec.NoError),
		`Copy(testdataSrcDir, ..., Skip dir_a): `+
			`expected NoError, got '%v'`, e)
	testdataWithSkipDirExists, e := DirExists(testdataWithSkipDir)
	expect(t, testdataWithSkipDirExists && e.None())
	dir_a_path := join(testdataWithSkipDir, "dir_a")
	exists, _, e := PathExists(dir_a_path)
	expect(t, !exists, `Copy(testdataSrcDir, ..., Skip dir_a): `+
		`testdataDestDir still contains "dir_a"`)
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
		e := CreateTextFile(td.Path, td.Data, false)
		expect(t, e.Eq(ec.NoError),
			`CreateTextFile(): `+
				`failed to create file '%s', '%v'`, td.Path, e)
	}
	// Read the files and verify data
	for _, td := range testData {
		c, e := ReadLines(td.Path)
		expect(t, e.Eq(ec.NoError),
			`ReadLines(): `+
				`failed to read file '%s', '%v'`, td.Path, e)
		expect(t, stringSlicesEqual(c, td.Verify),
			`ReadLines(): contents verification failed for file '%s': `+
				`expected '%v', got '%v'`, td.Path, td.Verify, c)
	}

	// UNIX-ONLY
	// When reading symlink.txt, should contain single line 'test'
	c, e := ReadLines(symlinkToFile)
	expect(t, e.Eq(ec.NoError),
		`ReadLines(): `+
			`failed to read symlink to file '%s', '%v'`, symlinkToFile, e)
	expectedResult := []string{"test"}
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
	newSymlinkExists, e := SymlinkExists(newSymlink)
	expect(t, (newSymlinkExists && e.None()),
		`SymlinkExists(): `+
			`returned negative result '%s', '%v'`, newSymlink, e)

	// When removing symlink, the original file must not be removed
	err = os.Remove(newSymlink)
	expect(t, err == nil, `TestSymlinks(): `+
		`failed to remove symlink '%s', '%v'`, newSymlink, err)
	targetFileExists, e := FileExists(existingFile)
	expect(t, (targetFileExists && e.None()), `TestSymlinks(): `+
		`original file removed with symlink '%s', '%v'`, existingFile, e)
	// Symlink must not exist
	newSymlinkExists, e = SymlinkExists(newSymlink)
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
	newHardlinkExists, e := HardlinkExists(newHardlink)
	expect(t, (newHardlinkExists && e.None()),
		`HardlinkExists(): `+
			`returned negative result '%s', '%v'`, newHardlink, e)
	// When removing hardlink, the original file must not be removed
	err = os.Remove(newHardlink)
	expect(t, err == nil, `TestHardlinks(): `+
		`failed to remove hardlink '%s', '%v'`, newHardlink, err)
	targetFileExists, e := FileExists(existingFile)
	expect(t, (targetFileExists && e.None()), `TestHardlinks(): `+
		`original file removed with hardlink '%s', '%v'`, existingFile, e)
	// Hardlink must not exist
	newHardlinkExists, e = HardlinkExists(newHardlink)
	expect(t, (!newHardlinkExists && e.None()),
		`HardlinkExists(): `+
			`failed to remove hardlink '%s', '%v'`, newHardlink, e)

	// IT SHOULD NOT BE POSSIBLE TO CREATE A HARDLINK TO DIRECTORY
	newHardlink = join(tmpDir, "hardlink_a")
	err = os.Link(join(testdataSrcDir, "dir_a"), newHardlink)
	expect(t, err != nil, `TestHardlinks(): `+
		`create hardlink to directory expected to fail: '%s', '%v'`,
		newHardlink, err)
	// Verify that new hardlink not exists
	newHardlinkExists, e = HardlinkExists(newHardlink)
	expect(t, (!newHardlinkExists && e.None()),
		`HardlinkExists(): `+
			`hardlink to directory exists but it should NOT '%s', '%v'`, newHardlink, e)
}

func TestNamedPipes(t *testing.T) {
	// UNIX-ONLY

}

func TestFsItemType(t *testing.T) {
	testData := []FsItemType{TYPE_UNKNOWN, TYPE_FILE, TYPE_DIR,
		TYPE_SYMLINK, TYPE_HARDLINK, TYPE_NAMED_PIPE}
	for _, st := range testData {
		expect(t, len(st.String()) != 0,
			"string representation expected to have non-zero length")
	}
}
