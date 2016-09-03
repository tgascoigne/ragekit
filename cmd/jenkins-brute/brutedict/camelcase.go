package brutedict

import "strings"

type CamelCase struct {
	*WordDict
}

func lowerCamelCase(parts []string) string {
	return parts[0] + upperCamelCase(parts[1:])
}

func upperCamelCase(parts []string) string {
	newParts := make([]string, len(parts))
	for i, part := range parts {
		newParts[i] = strings.Title(part)
	}
	return strings.Join(newParts, "")
}

func (bc *CamelCase) Chan() chan string {
	result := make(chan string)
	go func() {
		for parts := range bc.CombinationsChan() {
			result <- lowerCamelCase(parts)
			result <- upperCamelCase(parts)
		}
	}()
	return result
}
