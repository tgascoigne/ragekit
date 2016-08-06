package crypto

func CalculateHash(keys Keys, filename string) (result uint32) {
	for _, c := range filename {
		temp := 1025 * (uint32(keys.hashLookup[int(c)]) + result)
		result = (temp >> 6) ^ temp
	}
	return 32769 * (((9 * result) >> 11) ^ 9*result)
}
