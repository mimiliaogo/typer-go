// Demo code for the Flex primitive.
package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	s := " Apple two space  th[r[]ee space   bla()123(123) "
	re := regexp.MustCompile("[[:^ascii:]]")
	text := re.ReplaceAllLiteralString(s, "")
	re = regexp.MustCompile("[ ]{2,}")
	text = re.ReplaceAllLiteralString(text, " ")
	re = regexp.MustCompile("[()0-9]")
	text = re.ReplaceAllLiteralString(text, "")
	text = strings.Replace(text, "\n", " ", -1)
	text = strings.Replace(text, "[", "", -1)
	text = strings.Replace(text, "]", "", -1)
	text = strings.TrimSpace(text)

	fmt.Println(text)
}
