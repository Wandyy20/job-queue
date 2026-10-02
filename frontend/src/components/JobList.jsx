import { useState, useEffect } from "react";
import { listJobs } from "../api/jobsApi";
import StatusDot from "./StatusDot";
import StatsOverview from "./StatsOverview";

const TYPE_ICONS = {
  test_job: "○",
  send_webhook: "↗",
  summarize_text: "≡",
  classify_sentiment: "◐",
  translate_text: "⇄",
  resize_image: "▢",
  generate_pdf_report: "▤",
  csv_export: "▦",
  classify_toxic_comment: "⚠",
};

const STATUS_COLOR = {
  pending: "#FBBF24",
  processing: "#818CF8",
  completed: "#34D399",
  failed: "#F87171",
  dead: "#71717A",
  cancelled: "#71717A",
};

function formatDateTime(value) {
  const d = new Date(value);
  return d.toLocaleString([], {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function JobList({ refreshTrigger, selectedJobId, onSelectJob }) {
  const [jobs, setJobs] = useState([]);
  const [filter, setFilter] = useState("all");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchJobs() {
      try {
        const data = await listJobs();
        setJobs(data || []);
      } catch (err) {
        console.error("Failed to fetch jobs:", err);
      } finally {
        setLoading(false);
      }
    }

    fetchJobs();
    const interval = setInterval(fetchJobs, 3000);
    return () => clearInterval(interval);
  }, [refreshTrigger]);

  const filtered = filter === "all" ? jobs : jobs.filter((j) => j.Status === filter);
  const tabs = ["all", "pending", "processing", "completed", "failed", "dead"];

  return (
    <div className="space-y-4">
      <StatsOverview jobs={jobs} />

      <div className="flex gap-1 border-b border-[var(--border)]">
        {tabs.map((t) => (
          <button
            key={t}
            onClick={() => setFilter(t)}
            className={`px-3 py-2 text-sm capitalize border-b-2 transition-colors ${
              filter === t
                ? "border-[var(--accent)] text-[var(--text)]"
                : "border-transparent text-[var(--text-muted)] hover:text-[var(--text)]"
            }`}
          >
            {t}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="p-3 space-y-2 rounded-lg border border-[var(--border)]">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="h-12 bg-[var(--surface)] border border-[var(--border)] rounded-md animate-pulse"
            />
          ))}
        </div>
      ) : (
        <div className="divide-y divide-[var(--border)] rounded-lg border border-[var(--border)] overflow-hidden">
          {filtered.length === 0 && (
            <div className="p-8 text-center text-[var(--text-muted)] text-sm">
              No jobs here yet.
            </div>
          )}
          {filtered.map((job) => (
            <button
              key={job.ID}
              onClick={() => onSelectJob(job.ID)}
              className="w-full text-left flex items-center gap-4 p-3 transition-colors border-l-2"
              style={{
                borderLeftColor: STATUS_COLOR[job.Status] || "#71717A",
                backgroundColor: selectedJobId === job.ID ? "#1A1A22" : "var(--surface)",
              }}
            >
              <span className="text-lg text-[var(--text-muted)] w-5 text-center">
                {TYPE_ICONS[job.Type] || "•"}
              </span>
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium truncate">{job.Type}</div>
                <div className="font-mono text-xs text-[var(--text-muted)] truncate">
                  {job.ID}
                </div>
              </div>
              <StatusDot status={job.Status} />
              <span className="text-xs text-[var(--text-muted)] font-mono">
                {job.Attempts}/{job.MaxAttempts}
              </span>
              <span className="text-xs text-[var(--text-muted)] w-36 text-right font-mono">
                {formatDateTime(job.CreatedAt)}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}