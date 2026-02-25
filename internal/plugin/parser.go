package plugin

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

// ParseSpecies reads and decodes species.toml from the given filesystem.
func ParseSpecies(fsys fs.FS, dir string) (*SpeciesPack, error) {
	filePath := path.Join(dir, "species.toml")
	data, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		return nil, fmt.Errorf("read species.toml: %w", err)
	}

	var pack SpeciesPack
	if err := toml.Unmarshal(data, &pack); err != nil {
		return nil, fmt.Errorf("parse species.toml: %w", err)
	}

	return &pack, nil
}

// ParseDialogues reads and decodes dialogues.toml from the given filesystem.
// Returns nil (no error) if the file does not exist.
func ParseDialogues(fsys fs.FS, dir string) ([]DialogueGroup, error) {
	filePath := path.Join(dir, "dialogues.toml")
	data, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		// dialogues.toml is optional
		return nil, nil
	}

	var df DialoguesFile
	if err := toml.Unmarshal(data, &df); err != nil {
		return nil, fmt.Errorf("parse dialogues.toml: %w", err)
	}

	return df.Dialogues, nil
}

// ParseAdventures reads and decodes adventures.toml from the given filesystem.
// Returns nil (no error) if the file does not exist.
func ParseAdventures(fsys fs.FS, dir string) ([]Adventure, error) {
	filePath := path.Join(dir, "adventures.toml")
	data, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		// adventures.toml is optional
		return nil, nil
	}

	var af AdventuresFile
	if err := toml.Unmarshal(data, &af); err != nil {
		return nil, fmt.Errorf("parse adventures.toml: %w", err)
	}

	return af.Adventures, nil
}

// ParseFrames scans the frames/ subdirectory for ASCII art frame files.
// Frame files must follow the naming convention: {stageID}_{animState}_{index}.txt
// Returns a map keyed by FrameKey(stageID, animState).
func ParseFrames(fsys fs.FS, dir string) (map[string]Frame, error) {
	framesDir := path.Join(dir, "frames")
	frames := make(map[string]Frame)

	entries, err := fs.ReadDir(fsys, framesDir)
	if err != nil {
		// frames directory is optional
		return frames, nil
	}

	// Collect all .txt files
	type frameFile struct {
		stageID   string
		animState string
		index     string
		content   string
	}

	var files []frameFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".txt")
		parts := splitFrameName(name)
		if parts == nil {
			continue // skip files that don't match the naming convention
		}

		data, err := fs.ReadFile(fsys, path.Join(framesDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read frame %s: %w", entry.Name(), err)
		}

		files = append(files, frameFile{
			stageID:   parts[0],
			animState: parts[1],
			index:     parts[2],
			content:   string(data),
		})
	}

	// Sort by index to ensure correct frame order
	sort.Slice(files, func(i, j int) bool {
		if files[i].stageID != files[j].stageID {
			return files[i].stageID < files[j].stageID
		}
		if files[i].animState != files[j].animState {
			return files[i].animState < files[j].animState
		}
		return files[i].index < files[j].index
	})

	// Group into Frame objects
	for _, f := range files {
		key := FrameKey(f.stageID, f.animState)
		frame, ok := frames[key]
		if !ok {
			frame = Frame{
				StageID:   f.stageID,
				AnimState: f.animState,
			}
		}
		frame.Frames = append(frame.Frames, f.content)
		// Auto-calculate width from frame content
		for _, line := range strings.Split(f.content, "\n") {
			if len(line) > frame.Width {
				frame.Width = len(line)
			}
		}
		frames[key] = frame
	}

	return frames, nil
}

// splitFrameName parses a frame filename like "baby_cat_idle_0" into
// [stageID, animState, index]. The stage ID can contain underscores,
// so we split from the right: last part = index, second-to-last = animState,
// everything else = stageID.
func splitFrameName(name string) []string {
	parts := strings.Split(name, "_")
	if len(parts) < 3 {
		return nil
	}

	index := parts[len(parts)-1]
	animState := parts[len(parts)-2]
	stageID := strings.Join(parts[:len(parts)-2], "_")

	if stageID == "" || animState == "" || index == "" {
		return nil
	}

	return []string{stageID, animState, index}
}

// ParsePack loads a complete species pack from a directory in the given filesystem.
func ParsePack(fsys fs.FS, dir string) (*SpeciesPack, error) {
	pack, err := ParseSpecies(fsys, dir)
	if err != nil {
		return nil, err
	}

	dialogues, err := ParseDialogues(fsys, dir)
	if err != nil {
		return nil, err
	}
	pack.Dialogues = dialogues

	adventures, err := ParseAdventures(fsys, dir)
	if err != nil {
		return nil, err
	}
	pack.Adventures = adventures

	frames, err := ParseFrames(fsys, dir)
	if err != nil {
		return nil, err
	}
	pack.Frames = frames

	return pack, nil
}
