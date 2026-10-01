export default function StatsOverview({ jobs }) {
  const counts = jobs.reduce((acc, job) => {
    acc[job.Status] = (acc[job.Status] || 0) + 1;
    return acc;
  }, {});

  const stats = [
    { label: "Total", value: jobs.length, color: "#F4F4F5" },
    { label: "Pending", value: counts.pending || 0, color: "#FBBF24" },
    { label: "Processing", value: counts.processing || 0, color: "#818CF8" },
    { label: "Completed", value: counts.completed || 0, color: "#34D399" },
    { label: "Failed / Dead", value: (counts.failed || 0) + (counts.dead || 0), color: "#F87171" },
  ];

  return (
    <div className="flex items-stretch divide-x divide-[var(--border)] rounded-lg border border-[var(--border)] bg-[var(--surface)] overflow-x-auto">
      {stats.map((s) => (
        <div key={s.label} className="flex items-baseline gap-2 px-5 py-3.5 flex-1 whitespace-nowrap">
          <span className="font-mono text-xl font-semibold" style={{ color: s.color }}>
            {s.value}
          </span>
          <span className="text-xs text-[var(--text-muted)]">{s.label}</span>
        </div>
      ))}
    </div>
  );
}