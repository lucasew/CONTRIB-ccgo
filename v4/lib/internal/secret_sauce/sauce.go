package sauce

import (
	"modernc.org/gc/v3"
)

func RemoveDeadVariables(filename string, buf []byte) (out []byte, err error) {
	_, err = gc.ParseFile(filename, buf)
	return buf, err //TODO
}
