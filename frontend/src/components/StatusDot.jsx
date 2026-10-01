const STATUS_CONFIG = {
  pending:    { color: "#FBBF24", label: "pending" },
  processing: { color: "#818CF8", label: "processing", pulse: true },
  completed:  { color: "#34D399", label: "completed" },
  failed:     { color: "#F87171", label: "failed" },
  dead:       { color: "#71717A", label: "dead" },
  cancelled:  { color: "#71717A", label: "cancelled" },
};

export default function StatusDot({ status }) {
  const config = STATUS_CONFIG[status] || STATUS_CONFIG.dead;
  return (
    <span className="inline-flex items-center gap-2">
      <span className="relative flex h-2 w-2">
        {config.pulse && (
          <span
            className="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75"
            style={{ backgroundColor: config.color }}
          />
        )}
        <span
          className="relative inline-flex rounded-full h-2 w-2"
          style={{ backgroundColor: config.color }}
        />
      </span>
      <span className="text-sm capitalize" style={{ color: config.color }}>
        {config.label}
      </span>
    </span>
  );
}