package stdin

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var ErrEmpty = errors.New("stdin is empty")

// Read reads input from an stdin pipe.
func Read() (string, error) {
	if IsEmpty() {
		return "", ErrEmpty
	}

	reader := bufio.NewReader(os.Stdin)
	var b strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		_, err = b.WriteRune(r)
		if err != nil {
			return "", fmt.Errorf("failed to write rune: %w", err)
		}
	}

	return b.String(), nil
}

// IsEmpty returns whether stdin is empty.
func IsEmpty() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return true
	}

	if stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0 {
		return true
	}

	return false
}
