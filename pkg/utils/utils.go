package utils

import (
	"strconv"
	"strings"
)

//Convert Int32 to String
func Int32ToString(num int32) string {
	return strconv.FormatInt(int64(num), 10)
}

func RemoveSHATags(image string) string {
	if strings.Contains(image,"@") {
		return image[0:strings.Index(image, "@")]
	}

	return image
}

