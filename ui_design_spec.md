# Mycelis UI Design Specification

**Status**: Draft
**Target Audience**: Architect Agent
**Goal**: Transform the current functional UI into a premium, "Sci-Fi / Cyberpunk" inspired professional interface.

## 1. Design Philosophy: "Mycelis Dark"
The interface should feel like a high-end command center for orchestrating intelligent agents.
*   **Aesthetic**: Dark, sleek, data-dense but breathable.
*   **Metaphor**: "The Substrate" - organic connections, glowing nodes, living data.
*   **Key Characteristics**:
    *   **Glassmorphism**: Heavy use of `backdrop-blur` and semi-transparent backgrounds (`bg-opacity-90`).
    *   **Neon Accents**: Status indicators should glow.
    *   **Micro-Interactions**: Hover states should be smooth (`transition-all duration-300`).
    *   **Monospace Data**: All IDs, logs, and raw data should use a high-quality monospace font (JetBrains Mono or Fira Code).

## 2. Design Tokens (Tailwind v4)

### Colors
Update `globals.css` with these refined tokens:
```css
:root {
  /* Backgrounds */
  --bg-primary: #0f1115;       /* Deepest Void */
  --bg-secondary: #161b22;     /* Panel Background */
  --bg-glass: rgba(22, 27, 34, 0.7); /* Glass Effect */

  /* Borders */
  --border-subtle: #30363d;
  --border-highlight: #58a6ff;

  /* Text */
  --text-primary: #f0f6fc;
  --text-secondary: #8b949e;
  --text-mono: #d2a8ff;        /* Code/Data Text */

  /* Accents */
  --accent-success: #2ea043;   /* Success Green */
  --accent-error: #da3633;     /* Error Red */
  --accent-warn: #d29922;      /* Warning Gold */
  --accent-info: #58a6ff;      /* Info Blue */
  --accent-glow: rgba(88, 166, 255, 0.15);
}
```

### Typography
*   **Headings**: Inter (Bold/ExtraBold). Tracking tight.
*   **Body**: Inter (Regular).
*   **Data/Logs**: JetBrains Mono (if available) or ui-monospace.

## 3. Component Specifications

### A. Navigation Bar (`Navbar.tsx`)
*   **Style**: Sticky top, `backdrop-blur-xl`, border-b `border-[--border-subtle]`.
*   **Logo**: "Mycelis" in white, "AI" in gradient text (Blue -> Purple).
*   **Links**:
    *   Default: Text secondary, hover text primary.
    *   Active: Text primary, subtle background pill (`bg-[--accent-info]/10`), glowing bottom border.
*   **System Status**:
    *   Pulse animation on the green dot.
    *   Tooltip on hover showing "All Systems Operational".

### B. Dashboard (`page.tsx`)
*   **Hero Section**:
    *   Greeting text with a subtle gradient.
    *   Quick Action buttons (Manage Agents, System Config) should be "Ghost" buttons with glowing borders on hover.
*   **Stats Grid**:
    *   Cards should use `bg-[--bg-secondary]` with a subtle gradient overlay.
    *   **Hover Effect**: Lift up (`-translate-y-1`) and shadow glow.
    *   **Numbers**: Large, bold, monospace font.
*   **Live Event Feed**:
    *   **Container**: Look like a terminal window. Darker background (`#0d1117`).
    *   **Rows**: Zebra striping (very subtle).
    *   **New Events**: Animate in (slide down + fade in).
    *   **Syntax Highlighting**: JSON payloads should be colored (Keys: Blue, Values: Green/String).

### C. Agent Card (New Component)
*   **Layout**: Grid of cards.
*   **Header**: Agent Name + Role Badge (Pill shape).
*   **Body**:
    *   "Status": Online/Offline (Dot).
    *   "Model": Icon (Ollama/OpenAI).
    *   "Last Active": Time ago.
*   **Footer**: Action buttons (Chat, Logs, Edit).

## 4. Implementation Plan (For Architect)

### Phase 1: Foundation
1.  **Update `globals.css`**: Apply the new color palette.
2.  **Install Dependencies**:
    *   `framer-motion` (for animations).
    *   `clsx`, `tailwind-merge` (for class utility).
    *   `lucide-react` (ensure latest version).

### Phase 2: Component Refactor
1.  **Refactor `Navbar`**: Implement the glassmorphism and active states.
2.  **Refactor `Dashboard`**: Break down into smaller components (`StatsCard`, `EventFeed`, `StatusPanel`).
3.  **Apply Animations**: Add entry animations to the dashboard widgets.

### Phase 3: Verification
1.  **Local Test**: Run `npm run dev` and verify visually.
2.  **Build Test**: Run `npm run build` to ensure no type errors.
3.  **Deploy**: Update Docker image and redeploy to K8s.

## 5. Testing Instructions
*   **Local**: `cd ui && npm run dev`. Open `http://localhost:3000`.
*   **K8s**: `make k8s-dev SERVICE=ui`. Open `http://localhost/`.
