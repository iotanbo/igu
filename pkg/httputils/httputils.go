package httputils

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/iotanbo/igu/pkg/ecdef"
	//lint:ignore ST1001 - for concise error handling.
	. "github.com/iotanbo/igu/pkg/errs"
)

// Download can download large files without risk
// depleting memory.
// Based on https://stackoverflow.com/a/33853856/3824328
func Download(url string, destPath string) Err {
	// Create the file
	out, err := os.Create(destPath)
	if err != nil {
		return FromError(err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return FromError(err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return Err{
			Code: ecdef.ErrCode(resp.StatusCode),
			Msg:  fmt.Sprintf("bad status: %s", resp.Status),
		}
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return FromError(err)
	}

	return NoError
}
