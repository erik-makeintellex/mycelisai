"use client";

import { useEffect, useRef, useState } from "react";
import { Brain, Megaphone } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import { SomaConversationThread } from "@/components/soma/SomaConversationThread";
import { SomaIntentInput } from "@/components/soma/SomaIntentInput";
import {
    DEFAULT_SOMA_SUGGESTIONS,
    type SomaSuggestion,
} from "@/components/soma/SomaSuggestionBar";
import CouncilCallErrorCard from "./CouncilCallErrorCard";
import {
    BroadcastModeIndicator,
    MissionControlEmptyState,
    SomaOfflineGuide,
} from "./MissionControlChatStates";
import {
    MissionControlChatHeader,
    SomaActivityIndicator,
} from "./MissionControlChatChrome";
import MissionControlMessageBubble from "./MissionControlMessageBubble";
import { MissionControlAdvancedInput } from "./MissionControlAdvancedInput";
import MissionControlTeamContinuationPrompt from "./MissionControlTeamContinuationPrompt";
import OrchestrationInspector from "./OrchestrationInspector";
import { somaPlaceholder, teamSuggestions } from "./missionControlChatUi";

export default function MissionControlChat({
    simpleMode = false,
    autoFocus = false,
    organizationId,
    suggestions = DEFAULT_SOMA_SUGGESTIONS,
}: {
    simpleMode?: boolean;
    autoFocus?: boolean;
    organizationId?: string;
    suggestions?: readonly SomaSuggestion[];
}) {
    const missionChat = useCortexStore((s) => s.missionChat);
    const isMissionChatting = useCortexStore((s) => s.isMissionChatting);
    const missionChatFailure = useCortexStore((s) => s.missionChatFailure);
    const sendMissionChat = useCortexStore((s) => s.sendMissionChat);
    const clearMissionChat = useCortexStore((s) => s.clearMissionChat);
    const setMissionChatScope = useCortexStore((s) => s.setMissionChatScope);
    const broadcastToSwarm = useCortexStore((s) => s.broadcastToSwarm);
    const isBroadcasting = useCortexStore((s) => s.isBroadcasting);
    const assistantName = useCortexStore((s) => s.assistantName);
    const councilTarget = useCortexStore((s) => s.councilTarget);
    const councilMembers = useCortexStore((s) => s.councilMembers);
    const setCouncilTarget = useCortexStore((s) => s.setCouncilTarget);
    const fetchCouncilMembers = useCortexStore((s) => s.fetchCouncilMembers);
    const selectedTeamId = useCortexStore((s) => s.selectedTeamId);
    const teamsDetail = useCortexStore((s) => s.teamsDetail);

    const [input, setInput] = useState("");
    const [broadcastMode, setBroadcastMode] = useState(false);
    const [fetchedMembers, setFetchedMembers] = useState(false);
    const [directTarget, setDirectTarget] = useState<string | null>(null);
    const inputRef = useRef<HTMLInputElement>(null);
    const scrollRef = useRef<HTMLDivElement>(null);
    const showAdvancedRouting = !simpleMode;
    const currentTeam = selectedTeamId
        ? teamsDetail.find((team) => team.id === selectedTeamId) ?? null
        : null;
    const activeSuggestions = currentTeam ? teamSuggestions(currentTeam.name) : suggestions;
    const isLoading = isMissionChatting || isBroadcasting;
    const lastUserMessage = [...missionChat].reverse().find((m) => m.role === "user");

    useEffect(() => {
        setCouncilTarget("admin");
    }, [setCouncilTarget]);

    useEffect(() => {
        setMissionChatScope(organizationId ?? null);
    }, [organizationId, setMissionChatScope]);

    useEffect(() => {
        setDirectTarget(councilTarget === "admin" ? null : councilTarget);
    }, [councilTarget]);

    useEffect(() => {
        if (!showAdvancedRouting) {
            setFetchedMembers(false);
            return;
        }
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
        if (isBroadcast) {
            broadcastToSwarm(content);
        } else {
            sendMissionChat(content);
        }
        setInput("");
    };

    const applyStarterPrompt = (prompt: string) => {
        if (isLoading) return;
        setInput(prompt);
        inputRef.current?.focus();
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
                clearMissionChat={clearMissionChat}
                councilMembers={councilMembers}
                directTarget={directTarget}
                isLoading={isLoading}
                messageCount={missionChat.length}
                setBroadcastMode={setBroadcastMode}
                setCouncilTarget={setCouncilTarget}
                setDirectTarget={setDirectTarget}
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
                            setDirectTarget(null);
                            setCouncilTarget("admin");
                            retryLastMessage();
                        }}
                        onContinueWithSoma={() => {
                            setDirectTarget(null);
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
                            onStarterPrompt={applyStarterPrompt}
                            showAdvancedRouting={showAdvancedRouting}
                            simpleMode={simpleMode}
                            suggestions={activeSuggestions}
                        />
                    )
                ) : (
                    missionChat.map((msg, i) => <MissionControlMessageBubble key={i} msg={msg} />)
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
                {simpleMode ? (
                    <SomaIntentInput
                        value={input}
                        onChange={setInput}
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
                        onChange={setInput}
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
