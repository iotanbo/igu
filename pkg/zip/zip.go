package zip

// import (
// 	//"github.com/iotanbo/igu/pkg/ec"
// 	//lint:ignore ST1001 - for clear and concise error handling.
// 	. "github.com/iotanbo/igu/pkg/errs"
// )

// // Zip creates .zip archive from srcPath and saves it as a destPath.
// func Zip(srcPath string, destPath string) Err {
// 	return NoError
// }

// Based on https://stackoverflow.com/a/63233911/3824328
// For alternative possible implementation, see
// https://github.com/mimoo/eureka/blob/master/folders.go
import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/iotanbo/igu/pkg/ec"
	"github.com/iotanbo/igu/pkg/fu"

	//lint:ignore ST1001 - for clear and concise error handling.
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
	file, err := os.Create(destPath)
	if err != nil {
		return FromError(err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	//pathDir := path.Dir(srcPath)
	base, _ := path.Split(srcPath)
	offset := len(base)
	// fmt.Printf("base, offset: '%s', %d\n", base, offset)

	walker := func(path string, info os.FileInfo, err error) error {
		// Remove srcPath root from path
		// fmt.Printf("srcPath, path, offset: '%s', '%s', %d\n", srcPath, path, offset)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		inputFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer inputFile.Close()

		// Ensure that `path` is not absolute; it should not start with "/".
		// This snippet happens to work because I don't use
		// absolute paths, but ensure your real-world code
		// transforms path into a zip-root relative path.
		zipArchPath := path[offset:]
		fmt.Printf("zipArchPath: '%s'\n", zipArchPath)
		f, err := w.Create(zipArchPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, inputFile)
		if err != nil {
			return err
		}

		return nil
	}
	err = filepath.Walk(srcPath, walker)
	if err != nil {
		FromError(err)
	}
	return NoError
}

// Based on https://stackoverflow.com/a/24792688/3824328
func Unzip(srcPath, destPath string,
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

	r, err := zip.OpenReader(srcPath)
	if err != nil {
		return FromError(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(destPath, f.Name)
		fmt.Printf("* Extracting file '%s'\n", path)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(destPath)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			//if overwriteMode != Merge && overwriteMode
			os.MkdirAll(path, f.Mode())
		} else {

			if overwriteMode == Merge {
				// Check if destination file already exists
				exists, _, e := fu.PathExists(path)
				if e.Some() {
					return fmt.Errorf("can't check if path exists: %s", path)
				}
				// Skip overwriting in merge mode
				if exists {
					fmt.Printf("* Merge: file already exists: %s\n", path)
					return nil
				} else {
					fmt.Printf("* Merge: file NOT YET exists: %s\n", path)
				}
			}
			if overwriteMode == SoftOverwrite {
				exists, _, e := fu.PathExists(path)
				if e.Some() {
					return fmt.Errorf("can't check if path exists: %s", path)
				}
				// Remove destination if exists
				if exists {
					if err := os.Remove(path); err != nil {
						return err
					}
				}
			}

			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return FromError(err)
		}
	}

	return NoError
}

// Based on https://stackoverflow.com/a/24792688/3824328
/*
func Unzip(srcPath, destPath string) Err {
	r, err := zip.OpenReader(srcPath)
	if err != nil {
		return FromError(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(destPath, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(destPath, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(destPath)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return FromError(err)
		}
	}

	return NoError
}
*/

/*
Based on https://github.com/mimoo/eureka/blob/master/folders.go

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)


func Compress(src string, buf io.Writer) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	// is file a folder?
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	mode := fi.Mode()
	if mode.IsRegular() {
		// get header
		header, err := tar.FileInfoHeader(fi, src)
		if err != nil {
			return err
		}
		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// get content
		data, err := os.Open(src)
		if err != nil {
			return err
		}
		if _, err := io.Copy(tw, data); err != nil {
			return err
		}
	} else if mode.IsDir() { // folder

		// walk through every file in the folder
		filepath.Walk(src, func(file string, fi os.FileInfo, e error) error {
			// generate tar header
			header, err := tar.FileInfoHeader(fi, file)
			if err != nil {
				return err
			}

			// must provide real name
			// (see https://golang.org/src/archive/tar/common.go?#L626)
			header.Name = filepath.ToSlash(file)

			// write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			// if not a dir, write file content
			if !fi.IsDir() {
				data, err := os.Open(file)
				if err != nil {
					return err
				}
				if _, err := io.Copy(tw, data); err != nil {
					return err
				}
			}
			return nil
		})
	} else {
		return fmt.Errorf("error: file type not supported")
	}

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}
	//
	return nil
}

// check for path traversal and correct forward slashes
func ValidRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}

func Decompress(src io.Reader, dst string) error {
	// ungzip
	zr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	// untar
	tr := tar.NewReader(zr)

	// decompress each element
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		target := header.Name

		// validate name against path traversal
		if !ValidRelPath(header.Name) {
			return fmt.Errorf("tar contained invalid name error %q", target)
		}

		// add dst + re-format slashes according to system
		target = filepath.Join(dst, header.Name)
		// if no join is needed, replace with ToSlash:
		// target = filepath.ToSlash(header.Name)

		// check the type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it (with 0755 permission)
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		// if it's a file create it (with same permission)
		case tar.TypeReg:
			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// copy over contents
			if _, err := io.Copy(fileToWrite, tr); err != nil {
				return err
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			fileToWrite.Close()
		}
	}

	return nil
}
*/
