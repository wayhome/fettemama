package shared

import (
  "strings"
)

//strip evil console command codes out ...
func Telstrip(s string) string {
    return strings.Trim(s, " \r\n")
	//ts := make([]int, 0, len(s)) //our return slice

	//for _, c := range s {
	//	if c < ' ' {
	//		continue
	//	}
	//	ts = append(ts, c)
	//}

	//return string(ts)
}
