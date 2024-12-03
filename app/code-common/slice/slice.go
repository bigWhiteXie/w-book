package slice

import "strconv"

func StringsToInt64s(strs []string) ([]int64, error) {
	ints := make([]int64, len(strs))
	for i, s := range strs {
		num, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err // return error if conversion fails
		}
		ints[i] = num
	}
	return ints, nil
}
