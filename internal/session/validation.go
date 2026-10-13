package session

import (
	"fmt"
	"math"
	"path/filepath"
	"slices"
	"strings"
)

func validateTags(tags []string) error {
	if len(tags) > 20 {
		return fmt.Errorf("too many tags")
	}
	for _, tag := range tags {
		if strings.TrimSpace(tag) != tag || tag == "" || len(tag) > 80 || strings.ContainsAny(tag, "\n\r\x00,") {
			return fmt.Errorf("invalid tag")
		}
	}
	return nil
}

// Validate the whole batch before any engine call or persistence side effect.
func (c Command) validate(records []Record) error {
	switch c.Kind {
	case CommandPauseAll, CommandResumeBatch:
		if len(c.IDs) != 0 {
			return fmt.Errorf("batch commands do not take task IDs")
		}
		return nil
	case CommandPause, CommandResume, CommandVerify, CommandPriority, CommandRemove, CommandSeed, CommandFirst, CommandNotified, CommandRelocate, CommandTags:
	default:
		return fmt.Errorf("unknown command %q", c.Kind)
	}
	if len(c.IDs) == 0 {
		return fmt.Errorf("select at least one task")
	}
	for _, id := range c.IDs {
		index := slices.IndexFunc(records, func(r Record) bool { return r.ID == id })
		if index < 0 {
			return fmt.Errorf("unknown task %q", id)
		}
		r := records[index]
		if r.Removing && c.Kind != CommandRemove {
			return fmt.Errorf("task is being removed")
		}
		if c.Kind == CommandPriority {
			indices := c.FileIndices
			if len(indices) == 0 {
				indices = []int{c.FileIndex}
			}
			if c.Priority < 0 || c.Priority > 2 {
				return fmt.Errorf("invalid priority")
			}
			for _, i := range indices {
				if i < 0 || i >= len(r.Source.Files) {
					return fmt.Errorf("invalid file")
				}
			}
		}
		if c.Kind == CommandRelocate && (len(c.IDs) != 1 || r.Metadata == "" || !filepath.IsAbs(c.Destination)) {
			return fmt.Errorf("relocation requires one task with metadata and an absolute destination")
		}
	}
	if c.Kind == CommandSeed && (math.IsNaN(c.Ratio) || math.IsInf(c.Ratio, 0) || c.Ratio < -1 || c.SeedLimitSeconds < 0) {
		return fmt.Errorf("invalid seeding policy")
	}
	if c.Kind == CommandTags {
		return validateTags(c.Tags)
	}
	return nil
}
