export function compactStrings(values: unknown[]): string[] {
  return Array.from(new Set(values.filter((value): value is string =>
    typeof value === "string" && value.trim().length > 0,
  ).map((value) => value.trim())));
}

export function compactId(id: string) {
  return id.length <= 12 ? id : `${id.slice(0, 8)}...`;
}
