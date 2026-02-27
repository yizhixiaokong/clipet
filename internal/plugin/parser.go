package plugin

import (
	"clipet/internal/game/capabilities"
	"fmt"
	"io/fs"
	"log"
	"path"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/x/ansi"
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

	// Apply defaults for backward compatibility
	pack.Lifecycle = pack.Lifecycle.Defaults()

	// Phase 6: Apply safety constraints (clamp if not explicitly overridden)
	constraints := capabilities.DefaultConstraints()
	if pack.Lifecycle.MaxAgeHours < constraints.MinLifespanHours &&
		pack.Lifecycle.EndingType != "eternal" {
		// Fallback to minimum safe value
		pack.Lifecycle.MaxAgeHours = constraints.MinLifespanHours
		log.Printf("[Plugin] Warning: species %q lifespan too short, clamped to %.0f hours",
			pack.Species.ID, pack.Lifecycle.MaxAgeHours)
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
// Directory layout (highest-priority first):
//
//  1. Multi-level subdirectory tree (recommended):
//     frames/{phase}/{variant}/.../animState.txt
//     Path components are joined with "_" to form the stageID.
//     Example: frames/adult/arcane_shadow/idle.txt  →  stageID="adult_arcane_shadow"
//              frames/egg/idle.txt                    →  stageID="egg"
//
//  2. Root sprite sheet: frames/{stageID}_{animState}.txt
//
//  3. Root legacy per-frame: frames/{stageID}_{animState}_{index}.txt
//
// Within each level, sprite sheets (multi-frame files separated by "---")
// take precedence over per-frame files for the same (stageID, animState) pair.
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

	// Pass 1: walk subdirectory tree — stageID derived from directory path (highest priority)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if err := walkFrameTree(fsys, framesDir, entry.Name(), frames); err != nil {
			return nil, err
		}
	}

	// Pass 2: root-level sprite sheets ({stageID}_{animState}.txt)
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
		key := FrameKey(stageID, animState)
		if _, exists := frames[key]; exists {
			continue // subdirectory takes precedence
		}

		data, err := fs.ReadFile(fsys, path.Join(framesDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read frame %s: %w", entry.Name(), err)
		}
		frames[key] = parseSpriteSheet(stageID, animState, string(data))
	}

	// Pass 3: root-level legacy per-frame ({stageID}_{animState}_{index}.txt)
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

	sort.Slice(legacy, func(i, j int) bool {
		if legacy[i].stageID != legacy[j].stageID {
			return legacy[i].stageID < legacy[j].stageID
		}
		if legacy[i].animState != legacy[j].animState {
			return legacy[i].animState < legacy[j].animState
		}
		return legacy[i].index < legacy[j].index
	})

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

// walkFrameTree recursively walks a directory tree under framesDir.
// relDir is the path relative to framesDir (e.g. "adult", "adult/shadow_mage").
//
// If a directory contains .txt files, they are loaded as frames for the
// stageID derived from relDir (with "/" replaced by "_").
// If a directory contains subdirectories, it recurses into them.
// Both can coexist — a directory may have both .txt files and subdirectories.
func walkFrameTree(fsys fs.FS, framesDir, relDir string, frames map[string]Frame) error {
	absDir := path.Join(framesDir, relDir)
	entries, err := fs.ReadDir(fsys, absDir)
	if err != nil {
		return nil
	}

	// Check if this directory has any .txt files (i.e. it's a leaf / frame dir)
	hasTxtFiles := false
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txt") {
			hasTxtFiles = true
			break
		}
	}

	if hasTxtFiles {
		// stageID = relDir with "/" → "_"
		stageID := strings.ReplaceAll(relDir, "/", "_")
		if err := loadStageFrames(fsys, absDir, stageID, frames); err != nil {
			return err
		}
	}

	// Recurse into subdirectories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if err := walkFrameTree(fsys, framesDir, relDir+"/"+entry.Name(), frames); err != nil {
			return err
		}
	}

	return nil
}

// loadStageFrames loads all frames for a single stageID from a leaf directory.
// Files named "{animState}.txt" (where animState is a known state) are sprite sheets (preferred).
// Files named "{animState}_{index}.txt" are legacy per-frame.
func loadStageFrames(fsys fs.FS, dir, stageID string, frames map[string]Frame) error {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil
	}

	// Sprite sheets: filename is an exact known anim state
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".txt")
		if !knownAnimStates[name] {
			continue
		}
		data, err := fs.ReadFile(fsys, path.Join(dir, entry.Name()))
		if err != nil {
			return fmt.Errorf("read frame %s/%s: %w", stageID, entry.Name(), err)
		}
		frames[FrameKey(stageID, name)] = parseSpriteSheet(stageID, name, string(data))
	}

	// Legacy per-frame: {animState}_{index}.txt
	type legacyFile struct {
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
		parts := strings.Split(name, "_")
		if len(parts) < 2 {
			continue
		}
		index := parts[len(parts)-1]
		if !isNumeric(index) {
			continue
		}
		animState := strings.Join(parts[:len(parts)-1], "_")
		key := FrameKey(stageID, animState)
		if _, exists := frames[key]; exists {
			continue // sprite sheet already loaded
		}
		data, err := fs.ReadFile(fsys, path.Join(dir, entry.Name()))
		if err != nil {
			return fmt.Errorf("read frame %s/%s: %w", stageID, entry.Name(), err)
		}
		legacy = append(legacy, legacyFile{animState: animState, index: index, content: string(data)})
	}

	sort.Slice(legacy, func(i, j int) bool {
		if legacy[i].animState != legacy[j].animState {
			return legacy[i].animState < legacy[j].animState
		}
		return legacy[i].index < legacy[j].index
	})

	for _, f := range legacy {
		key := FrameKey(stageID, f.animState)
		frame, ok := frames[key]
		if !ok {
			frame = Frame{StageID: stageID, AnimState: f.animState}
		}
		frame.Frames = append(frame.Frames, f.content)
		for _, line := range strings.Split(f.content, "\n") {
			if len(line) > frame.Width {
				frame.Width = len(line)
			}
		}
		frames[key] = frame
	}

	return nil
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

	// Calculate max display width across all frames
	for _, art := range frame.Frames {
		for _, line := range strings.Split(art, "\n") {
			w := ansi.StringWidth(line)
			if w > frame.Width {
				frame.Width = w
			}
		}
	}

	return frame
}

// splitSpriteSheetName parses a sprite sheet filename like "baby_idle"
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

// splitFrameName parses a frame filename like "baby_idle_0" into
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
