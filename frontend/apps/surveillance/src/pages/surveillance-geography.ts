// Explicit spelling variants across the legacy district and UBOS 2024 layers.
// Keep source labels for display; only normalize the keys used for matching.
export function districtKey(value: unknown): string {
  const name = typeof value === "string" ? value.trim().toLowerCase() : "";
  if (name === "luwero") return "luweero";
  if (name === "sembabule") return "ssembabule";
  return name;
}
