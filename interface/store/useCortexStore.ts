import { create } from 'zustand';
import type { CortexState } from '@/store/cortexStoreState';
import { createCortexAutomationRuntimeSlice } from '@/store/cortexStoreAutomationRuntimeSlice';
import { createCortexGovernanceSystemSlice } from '@/store/cortexStoreGovernanceSystemSlice';
import { createCortexGraphUiSlice } from '@/store/cortexStoreGraphUiSlice';
import { initialCortexState } from '@/store/cortexStoreInitialState';
import { createCortexMcpSlice } from '@/store/cortexStoreMcpSlice';
import { createCortexMissionChatSlice } from '@/store/cortexStoreMissionChatSlice';
import { createCortexMissionDraftSlice } from '@/store/cortexStoreMissionDraftSlice';
import { createCortexProposalExecutionSlice } from '@/store/cortexStoreProposalExecutionSlice';
import { createCortexResourceCatalogSlice } from '@/store/cortexStoreResourceCatalogSlice';
import { createCortexRuntimeSlice } from '@/store/cortexStoreRuntimeSlice';
import type { CortexGet, CortexSet } from '@/store/cortexStoreSliceTypes';
import { createCortexStreamSlice } from '@/store/cortexStoreStreamSlice';
import { createCortexUserSettingsSlice } from '@/store/cortexStoreUserSettingsSlice';
import { persistChat } from '@/store/cortexStoreUtils';
export * from '@/store/cortexStoreTypes';

function createCortexComposition(set: CortexSet, get: CortexGet) {
    return {
        ...initialCortexState,

        // Draft / graph + streaming shell
        ...createCortexGraphUiSlice(set, get),
        ...createCortexStreamSlice(set, get),

        // Resources + settings surfaces
        ...createCortexResourceCatalogSlice(set, get),
        ...createCortexMcpSlice(set, get),
        ...createCortexUserSettingsSlice(set, get),

        // Mission chat + governance contract
        ...createCortexMissionChatSlice(set, get),
        ...createCortexGovernanceSystemSlice(set, get),

        // Draft execution + proposal lifecycle
        ...createCortexMissionDraftSlice(set, get),
        ...createCortexProposalExecutionSlice(set, get),

        // Runtime automation and runs
        ...createCortexRuntimeSlice(set, get),
        ...createCortexAutomationRuntimeSlice(set, get),
    };
}

function shouldPersistMissionChat(state: CortexState, prevState: CortexState) {
    return (
        state.missionChat !== prevState.missionChat
        || state.workspaceChatScope !== prevState.workspaceChatScope
    );
}

export const useCortexStore = create<CortexState>((set, get) => createCortexComposition(set, get));

// ── Auto-sync missionChat → localStorage ─────────────────────
useCortexStore.subscribe((state, prevState) => {
    if (shouldPersistMissionChat(state, prevState)) {
        persistChat(state.missionChat, state.workspaceChatScope);
    }
});
