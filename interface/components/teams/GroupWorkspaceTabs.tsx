import type { KeyboardEvent } from "react";
import type { LucideIcon } from "lucide-react";
import {
  ClipboardList,
  MessageSquare,
  Plus,
  Settings2,
  Users,
} from "lucide-react";

export type GroupWorkspacePanel =
  | "overview"
  | "outputs"
  | "message"
  | "settings"
  | "create";

type GroupWorkspaceTabsProps = {
  activePanel: GroupWorkspacePanel;
  outputCount: number;
  onSelect: (panel: GroupWorkspacePanel) => void;
};

export function GroupWorkspaceTabs({
  activePanel,
  outputCount,
  onSelect,
}: GroupWorkspaceTabsProps) {
  return (
    <div
      className="grid grid-cols-2 gap-2 border-b border-cortex-border bg-cortex-bg/40 p-2 sm:flex sm:overflow-x-auto sm:[scrollbar-width:none] sm:[&::-webkit-scrollbar]:hidden"
      role="tablist"
      aria-label="Group workspace sections"
      onKeyDown={(event) => handleTabKeyDown(event, activePanel, onSelect)}
    >
      {workspacePanelTabs(outputCount).map((panel) => (
        <GroupWorkspaceTab
          key={panel.id}
          panel={panel}
          selected={activePanel === panel.id}
          onSelect={onSelect}
        />
      ))}
    </div>
  );
}

function GroupWorkspaceTab({
  panel,
  selected,
  onSelect,
}: {
  panel: PanelTab;
  selected: boolean;
  onSelect: (panel: GroupWorkspacePanel) => void;
}) {
  const Icon = panel.icon;
  return (
    <button
      type="button"
      role="tab"
      aria-selected={selected}
      aria-controls={`groups-${panel.id}-panel`}
      id={`groups-${panel.id}-tab`}
      onClick={() => onSelect(panel.id)}
      className={`rounded-xl border px-3 py-2 text-left transition-colors ${
        selected
          ? "border-cortex-primary/45 bg-cortex-primary/10 text-cortex-text-main"
          : "border-cortex-border bg-cortex-surface text-cortex-text-muted hover:text-cortex-text-main"
      } min-w-0 sm:min-w-36 sm:shrink-0`}
    >
      <span className="flex items-center gap-2 text-xs font-semibold">
        <Icon className="h-3.5 w-3.5" />
        {panel.label}
      </span>
      <span className="mt-1 hidden truncate text-[11px] leading-4 sm:block">
        {panel.summary}
      </span>
    </button>
  );
}

type PanelTab = {
  id: GroupWorkspacePanel;
  label: string;
  summary: string;
  icon: LucideIcon;
};

function workspacePanelTabs(outputCount: number): PanelTab[] {
  return [
    {
      id: "overview",
      label: "Overview",
      summary: "Scope and links",
      icon: ClipboardList,
    },
    {
      id: "outputs",
      label: "Outputs",
      summary: `${outputCount} retained`,
      icon: Users,
    },
    {
      id: "message",
      label: "Message",
      summary: "Broadcast ask",
      icon: MessageSquare,
    },
    {
      id: "settings",
      label: "Settings",
      summary: "Policy and tools",
      icon: Settings2,
    },
    {
      id: "create",
      label: "Create",
      summary: "New group",
      icon: Plus,
    },
  ];
}

const panelOrder: GroupWorkspacePanel[] = [
  "overview",
  "outputs",
  "message",
  "settings",
  "create",
];

function handleTabKeyDown(
  event: KeyboardEvent<HTMLDivElement>,
  activePanel: GroupWorkspacePanel,
  onSelect: (panel: GroupWorkspacePanel) => void,
) {
  const keyOffsets: Record<string, number | undefined> = {
    ArrowRight: 1,
    ArrowDown: 1,
    ArrowLeft: -1,
    ArrowUp: -1,
  };
  const currentIndex = panelOrder.indexOf(activePanel);
  let nextIndex = currentIndex;

  if (event.key === "Home") nextIndex = 0;
  else if (event.key === "End") nextIndex = panelOrder.length - 1;
  else if (keyOffsets[event.key]) {
    nextIndex =
      (currentIndex + keyOffsets[event.key]! + panelOrder.length) %
      panelOrder.length;
  } else {
    return;
  }

  event.preventDefault();
  const nextPanel = panelOrder[nextIndex];
  onSelect(nextPanel);
  requestAnimationFrame(() => {
    document.getElementById(`groups-${nextPanel}-tab`)?.focus();
  });
}
