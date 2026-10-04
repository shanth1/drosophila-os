package malecns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/shanth1/drosophila-os/internal/engine"
)

func writeConnectome(path string, conn *engine.Connectome) (err error) {
	f, err := os.CreateTemp(filepath.Dir(path), ".malecns-importer-*")
	if err != nil {
		return fmt.Errorf("create temporary connectome: %w", err)
	}
	tempPath := f.Name()
	defer func() {
		if removeErr := os.Remove(tempPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			err = errors.Join(err, fmt.Errorf("remove temporary connectome: %w", removeErr))
		}
	}()
	writeErr := encodeConnectome(f, conn)
	if err := errors.Join(writeErr, f.Close()); err != nil {
		return fmt.Errorf("write temporary connectome: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace connectome %q: %w", path, err)
	}
	return nil
}

func encodeConnectome(output io.Writer, conn *engine.Connectome) error {
	for _, data := range []any{
		[4]byte{'D', 'R', 'O', 'S'}, uint32(1), conn.NumNeurons, conn.NumEdges,
		conn.Offsets, conn.Targets, conn.Weights,
	} {
		if err := binary.Write(output, binary.LittleEndian, data); err != nil {
			return fmt.Errorf("encode connectome: %w", err)
		}
	}
	return nil
}
