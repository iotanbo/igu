package zip

// https://github.com/mholt/archiver

// Zip Slip vulnerability:
// https://snyk.io/research/zip-slip-vulnerability
import (
	"compress/flate"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"

	"github.com/iotanbo/igu/pkg/ec"
	"github.com/iotanbo/igu/pkg/fu"

	//lint:ignore ST1001 - for concise error handling.
	. "github.com/iotanbo/igu/pkg/errs"
)

// OverwriteMode:
//	NoOverwrite // Skip operation if destination exists.
//	Merge // Do not delete anything, only add files that are new.
//	SoftOverwrite // Do not delete destination directory, but overwrite individual files.
//	FullOverwrite // Delete destination and its contents, then copy new files.
type OverwriteMode int

const (
	// Skip operation if destination exists.
	NoOverwrite OverwriteMode = iota
	// Do not delete any files inside destination, only add files that are new to destination.
	Merge
	// Do not delete destination directory, but overwrite individual files.
	// Other files (if any) leave intact.
	SoftOverwrite
	// Delete destination and its contents, then copy new files.
	FullOverwrite
)

// Zips "srcPath" into "destPath".
//
// Returned errors:
//	ec.NoError // success
//	ec.NotFound // srcPath does not exist
//	ec.AlreadyExists // destPath already exists
// May return other errors for other situations.
func Zip(srcPath string, destPath string) Err {
	srcPath = strings.TrimPrefix(srcPath, "./")

	exists, _, e := fu.PathExists(srcPath)
	if e.Some() {
		return e
	}
	if !exists {
		return Err{Code: ec.NotFound, Msg: "source " + srcPath}
	}
	exists, _, e = fu.PathExists(destPath)
	if e.Some() {
		return e
	}
	if exists {
		return Err{Code: ec.AlreadyExists, Msg: "destination " + destPath}
	}

	z := archiver.Zip{
		CompressionLevel:       flate.DefaultCompression,
		MkdirAll:               true,
		SelectiveCompression:   true,
		ContinueOnError:        false,
		OverwriteExisting:      false,
		ImplicitTopLevelFolder: false,
	}

	if err := z.Archive([]string{srcPath}, destPath); err != nil {
		return FromError(err)
	}

	return NoError
}

// Based on https://stackoverflow.com/a/24792688/3824328
func Unarchive(srcPath, destPath string,
	overwriteMode OverwriteMode) Err {

	srcPath = filepath.Clean(srcPath)
	destPath = filepath.Clean(destPath)

	srcExists, _, e := fu.PathExists(srcPath)
	if e.Some() {
		return e
	}
	if !srcExists {
		return Err{Code: ec.NotFound, Msg: "source " + srcPath}
	}
	destExists, _, e := fu.PathExists(destPath)
	if e.Some() {
		return e
	}

	if destExists {
		fmt.Printf("* Destination already exists: '%s'\n", destPath)
		// In NoOverwrite mode return ec.AlreadyExists
		if overwriteMode == NoOverwrite {
			return Err{Code: ec.AlreadyExists, Msg: "dest " + srcPath}
		}
		// In HardOverwrite mode remove destination and all its contents
		if overwriteMode == FullOverwrite {
			if err := os.RemoveAll(destPath); err != nil {
				rmError := FromError(err)
				rmError.Msg = "failed to remove destination " + destPath
				return rmError
			}
		}
	}
	os.MkdirAll(destPath, 0755)

	if err := archiver.Unarchive(srcPath, destPath); err != nil {
		return FromError(err)
	}

	//fu.Copy(fu.NO_OVERWRITE)
	return NoError
}

// Based on https://stackoverflow.com/a/24792688/3824328
func UnTarGzip(srcPath, destPath string,
	overwriteMode OverwriteMode) Err {

	srcPath = filepath.Clean(srcPath)
	destPath = filepath.Clean(destPath)

	srcExists, _, e := fu.PathExists(srcPath)
	if e.Some() {
		return e
	}
	if !srcExists {
		return Err{Code: ec.NotFound, Msg: "source " + srcPath}
	}
	destExists, _, e := fu.PathExists(destPath)
	if e.Some() {
		return e
	}

	if destExists {
		fmt.Printf("* Destination already exists: '%s'\n", destPath)
		// In NoOverwrite mode return ec.AlreadyExists
		if overwriteMode == NoOverwrite {
			return Err{Code: ec.AlreadyExists, Msg: "dest " + srcPath}
		}
		// In HardOverwrite mode remove destination and all its contents
		if overwriteMode == FullOverwrite {
			if err := os.RemoveAll(destPath); err != nil {
				rmError := FromError(err)
				rmError.Msg = "failed to remove destination " + destPath
				return rmError
			}
		}
	}
	os.MkdirAll(destPath, 0755)

	// TODO:

	return NoError
}

// See https://github.com/mimoo/eureka/blob/master/folders.go
