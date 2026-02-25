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
//
// Two formats are supported:
//
//  1. Sprite sheet (preferred): {stageID}_{animState}.txt
//     Multiple frames in a single file, separated by a line containing only "---".
//
//  2. Legacy per-frame: {stageID}_{animState}_{index}.txt
//     Each file contains exactly one frame. Index determines ordering.
//
// If both formats exist for the same (stageID, animState) pair, the sprite
// sheet takes precedence.
//
// Returns a map keyed by FrameKey(stageID, animState).
func ParseFrames(fsys fs.FS, dir string) (map[string]Frame, error) {
	framesDir := path.Join(dir, "frames")
	frames := make(map[string]Frame)

	entries, err := fs.ReadDir(fsys, framesDir)
	if err != nil {
		// frames directory is optional
		return frames, nil
	}

	// Pass 1: load sprite sheets ({stageID}_{animState}.txt — exactly 2 parts after split)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".txt")
		parts := splitSpriteSheetName(name)
		if parts == nil {
			continue
		}
		stageID, animState := parts[0], parts[1]

		data, err := fs.ReadFile(fsys, path.Join(framesDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read frame %s: %w", entry.Name(), err)
		}

		frame := parseSpriteSheet(stageID, animState, string(data))
		frames[FrameKey(stageID, animState)] = frame
	}

	// Pass 2: load legacy per-frame files (only for keys not already loaded)
	type legacyFile struct {
		stageID   string
		animState string
		index     string
		content   string
	}

	var legacy []legacyFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".txt")
		parts := splitFrameName(name)
		if parts == nil {
			continue
		}
		// Skip if sprite sheet already covers this key
		key := FrameKey(parts[0], parts[1])
		if _, exists := frames[key]; exists {
			continue
		}

		data, err := fs.ReadFile(fsys, path.Join(framesDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read frame %s: %w", entry.Name(), err)
		}

		legacy = append(legacy, legacyFile{
			stageID:   parts[0],
			animState: parts[1],
			index:     parts[2],
			content:   string(data),
		})
	}

	// Sort legacy files by index to ensure correct frame order
	sort.Slice(legacy, func(i, j int) bool {
		if legacy[i].stageID != legacy[j].stageID {
			return legacy[i].stageID < legacy[j].stageID
		}
		if legacy[i].animState != legacy[j].animState {
			return legacy[i].animState < legacy[j].animState
		}
		return legacy[i].index < legacy[j].index
	})

	// Group legacy files into Frame objects
	for _, f := range legacy {
		key := FrameKey(f.stageID, f.animState)
		frame, ok := frames[key]
		if !ok {
			frame = Frame{
				StageID:   f.stageID,
				AnimState: f.animState,
			}
		}
		frame.Frames = append(frame.Frames, f.content)
		for _, line := range strings.Split(f.content, "\n") {
			if len(line) > frame.Width {
				frame.Width = len(line)
			}
		}
		frames[key] = frame
	}

	return frames, nil
}

// parseSpriteSheet splits a single file's content into multiple frames
// using "---" as the separator line, and builds a Frame struct.
func parseSpriteSheet(stageID, animState, content string) Frame {
	frame := Frame{
		StageID:   stageID,
		AnimState: animState,
	}

	// Split on lines that are exactly "---" (trimmed)
	var current []string
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "---" {
			if len(current) > 0 {
				art := strings.Join(current, "\n")
				frame.Frames = append(frame.Frames, art)
				current = nil
			}
			continue
		}
		current = append(current, line)
	}
	// Last frame (no trailing ---)
	if len(current) > 0 {
		// Trim trailing empty lines
		for len(current) > 0 && strings.TrimSpace(current[len(current)-1]) == "" {
			current = current[:len(current)-1]
		}
		if len(current) > 0 {
			art := strings.Join(current, "\n")
			frame.Frames = append(frame.Frames, art)
		}
	}

	// Calculate max width across all frames
	for _, art := range frame.Frames {
		for _, line := range strings.Split(art, "\n") {
			if len(line) > frame.Width {
				frame.Width = len(line)
			}
		}
	}

	return frame
}

// splitSpriteSheetName parses a sprite sheet filename like "baby_cat_idle"
// into [stageID, animState]. The animState must be a known animation state.
// Returns nil if the name doesn't match (e.g. has an index suffix → legacy format).
func splitSpriteSheetName(name string) []string {
	parts := strings.Split(name, "_")
	if len(parts) < 2 {
		return nil
	}

	// The last segment is the animState candidate.
	// For sprite sheets the last part must be a known anim state name,
	// NOT a numeric index (which would indicate legacy per-frame format).
	animState := parts[len(parts)-1]
	if !knownAnimStates[animState] {
		return nil
	}
	// If it looks like a number, it's a legacy index, not an anim state
	if isNumeric(animState) {
		return nil
	}
	stageID := strings.Join(parts[:len(parts)-1], "_")
	if stageID == "" {
		return nil
	}
	return []string{stageID, animState}
}

// knownAnimStates is the set of recognized animation state names.
var knownAnimStates = map[string]bool{
	"idle":     true,
	"eating":   true,
	"sleeping": true,
	"playing":  true,
	"happy":    true,
	"sad":      true,
}

// isNumeric returns true if s consists entirely of digits.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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
