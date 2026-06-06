export default function MetadataItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div
        style={{
          fontSize: "0.75rem",
          color: "#6f6f6f",
          marginBottom: "0.25rem",
        }}
      >
        {label}
      </div>
      <div
        style={{
          fontSize: "0.9rem",
          color: "#161616",
        }}
      >
        {value || "—"}
      </div>
    </div>
  );
}
