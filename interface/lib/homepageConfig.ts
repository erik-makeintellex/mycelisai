export interface HomepageCTA {
    label: string;
    href: string;
    external?: boolean;
}

export interface HomepageSection {
    title: string;
    body: string;
}

export interface HomepageLink {
    label: string;
    href: string;
    description: string;
    external?: boolean;
}

export interface HomepageConfig {
    enabled: boolean;
    brand: {
        product_name: string;
        tagline: string;
        logo_url?: string;
    };
    hero: {
        headline: string;
        subheadline: string;
        primary_cta: HomepageCTA;
        secondary_cta: HomepageCTA;
    };
    announcement?: {
        enabled: boolean;
        text: string;
    };
    sections: HomepageSection[];
    links: HomepageLink[];
    footer_text?: string;
    config_issue?: string;
}

export const defaultHomepageConfig: HomepageConfig = {
    enabled: true,
    brand: {
        product_name: "Mycelis",
        tagline: "Soma-centered AI organization orchestration",
    },
    hero: {
        headline: "Operate AI Organizations through Soma",
        subheadline: "Express intent once. Soma coordinates teams, tools, reviews, and governed execution behind the scenes.",
        primary_cta: { label: "Start with Soma", href: "/dashboard" },
        secondary_cta: { label: "View Documentation", href: "/docs" },
    },
    sections: [
        { title: "Express intent", body: "Describe what you want to accomplish in plain language." },
        { title: "Soma coordinates", body: "Soma turns intent into structured work across teams, tools, memory, and reviews." },
        { title: "Governed execution", body: "Approvals, audit records, and capability boundaries keep actions controlled." },
        { title: "Review outcomes", body: "Activity, artifacts, and audit records show what changed and why." },
    ],
    links: [
        { label: "Documentation", href: "/docs", description: "Read setup and architecture guidance." },
        { label: "Resources and MCP", href: "/resources", description: "Review connected tools and runtime capabilities." },
        { label: "Activity and Audit", href: "/activity", description: "Inspect system activity and governed execution." },
        { label: "Status", href: "/system", description: "Review service health and recovery status." },
    ],
    footer_text: "Self-hosted AI Organization orchestration with Soma.",
};

export function markExternal(href: string): boolean {
    return /^https?:\/\//i.test(href) || /^mailto:/i.test(href);
}

export function safeHref(href: string): string {
    const trimmed = href.trim();
    if (trimmed.startsWith("/") || markExternal(trimmed)) return trimmed;
    return trimmed ? "#" : "";
}

export function normalizeHomepageConfig(raw: unknown): HomepageConfig {
    if (!raw || typeof raw !== "object") return defaultHomepageConfig;
    const candidate = raw as Partial<HomepageConfig>;
    const hero = candidate.hero ?? defaultHomepageConfig.hero;
    const brand = candidate.brand ?? defaultHomepageConfig.brand;
    const sections = Array.isArray(candidate.sections) && candidate.sections.length > 0
        ? candidate.sections
        : defaultHomepageConfig.sections;
    const links = Array.isArray(candidate.links) && candidate.links.length > 0
        ? candidate.links
        : defaultHomepageConfig.links;

    return {
        enabled: candidate.enabled !== false,
        brand: {
            product_name: textOr(brand.product_name, defaultHomepageConfig.brand.product_name),
            tagline: textOr(brand.tagline, defaultHomepageConfig.brand.tagline),
            logo_url: safeHref(brand.logo_url ?? ""),
        },
        hero: {
            headline: textOr(hero.headline, defaultHomepageConfig.hero.headline),
            subheadline: textOr(hero.subheadline, defaultHomepageConfig.hero.subheadline),
            primary_cta: normalizeCTA(hero.primary_cta, defaultHomepageConfig.hero.primary_cta),
            secondary_cta: normalizeCTA(hero.secondary_cta, defaultHomepageConfig.hero.secondary_cta),
        },
        announcement: candidate.announcement,
        sections: sections.map(normalizeSection).filter((item) => item.title && item.body),
        links: links.map(normalizeLink).filter((item) => item.label && item.href),
        footer_text: textOr(candidate.footer_text, defaultHomepageConfig.footer_text ?? ""),
        config_issue: candidate.config_issue,
    };
}

function normalizeCTA(raw: HomepageCTA | undefined, fallback: HomepageCTA): HomepageCTA {
    const href = safeHref(raw?.href ?? fallback.href);
    return {
        label: textOr(raw?.label, fallback.label),
        href,
        external: raw?.external || markExternal(href),
    };
}

function normalizeSection(raw: HomepageSection): HomepageSection {
    return { title: textOr(raw.title, ""), body: textOr(raw.body, "") };
}

function normalizeLink(raw: HomepageLink): HomepageLink {
    const href = safeHref(raw.href ?? "");
    return {
        label: textOr(raw.label, ""),
        href,
        description: textOr(raw.description, ""),
        external: raw.external || markExternal(href),
    };
}

function textOr(value: unknown, fallback: string): string {
    return typeof value === "string" && value.trim() ? value.trim() : fallback;
}
