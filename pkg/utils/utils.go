package utils

import (
	"errors"
	"strconv"
)

func ConvertCIDToInt(cid string) (int, error) {
	// Test cid
	if len(cid) < 6 {
		return 0, errors.New("malformed cid, " + cid)
	}
	// Remove 2 initial chars 0x
	res, err := strconv.ParseInt(cid[2:], 16, 32)

	if err != nil {
		return int(res), err
	}

	return int(res), nil
}
