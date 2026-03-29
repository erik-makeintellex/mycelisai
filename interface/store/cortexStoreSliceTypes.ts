import type { CortexState } from '@/store/cortexStoreState';

export type CortexSet = (
    partial:
        | Partial<CortexState>
        | ((state: CortexState) => Partial<CortexState>),
    replace?: false,
) => void;

export type CortexGet = () => CortexState;
