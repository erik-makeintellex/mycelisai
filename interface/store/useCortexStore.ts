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
import { createCortexStreamSlice } from '@/store/cortexStoreStreamSlice';
import { createCortexUserSettingsSlice } from '@/store/cortexStoreUserSettingsSlice';
import { persistChat } from '@/store/cortexStoreUtils';
export * from '@/store/cortexStoreTypes';

export const useCortexStore = create<CortexState>((set, get) => ({
    ...initialCortexState,
    ...createCortexGraphUiSlice(set, get),
    ...createCortexStreamSlice(set, get),

    ...createCortexResourceCatalogSlice(set, get),
    ...createCortexMcpSlice(set, get),
    ...createCortexUserSettingsSlice(set, get),

    ...createCortexMissionChatSlice(set, get),

    ...createCortexGovernanceSystemSlice(set, get),

    ...createCortexMissionDraftSlice(set, get),
    ...createCortexProposalExecutionSlice(set, get),

    ...createCortexRuntimeSlice(set, get),
    ...createCortexAutomationRuntimeSlice(set, get),
}));

// ── Auto-sync missionChat → localStorage ─────────────────────
useCortexStore.subscribe((state, prevState) => {
    if (
        state.missionChat !== prevState.missionChat
        || state.workspaceChatScope !== prevState.workspaceChatScope
    ) {
        persistChat(state.missionChat, state.workspaceChatScope);
    }
});
