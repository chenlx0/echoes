package version

import (
	"fmt"
	"runtime"
)

// Binary version
const Binary = "0.0.1-alpha"

// String return formated version information
func String() string {
	return fmt.Sprintf("Echoes Reverse Proxy Server v%s (built w/%s)", Binary, runtime.Version())
}
