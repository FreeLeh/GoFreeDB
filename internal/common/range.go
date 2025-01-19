package common

func GetA1Range(sheetName string, rng string) string {
	return sheetName + "!" + rng
}

type ColIdx struct {
	Name string
	Idx  int
}

type ColsMapping map[string]ColIdx

func (m ColsMapping) NameMap() map[string]string {
	result := make(map[string]string, 0)
	for col, val := range m {
		result[col] = val.Name
	}
	return result
}

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateColumnMapping(columns []string) map[string]ColIdx {
	mapping := make(map[string]ColIdx, len(columns))
	for n, col := range columns {
		mapping[col] = ColIdx{
			Name: GenerateColumnName(n),
			Idx:  n,
		}
	}
	return mapping
}

func GenerateColumnName(n int) string {
	// This is not purely a Base26 conversion since the second char can start from "A" (or 0) again.
	// In a normal Base26 int to string conversion, the second char can only start from "B" (or 1).
	// Hence, we need to hack it by checking the first round separately from the subsequent round.
	// For the subsequent rounds, we need to subtract by 1 first or else it will always start from 1 (not 0).
	col := string(alphabet[n%26])
	n = n / 26

	for {
		if n <= 0 {
			break
		}

		n -= 1
		col = string(alphabet[n%26]) + col
		n = n / 26
	}

	return col
}
