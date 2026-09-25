package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// StartersDirName is the subdirectory of the templates path that holds
// organization starter bundles. TemplateLoader.LoadBundles does not recurse,
// so starters never take part in boot bundle selection.
const StartersDirName = "starters"

// StarterEntry is one starter file as seen by the organization-starters
// catalog. A starter that fails parsing, validation, or policy resolution is
// kept with Available=false and a reason instead of failing the whole list.
type StarterEntry struct {
	ID                string
	File              string
	Bundle            *TemplateBundle // nil when the file failed parse or validation
	Available         bool
	UnavailableReason string
}

// StartersPath returns the starters directory for a templates path.
func StartersPath(templatesPath string) string {
	return filepath.Join(templatesPath, StartersDirName)
}

// LoadStarterEntries reads starter bundles from startersPath. reservedIDs
// holds ids already in use (for example top-level bundles); a starter that
// reuses one is unavailable. policyGroups is the policy.yaml group-name set
// used to resolve delivery_defaults.approval_policy_ref; nil means the
// policy could not be loaded, so starters are marked unavailable.
//
// A missing starters directory yields no entries. Only an unreadable
// directory returns an error.
func LoadStarterEntries(startersPath string, reservedIDs, policyGroups map[string]struct{}) ([]*StarterEntry, error) {
	files, err := os.ReadDir(startersPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read starters dir: %w", err)
	}

	names := make([]string, 0, len(files))
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if f.IsDir() || (ext != ".yaml" && ext != ".yml") {
			continue
		}
		names = append(names, f.Name())
	}
	sort.Strings(names)

	seen := make(map[string]struct{}, len(names))
	entries := make([]*StarterEntry, 0, len(names))
	for _, name := range names {
		entry := loadStarterEntry(filepath.Join(startersPath, name), policyGroups)
		if entry.Available {
			if _, clash := reservedIDs[entry.ID]; clash {
				entry.markUnavailable(fmt.Errorf("id %q is already used by a top-level template bundle", entry.ID))
			} else if _, dup := seen[entry.ID]; dup {
				entry.markUnavailable(fmt.Errorf("duplicate starter id %q", entry.ID))
			}
		}
		seen[entry.ID] = struct{}{}
		entries = append(entries, entry)
	}
	return entries, nil
}

func loadStarterEntry(path string, policyGroups map[string]struct{}) *StarterEntry {
	file := filepath.Base(path)
	stem := strings.TrimSuffix(file, filepath.Ext(file))
	data, err := os.ReadFile(path)
	if err != nil {
		entry := &StarterEntry{ID: stem, File: file}
		entry.markUnavailable(fmt.Errorf("parse %s: %w", file, err))
		return entry
	}
	return starterEntryFromDocument(file, stem, data, policyGroups)
}

// starterEntryFromDocument applies the starter availability rules to one
// document. label names the document in reasons (a file name, or
// "operator template <id>"); defaultID is used when the bundle has no id.
func starterEntryFromDocument(label, defaultID string, data []byte, policyGroups map[string]struct{}) *StarterEntry {
	entry := &StarterEntry{ID: defaultID, File: label, Available: true}

	bundle, err := ParseTemplateBundle(data)
	if err != nil {
		entry.markUnavailable(fmt.Errorf("parse %s: %w", label, err))
		return entry
	}
	if id := strings.TrimSpace(bundle.ID); id != "" {
		entry.ID = id
	}
	if err := bundle.Validate(); err != nil {
		entry.markUnavailable(fmt.Errorf("validate %s: %w", label, err))
		return entry
	}
	// Bundle is set only once it validates, so projections never see a
	// half-normalized bundle.
	entry.Bundle = bundle
	if err := checkStarterBlocks(bundle, policyGroups); err != nil {
		entry.markUnavailable(fmt.Errorf("%s: %w", label, err))
	}
	return entry
}

// checkStarterBlocks enforces what a starter needs beyond bundle validation:
// a starter block, delivery_defaults, and a resolvable approval policy.
func checkStarterBlocks(bundle *TemplateBundle, policyGroups map[string]struct{}) error {
	if bundle.Starter == nil {
		return fmt.Errorf("starter bundles must declare a starter block")
	}
	if bundle.DeliveryDefaults == nil {
		return fmt.Errorf("starter bundles must declare delivery_defaults")
	}
	if err := bundle.DeliveryDefaults.ResolveApprovalPolicy(policyGroups); err != nil {
		return fmt.Errorf("delivery_defaults: %w", err)
	}
	return nil
}

func (e *StarterEntry) markUnavailable(err error) {
	e.Available = false
	e.UnavailableReason = err.Error()
}
