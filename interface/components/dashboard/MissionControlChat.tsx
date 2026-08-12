"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Brain, Megaphone } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import { SomaConversationThread } from "@/components/soma/SomaConversationThread";
import { SomaIntentInput } from "@/components/soma/SomaIntentInput";
import { useSomaOutputContinuation } from "@/components/soma/outputContinuation";
import type { MissionChatContinuationContext } from "@/store/cortexStoreTypes";
import { DEFAULT_SOMA_SUGGESTIONS, type SomaSuggestion } from "@/components/soma/SomaSuggestionBar";
import CouncilCallErrorCard from "./CouncilCallErrorCard";
import { BroadcastModeIndicator, MissionControlEmptyState, SomaOfflineGuide } from "./MissionControlChatStates";
import { MissionControlChatHeader, SomaActivityIndicator } from "./MissionControlChatChrome";
import MissionControlMessageBubble from "./MissionControlMessageBubble";
import { MissionControlAdvancedInput } from "./MissionControlAdvancedInput";
import { MissionControlContinuationChip } from "./MissionControlContinuationChip";
import MissionControlTeamContinuationPrompt from "./MissionControlTeamContinuationPrompt";
import OrchestrationInspector from "./OrchestrationInspector";
import { somaPlaceholder, teamSuggestions } from "./missionControlChatUi";
import { buildMissionChatScope } from "@/store/cortexStoreMissionChatHelpers";
import { clearAllPersistedChat } from "@/store/cortexStoreUtils";
import { conversationalProposalReply } from "./conversationalProposalReply";
import { presentMissionChat } from "./missionControlChatPresentation";

export default function MissionControlChat({
    simpleMode = false,
    autoFocus = false,
    organizationId,
    focusedTeamId,
    suggestions = DEFAULT_SOMA_SUGGESTIONS,
}: {
    simpleMode?: boolean;
    autoFocus?: boolean;
    organizationId?: string;
    focusedTeamId?: string | null;
    suggestions?: readonly SomaSuggestion[];
}) {
    const missionChat = useCortexStore((s) => s.missionChat);
    const isMissionChatting = useCortexStore((s) => s.isMissionChatting);
    const missionChatFailure = useCortexStore((s) => s.missionChatFailure);
    const sendMissionChat = useCortexStore((s) => s.sendMissionChat);
    const pendingProposal = useCortexStore((s) => s.pendingProposal);
    const confirmProposal = useCortexStore((s) => s.confirmProposal);
    const cancelProposal = useCortexStore((s) => s.cancelProposal);
    const clearMissionChat = useCortexStore((s) => s.clearMissionChat);
    const setMissionChatScope = useCortexStore((s) => s.setMissionChatScope);
    const broadcastToSwarm = useCortexStore((s) => s.broadcastToSwarm);
    const isBroadcasting = useCortexStore((s) => s.isBroadcasting);
    const assistantName = useCortexStore((s) => s.assistantName);
    const presentedMissionChat = useMemo(() => presentMissionChat(missionChat), [missionChat]);
    const councilTarget = useCortexStore((s) => s.councilTarget);
    const councilMembers = useCortexStore((s) => s.councilMembers);
    const setCouncilTarget = useCortexStore((s) => s.setCouncilTarget);
    const fetchCouncilMembers = useCortexStore((s) => s.fetchCouncilMembers);
    const selectedTeamId = useCortexStore((s) => s.selectedTeamId);
    const teamsDetail = useCortexStore((s) => s.teamsDetail);

    const [input, setInput] = useState("");
    const [pendingContinuationContext, setPendingContinuationContext] = useState<MissionChatContinuationContext | null>(null);
    const [broadcastMode, setBroadcastMode] = useState(false);
    const [fetchedMembers, setFetchedMembers] = useState(false);
    const [isReplyingToProposal, setIsReplyingToProposal] = useState(false);
    const inputRef = useRef<HTMLTextAreaElement>(null);
    const scrollRef = useRef<HTMLDivElement>(null);
    const showAdvancedRouting = !simpleMode;
    const directTarget = councilTarget === "admin" ? null : councilTarget;
    const currentTeamId = focusedTeamId || selectedTeamId;
    const currentTeam = currentTeamId
        ? teamsDetail.find((team) => team.id === currentTeamId) ?? null
        : null;
    const chatScope = useMemo(
        () => buildMissionChatScope(organizationId, currentTeamId),
        [organizationId, currentTeamId],
    );
    const activeSuggestions = currentTeam ? teamSuggestions(currentTeam.name) : suggestions;
    const isLoading = isMissionChatting || isBroadcasting || isReplyingToProposal;
    const lastUserMessage = [...missionChat].reverse().find((m) => m.role === "user");
    useSomaOutputContinuation({ disabled: isLoading, inputRef, setInput, setContinuationContext: setPendingContinuationContext });

    useEffect(() => {
        setCouncilTarget("admin");
    }, [setCouncilTarget]);

    useEffect(() => {
        if (typeof window === "undefined") return;
        const params = new URLSearchParams(window.location.search);
        if (params.get("fresh") !== "1" && params.get("reset_chat") !== "1") return;
        clearAllPersistedChat();
        clearMissionChat();
        params.delete("fresh");
        params.delete("reset_chat");
        const nextSearch = params.toString();
        const nextUrl = `${window.location.pathname}${nextSearch ? `?${nextSearch}` : ""}${window.location.hash}`;
        window.history.replaceState(null, "", nextUrl);
    }, [clearMissionChat]);

    useEffect(() => {
        setMissionChatScope(chatScope);
    }, [chatScope, setMissionChatScope]);

    useEffect(() => {
        if (!showAdvancedRouting) return;
        fetchCouncilMembers().then(() => setFetchedMembers(true));
    }, [fetchCouncilMembers, showAdvancedRouting]);

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [missionChat.length]);

    const retryLastMessage = () => {
        if (!lastUserMessage) return;
        const content = lastUserMessage.content.replace(/^\[BROADCAST\]\s*/i, "");
        if (content.trim()) sendMissionChat(content);
    };

    const handleSubmit = () => {
        if (!input.trim() || isLoading) return;
        const trimmed = input.trim();
        const trimmedStart = input.trimStart();
        const isBroadcast = showAdvancedRouting && (broadcastMode || trimmedStart.startsWith("/all "));
        const content = isBroadcast && trimmedStart.startsWith("/all ")
            ? trimmedStart.slice(5).trim()
            : trimmed;

        if (!content) return;
        const proposalReply = pendingProposal && !isBroadcast
            ? conversationalProposalReply(content)
            : null;
        if (proposalReply === "confirm") {
            setInput("");
            setPendingContinuationContext(null);
            setIsReplyingToProposal(true);
            void confirmProposal(pendingProposal ?? undefined, content)
                .finally(() => setIsReplyingToProposal(false));
        } else if (proposalReply === "cancel") {
            setInput("");
            setPendingContinuationContext(null);
            cancelProposal(content);
        } else if (isBroadcast) {
            broadcastToSwarm(content);
            setPendingContinuationContext(null);
        } else {
            sendMissionChat(content, pendingContinuationContext ? { continuation_context: pendingContinuationContext } : undefined);
            setPendingContinuationContext(null);
        }
        if (!proposalReply) setInput("");
    };

    const applyStarterPrompt = (prompt: string) => {
        if (isLoading) return;
        setPendingContinuationContext(null);
        setInput(prompt);
        inputRef.current?.focus();
    };

    const updateInput = (value: string) => setInput(value);

    const clearContinuation = () => {
        setPendingContinuationContext(null);
        inputRef.current?.focus();
    };

    const clearChat = () => {
        setPendingContinuationContext(null);
        clearMissionChat();
    };

    const retryCouncilMembers = () => {
        setFetchedMembers(false);
        fetchCouncilMembers().then(() => setFetchedMembers(true));
    };

    return (
        <div className="flex h-full min-h-0 flex-col" data-testid="mission-chat">
            <MissionControlChatHeader
                assistantName={assistantName}
                broadcastMode={broadcastMode}
                clearMissionChat={clearChat}
                councilMembers={councilMembers}
                directTarget={directTarget}
                focusedTeamName={currentTeam?.name}
                isLoading={isLoading}
                messageCount={missionChat.length}
                setBroadcastMode={setBroadcastMode}
                setCouncilTarget={setCouncilTarget}
                setDirectTarget={(target) => setCouncilTarget(target ?? "admin")}
                showAdvancedRouting={showAdvancedRouting}
                simpleMode={simpleMode}
            />

            {showAdvancedRouting && broadcastMode && <BroadcastModeIndicator />}

            <SomaConversationThread scrollRef={scrollRef}>
                {missionChatFailure && (
                    <CouncilCallErrorCard
                        failure={missionChatFailure}
                        onRetry={retryLastMessage}
                        onSwitchToSoma={() => {
                            setCouncilTarget("admin");
                            retryLastMessage();
                        }}
                        onContinueWithSoma={() => {
                            setCouncilTarget("admin");
                        }}
                    />
                )}

                {missionChat.length === 0 ? (
                    showAdvancedRouting && fetchedMembers && councilMembers.length === 0 ? (
                        <SomaOfflineGuide assistantName={assistantName} onRetry={retryCouncilMembers} />
                    ) : (
                        <MissionControlEmptyState
                            assistantName={assistantName}
                            broadcastMode={broadcastMode}
                            currentTeamName={currentTeam?.name}
                            directTarget={directTarget}
                            showAdvancedRouting={showAdvancedRouting}
                            simpleMode={simpleMode}
                            suggestions={activeSuggestions}
                        />
                    )
                ) : (
                    presentedMissionChat.map((msg, i) => (
                        <MissionControlMessageBubble key={i} msg={msg} compactResult={simpleMode} />
                    ))
                )}

                {isLoading && (
                    <div className="flex gap-2 justify-start">
                        <div
                            className={`w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0 ${
                                showAdvancedRouting && isBroadcasting
                                    ? "bg-cortex-warning/10 border border-cortex-warning/20"
                                    : "bg-cortex-primary/10 border border-cortex-primary/20"
                            }`}
                        >
                            {showAdvancedRouting && isBroadcasting ? (
                                <Megaphone className="w-3.5 h-3.5 text-cortex-warning animate-pulse" />
                            ) : (
                                <Brain className="w-3.5 h-3.5 text-cortex-primary animate-pulse" />
                            )}
                        </div>
                        <div
                            className={`rounded-lg ${
                                showAdvancedRouting && isBroadcasting
                                    ? "bg-cortex-warning/5 border border-cortex-warning/20"
                                    : "bg-cortex-primary/5 border border-cortex-primary/20"
                            }`}
                        >
                            <SomaActivityIndicator isBroadcasting={isBroadcasting} assistantName={assistantName} />
                        </div>
                    </div>
                )}
            </SomaConversationThread>

            <div className="px-3 py-2 border-t border-cortex-border flex-shrink-0">
                <MissionControlTeamContinuationPrompt
                    messages={missionChat}
                    disabled={isLoading}
                    onStarterPrompt={applyStarterPrompt}
                />
                {pendingContinuationContext ? (
                    <MissionControlContinuationChip context={pendingContinuationContext} onClear={clearContinuation} />
                ) : null}
                {simpleMode ? (
                    <SomaIntentInput
                        value={input}
                        onChange={updateInput}
                        onSubmit={handleSubmit}
                        inputRef={inputRef}
                        autoFocus={autoFocus}
                        loading={isLoading}
                        disabled={isLoading}
                        placeholder={somaPlaceholder({
                            assistantName,
                            broadcastMode,
                            currentTeamName: currentTeam?.name,
                            directTarget,
                            showAdvancedRouting,
                            simpleMode,
                        })}
                    />
                ) : (
                    <MissionControlAdvancedInput
                        value={input}
                        onChange={updateInput}
                        onSubmit={handleSubmit}
                        inputRef={inputRef}
                        autoFocus={autoFocus}
                        isLoading={isLoading}
                        broadcastMode={broadcastMode}
                        placeholder={somaPlaceholder({
                            assistantName,
                            broadcastMode,
                            currentTeamName: currentTeam?.name,
                            directTarget,
                            showAdvancedRouting,
                            simpleMode,
                        })}
                    />
                )}
            </div>

            {showAdvancedRouting ? <OrchestrationInspector /> : null}
        </div>
    );
}
