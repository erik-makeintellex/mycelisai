import type { NextPageContext } from 'next';

type LegacyErrorShimProps = {
    statusCode?: number;
};

// Keep a minimal Pages Router error shell present so standalone builds on
// Windows emit the legacy pages manifest that Next still expects.
export default function LegacyErrorShim(_props: LegacyErrorShimProps) {
    return null;
}

LegacyErrorShim.getInitialProps = ({ res, err }: NextPageContext): LegacyErrorShimProps => ({
    statusCode: res?.statusCode ?? err?.statusCode ?? 500,
});
