// Maps a backend blocker `code` (plus viewer role and optional reason) to
// plain-language copy for OperationalAlert. This is the single source of
// truth for blocker vocabulary: no caller should hand-build this copy.
// See scratchpad/ux/complex-actions-copy-deck.md (vocabulary table, section 2)
// and scratchpad/ux/U1-spec.md (section 0, the contract U1 consumes).

export interface BlockerInput {
    code: string;
    httpStatus?: number;
    viewerIsAdmin: boolean;
    reason?: string;
}

export interface BlockerAction {
    label: string;
    href?: string;
    intent?: 'retry' | 'dismiss' | 'back' | 'open';
}

export interface BlockerCopy {
    title: string;
    whatHappened: string;
    nextAction: BlockerAction;
    whoCanHelp?: string;
    adminVariant?: Omit<BlockerCopy, 'adminVariant'>;
    code: string;
}

type CopyTemplate = Omit<BlockerCopy, 'code' | 'adminVariant'> & { adminVariant?: Omit<BlockerCopy, 'adminVariant' | 'code'> };

const REQUEST_FAILED: CopyTemplate = {
    title: 'Something went wrong',
    whatHappened: 'The request did not go through. Nothing changed.',
    nextAction: { label: 'Try again', intent: 'retry' },
    whoCanHelp: 'An admin can help if this keeps happening.',
    adminVariant: {
        title: 'Something went wrong',
        whatHappened: 'The request did not go through. Nothing changed.',
        nextAction: { label: 'Try again', intent: 'retry' },
    },
};

const ADMIN_REQUIRED_BY_REASON: Record<string, CopyTemplate> = {
    approvals: {
        title: 'Approvals are handled by admins',
        whatHappened: "When something you ask for needs an admin's OK, you'll see it in your Soma conversation.",
        nextAction: { label: 'Back to Soma', href: '/dashboard', intent: 'back' },
        whoCanHelp: 'Who can help: an admin.',
    },
    settings: {
        title: 'This setting is managed by admins',
        whatHappened: 'Your admin controls sign-in, AI engines, people and automation profiles. You can manage your own profile here.',
        nextAction: { label: 'Edit my profile', href: '/settings?tab=profile', intent: 'open' },
        whoCanHelp: 'Who can help: an admin.',
    },
    groups: {
        title: 'Groups are managed by admins',
        whatHappened: "You don't have access to groups here. Your own work is in Work.",
        nextAction: { label: 'Open Work', href: '/teams', intent: 'open' },
        whoCanHelp: 'Ask an admin to add you to a group.',
    },
    automation_profile: {
        title: 'Automation profiles are managed by admins',
        whatHappened: 'Ask an admin if an automation needs a different AI engine.',
        nextAction: { label: 'Back to my profile', href: '/settings?tab=profile', intent: 'back' },
        whoCanHelp: 'Who can help: an admin.',
    },
    default: {
        title: 'This area is for admins',
        whatHappened: 'It holds settings and information that only admins manage.',
        nextAction: { label: 'Back to Soma', href: '/dashboard', intent: 'back' },
        whoCanHelp: 'Who can help: an admin.',
    },
};

const CODE_COPY: Record<string, CopyTemplate> = {
    approver_required: {
        title: 'Waiting for an admin to approve',
        whatHappened: 'Nothing has run. This proposal stays open until an admin approves it.',
        nextAction: { label: 'OK, keep it open', intent: 'dismiss' },
        whoCanHelp: 'An admin can approve this.',
    },
    confirmer_not_proposer: {
        title: 'Only the requester can approve this',
        whatHappened: 'The person who asked for this can approve it, or an admin can. Nothing ran.',
        nextAction: { label: 'Ask Soma for my own version', intent: 'open' },
        whoCanHelp: 'An admin can also approve this.',
    },
    token_already_used: {
        title: 'Already approved',
        whatHappened: "This proposal was approved a moment ago, maybe in another tab. It won't run twice.",
        nextAction: { label: 'See the result', intent: 'open' },
    },
    token_purpose_unknown: {
        title: 'This proposal is out of date',
        whatHappened: 'It was made before a safety update. Soma needs to propose it again. Nothing ran.',
        nextAction: { label: 'Ask Soma to propose again', intent: 'retry' },
    },
    token_wrong_purpose: {
        title: "This can't be approved here",
        whatHappened: 'This kind of proposal is approved from a different screen. Nothing ran.',
        nextAction: { label: 'Ask Soma to propose again', intent: 'retry' },
    },
    outcome_template_unresolved: {
        title: 'Start this from Soma',
        whatHappened: 'This outcome template belongs to an organization, so start it with Soma in that organization.',
        nextAction: { label: 'Open Soma', href: '/dashboard', intent: 'open' },
    },
    blueprint_mismatch: {
        title: 'You changed the plan',
        whatHappened: 'Edits made after Soma proposed the plan need a new proposal. Nothing launched.',
        nextAction: { label: 'Ask Soma to update the plan', intent: 'retry' },
    },
    confirm_required: {
        title: 'Team plan needs a fresh approval',
        whatHappened: 'Soma has to propose this plan before it can launch. Nothing launched.',
        nextAction: { label: 'Ask Soma to propose the plan', intent: 'retry' },
    },
    // The backend's real code is `governance_policy_unavailable`
    // (governance_authority.go); `policy_unavailable` is kept below only as
    // an alias for any caller still passing the old name.
    governance_policy_unavailable: {
        title: 'Safety lock is on',
        whatHappened: "Mycelis couldn't load its safety rules, so most actions are paused until an admin fixes them. Your saved work and results are safe.",
        nextAction: { label: 'OK', intent: 'dismiss' },
        whoCanHelp: 'An admin can fix this.',
        adminVariant: {
            title: "Safety rules didn't load",
            whatHappened: 'Everything except health checks is paused. Save a valid set of rules to unlock it. No restart is needed.',
            nextAction: { label: 'Open safety rules', href: '/automations?tab=approvals', intent: 'open' },
        },
    },
    invalid_confirm_token: {
        title: 'This proposal is no longer valid',
        whatHappened: 'Ask Soma to propose it again. Nothing ran.',
        nextAction: { label: 'Ask Soma to propose again', intent: 'retry' },
    },
    provider_bound: {
        title: 'This AI engine is in use',
        whatHappened: 'This engine handles active task types. Choose another engine for those tasks first, then turn this one off.',
        nextAction: { label: 'Choose engines for these tasks', href: '/settings?tab=engines', intent: 'open' },
    },
    model_missing: {
        title: "Soma can't answer right now",
        whatHappened: "The AI engine Soma uses isn't fully set up yet. An admin needs to finish it.",
        nextAction: { label: 'Try again', intent: 'retry' },
        whoCanHelp: 'An admin can help with this.',
        adminVariant: {
            title: "Soma's AI engine has no model",
            whatHappened: 'The engine chosen for this task type has no model selected.',
            nextAction: { label: 'Open AI Engines', href: '/settings?tab=engines', intent: 'open' },
        },
    },
    service_unavailable: {
        title: 'Changes are paused',
        whatHappened: 'Changes are paused because the activity log is unavailable.',
        nextAction: { label: 'Try again', intent: 'retry' },
        whoCanHelp: 'An admin can restore the activity log.',
        adminVariant: {
            title: 'Activity log is unavailable',
            whatHappened: 'Changes are paused because the activity log is unavailable. Restore it, then retry.',
            nextAction: { label: 'Try again', intent: 'retry' },
        },
    },
    transport_unavailable: {
        title: "Soma's team service isn't running",
        whatHappened: "Soma can't reach the team service right now. Nothing was lost.",
        nextAction: { label: 'Try again', intent: 'retry' },
        whoCanHelp: 'An admin can check System Status.',
        adminVariant: {
            title: "Soma's team service isn't running",
            whatHappened: "The team service is unreachable. Check System Status and restart it if needed.",
            nextAction: { label: 'Open System Status', href: '/system', intent: 'open' },
        },
    },
    provider_timeout: {
        title: "Soma couldn't draft that file",
        whatHappened: 'The AI engine took too long to respond. Nothing was written. Try again, or add the content yourself.',
        nextAction: { label: 'Try again', intent: 'retry' },
    },
    empty_provider_output: {
        title: "Soma couldn't draft that file",
        whatHappened: 'The AI engine returned nothing usable. Nothing was written. Try again, or add the content yourself.',
        nextAction: { label: 'Try again', intent: 'retry' },
    },
    connector_deployment_unavailable: {
        title: "Connector installs aren't available yet",
        whatHappened: 'Nothing was installed. Connector deployment has not shipped yet, so nothing was recorded.',
        nextAction: { label: 'Use MCP servers or providers instead', href: '/settings?tab=tools', intent: 'open' },
        whoCanHelp: 'An admin can check System Status for updates.',
    },
    request_failed: REQUEST_FAILED,
};

function toBlockerCopy(code: string, template: CopyTemplate): BlockerCopy {
    return {
        title: template.title,
        whatHappened: template.whatHappened,
        nextAction: template.nextAction,
        whoCanHelp: template.whoCanHelp,
        adminVariant: template.adminVariant ? { code, ...template.adminVariant } : undefined,
        code,
    };
}

/**
 * Maps {code, httpStatus, viewerIsAdmin, reason} to plain-language blocker
 * copy. Never returns raw backend text: unknown codes fall back to the
 * generic `request_failed` copy while still echoing the original code for
 * "Details".
 *
 * Per the U1 contract, this always returns the non-admin copy plus an
 * `adminVariant` when one exists; it does not resolve `viewerIsAdmin`
 * itself (callers such as OperationalAlert do that). Use
 * `resolveBlockerCopy` below when a caller needs the resolved shape
 * directly.
 */
// Aliases for codes this map used to key on before the backend's real name
// was confirmed. Kept so any caller still passing the old name still works.
const CODE_ALIASES: Record<string, string> = {
    policy_unavailable: 'governance_policy_unavailable',
};

// PH-B: the backend degradation code for an unverified saved output varies
// by what could not be confirmed (`output_readback_missing`,
// `output_readback_mismatch`, `output_missing`, ...). Any code with this
// prefix that has no exact CODE_COPY entry gets this shared copy.
const OUTPUT_READBACK_FALLBACK: CopyTemplate = {
    title: "Couldn't confirm the file",
    whatHappened: "Soma finished the work, but couldn't confirm the file it wrote. Nothing was lost.",
    nextAction: { label: 'Open the file', intent: 'open' },
    whoCanHelp: 'You can also ask Soma to redo it.',
};

export function blockerCopy(input: BlockerInput): BlockerCopy {
    const code = CODE_ALIASES[input.code] ?? input.code;
    const { reason } = input;

    if (code === 'admin_required') {
        const template = ADMIN_REQUIRED_BY_REASON[reason ?? 'default'] ?? ADMIN_REQUIRED_BY_REASON.default;
        return toBlockerCopy(code, template);
    }

    const template = CODE_COPY[code];
    if (template) {
        return toBlockerCopy(code, template);
    }

    if (code.startsWith('output_')) {
        return toBlockerCopy(code, OUTPUT_READBACK_FALLBACK);
    }

    return toBlockerCopy(code, REQUEST_FAILED);
}

/** A BlockerCopy resolved for one viewer: the admin variant when both
 * `viewerIsAdmin` and an `adminVariant` are present, otherwise the base copy. */
export type ResolvedBlockerCopy = Omit<BlockerCopy, 'adminVariant'>;

export function resolveBlockerCopy(copy: BlockerCopy, viewerIsAdmin: boolean): ResolvedBlockerCopy {
    if (viewerIsAdmin && copy.adminVariant) {
        return { ...copy.adminVariant, code: copy.code };
    }
    const { adminVariant: _adminVariant, ...rest } = copy;
    return rest;
}
