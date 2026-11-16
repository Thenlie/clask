package taskfile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// EncodeTask renders a task's metadata as the YAML that sits between its
// --- lines. Keys follow Task field order; empty optional keys are omitted.
func EncodeTask(t *Task) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(t); err != nil {
		return nil, fmt.Errorf("encode task %q: %w", t.ID, err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("encode task %q: %w", t.ID, err)
	}
	return buf.Bytes(), nil
}

// Replace returns the file contents with t's YAML range rewritten from t's
// current field values. Every other byte of the file is unchanged.
func (f *File) Replace(t *Task) ([]byte, error) {
	y, err := EncodeTask(t)
	if err != nil {
		return nil, err
	}
	sp := t.Span
	if sp.YAMLStart < 0 || sp.YAMLEnd > len(f.Data) || sp.YAMLStart > sp.YAMLEnd {
		return nil, fmt.Errorf("task %q has an invalid span %+v", t.ID, sp)
	}
	out := make([]byte, 0, len(f.Data)-(sp.YAMLEnd-sp.YAMLStart)+len(y))
	out = append(out, f.Data[:sp.YAMLStart]...)
	out = append(out, y...)
	out = append(out, f.Data[sp.YAMLEnd:]...)
	return out, nil
}

// Append returns the file contents with a new task block (and optional
// body) added at the end, separated from the previous content by exactly
// one blank line.
func (f *File) Append(t *Task, body string) ([]byte, error) {
	y, err := EncodeTask(t)
	if err != nil {
		return nil, err
	}
	out := bytes.TrimRight(f.Data, " \t\r\n")
	out = append(out, "\n\n---\n"...)
	out = append(out, y...)
	out = append(out, "---\n"...)
	if body = trimBody([]byte(body)); body != "" {
		out = append(out, body...)
		out = append(out, '\n')
	}
	return out, nil
}

// WriteFile replaces path atomically: data goes to a temp file in the same
// directory which is then renamed over the original. The original's
// permissions are kept when it exists.
func WriteFile(path string, data []byte) error {
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	tmpName := tmp.Name()
	cleanup := func(err error) error {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write %s: %w", path, err)
	}
	if _, err := tmp.Write(data); err != nil {
		return cleanup(err)
	}
	if err := tmp.Chmod(mode); err != nil {
		return cleanup(err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
