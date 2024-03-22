package utils

import (
	"github.com/jinzhu/copier"
)

// CopyStruct copies the fields of the source struct to the destination struct.
func CopyStruct(src interface{}, dst interface{}) error {
	return copier.Copy(dst, src)
}
