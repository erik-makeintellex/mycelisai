// MEM W2: blocker copy for the two memory edit-in-place codes
// (memory_entry_archived, memory_entry_changed). Kept out of
// blockerCopy.ts because that file sits at its 385-line cap;
// blockerCopy.ts imports and merges these into its own CODE_COPY map,
// so it stays the single lookup callers use.
import type { CopyTemplate } from './blockerCopy';

export const MEMORY_EDIT_CODE_COPY: Record<string, CopyTemplate> = {
    // The entry was archived (by anyone) since the editor opened; an
    // archived entry must be restored before it can be edited (mirrors
    // memory_entry_archived, deploymentcontext edit.go).
    memory_entry_archived: {
        title: 'Restore this before editing',
        whatHappened: 'This item was archived, so it cannot be changed yet. Nothing was saved.',
        nextAction: { label: 'OK', intent: 'dismiss' },
        whoCanHelp: 'Restore it first, then edit.',
    },
    // The entry moved (another edit, archive, or class/visibility change)
    // since this editor read it; reload before retrying (mirrors
    // memory_entry_changed, deploymentcontext edit.go).
    memory_entry_changed: {
        title: 'This item changed since you opened it',
        whatHappened: 'Someone else may have changed it. Reload to see the latest version before editing again.',
        nextAction: { label: 'Reload', intent: 'retry' },
    },
};
