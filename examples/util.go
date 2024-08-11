package examples

import (
	"encoding/json"
	"fmt"
	"io"
)

func PrettyPrint(w io.Writer, v any) {
	prettyValue, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	fmt.Fprintf(w, "%s", prettyValue)
}
