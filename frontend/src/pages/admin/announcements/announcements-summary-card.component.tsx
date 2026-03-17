export default function SummaryCard({ label, value }: { label: string; value: number }) {
  return (
    <div
      style={{
        border: "1px solid #e0e0e0",
        padding: "0.85rem",
        background: "#ffffff",
      }}
    >
      <div
        style={{
          fontSize: "0.75rem",
          color: "#6f6f6f",
          marginBottom: "0.35rem",
        }}
      >
        {label}
      </div>
      <div
        style={{
          fontSize: "1.25rem",
          fontWeight: 600,
          color: "#161616",
        }}
      >
        {value}
      </div>
    </div>
  );
}
