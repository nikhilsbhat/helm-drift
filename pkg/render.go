package pkg

import (
	"bufio"
	"fmt"
)

func (drift *Drift) render(drifts string) {
	_, err := drift.writer.Write([]byte(drifts))
	if err != nil {
		drift.log.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			drift.log.Fatalln(err)
		}
	}(drift.writer)
}

func addNewLine(message string) string {
	return fmt.Sprintf("%s\n", message)
}
