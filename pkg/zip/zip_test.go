package zip

import (
	//"fmt"

	"os"
	"testing"

	"github.com/iotanbo/igu/pkg/ec"
	"github.com/iotanbo/igu/pkg/fu"
	"github.com/stretchr/testify/require"
)

// Type aliases to improve readability
//var printf = fmt.Printf
var expect = require.True

func TestZipUnzip(t *testing.T) {
	// Zip an existing file
	srcFile := "./_testdata/testfile.txt"
	destFile := "./_testdata/temp/testfile.zip"
	destExists, _, e := fu.PathExists(destFile)
	expect(t, e.None())
	if destExists {
		os.Remove(destFile)
	}
	e = Zip(srcFile, destFile)
	expect(t, e.None(), "Zip(srcFile, destFile): Expected NoError, got %v.", e)

	// Zip an existing directory
	srcDir := "./_testdata/testdir"
	destDirArch := "./_testdata/temp/testdir.zip"
	destExists, _, e = fu.PathExists(destDirArch)
	expect(t, e.None())
	if destExists {
		os.Remove(destDirArch)
	}
	e = Zip(srcDir, destDirArch)
	expect(t, e.None(), "Zip(srcDir, destDirArch): Expected NoError, got %v.", e)

	// Unzip file
	e = Unarchive(destFile,
		"./_testdata/temp/extracted/",
		FullOverwrite,
	)
	expect(t, e.None())
	// read the unzipped file and verify its contents
	contents, e := fu.ReadTextFile("./_testdata/temp/extracted/testfile.txt")
	expect(t, e.None())
	expect(t, contents == "testfile.txt")

	// TEST OVERWRITE MODES

	// When overwriteMode == Merge and destPath exists,
	// old files and folders must be kept intact.
	e = fu.CreateTextFile("./_testdata/temp/extracted/testdir/already_exists.txt", "already_exists.txt", true)
	e = fu.CreateTextFile("./_testdata/temp/extracted/testdir/testfile.txt", "previous_contents", true)
	expect(t, e.None(), "Error: %v", e)
	arch := "_testdata/temp/testdir.zip"
	e = Unarchive(arch,
		"./_testdata/temp/extracted",
		Merge,
	)
	expect(t, e.None(), "Error: %v", e)
	// Check that old files are intact
	contents, e = fu.ReadTextFile("./_testdata/temp/extracted/testdir/already_exists.txt")
	expect(t, e.None())
	expect(t, contents == "already_exists.txt")

	contents, e = fu.ReadTextFile("./_testdata/temp/extracted/testdir/testfile.txt")
	expect(t, e.None())
	expect(t, contents == "previous_contents")

	// When overwriteMode == SoftOverwrite and destPath exists,
	// existing old files must be overwritten with new ones.
	e = Unarchive(arch,
		"./_testdata/temp/extracted",
		SoftOverwrite,
	)
	expect(t, e.None())
	// This file must be intact
	contents, e = fu.ReadTextFile("./_testdata/temp/extracted/testdir/already_exists.txt")
	expect(t, e.None())
	expect(t, contents == "already_exists.txt")
	// This file must be overwritten
	contents, e = fu.ReadTextFile("./_testdata/temp/extracted/testdir/testfile.txt")
	expect(t, e.None())
	expect(t, contents == "testfile.txt")

	// When overwriteMode == FullOverwrite and destPath exists,
	// destination must be deleted and fully overwritten with new ones.
	e = Unarchive(arch,
		"./_testdata/temp/extracted",
		FullOverwrite,
	)
	expect(t, e.None())
	// This file must no longer exist
	_, e = fu.ReadTextFile("./_testdata/temp/extracted/testdir/already_exists.txt")
	expect(t, e.Code == ec.NotFound)
	// This file must be overwritten
	contents, e = fu.ReadTextFile("./_testdata/temp/extracted/testdir/testfile.txt")
	expect(t, e.None())
	expect(t, contents == "testfile.txt")

	// When overwriteMode == NoOverwrite and destPath exists,
	// destination must be kept intact and ec.AlreadyExists must be returned.
	e = Unarchive(arch,
		"./_testdata/temp/extracted",
		NoOverwrite,
	)
	expect(t, e.Code == ec.AlreadyExists)
}
