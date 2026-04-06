import type { AppProps } from 'next/app';

// Keep a minimal Pages Router shell present so standalone builds on Windows
// still emit the legacy pages manifest that Next expects during packaging.
export default function LegacyAppShim({ Component, pageProps }: AppProps) {
    return <Component {...pageProps} />;
}
