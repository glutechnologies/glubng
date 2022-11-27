package core

import "strconv"

func ConvertCIDToInt(cid string) (int, error) {
	// Remove 2 initial chars 0x
	res, err := strconv.ParseInt(cid[2:], 16, 32)

	if err != nil {
		return int(res), err
	}

	return int(res), nil
}
