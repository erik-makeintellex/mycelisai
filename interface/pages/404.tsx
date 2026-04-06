// Keep a minimal Pages Router 404 present so standalone builds on Windows
// emit the legacy pages manifest that Next still expects for error routing.
export default function LegacyNotFoundShim() {
    return null;
}
