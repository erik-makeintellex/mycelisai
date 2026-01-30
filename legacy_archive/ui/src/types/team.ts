export interface Team {
    id: string;
    name: string;
    description: string;
    agents: string[];
    channels: string[];
    inter_comm_channel?: string;
    resource_access: Record<string, string>;
    created_at: string;
}
