import { useState, useEffect } from "react";
import { getJob, getJobEvents, cancelJob } from "../api/jobsApi";
import StatusDot from "./StatusDot";

function formatDateTime(value) {
  if (!value) return "—";
  return new Date(value).toLocaleString([], {
    month: "short", day: "numeric", hour: "2-digit", minute: "2-digit", second: "2-digit",
  });
}

function base64ToBlob(base64, mime) {
  const byteChars = atob(base64);
  const byteNumbers = new Array(byteChars.length);
  for (let i = 0; i < byteChars.length; i++) {
    byteNumbers[i] = byteChars.charCodeAt(i);
  }
  return new Blob([new Uint8Array(byteNumbers)], { type: mime });
}

function downloadBlob(blob, filename) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

function ResultView({ type, result }) {
  if (type === "generate_pdf_report" && result.pdf_base64) {
    return (
      <button
        onClick={() => downloadBlob(base64ToBlob(result.pdf_base64, "application/pdf"), `${result.title || "report"}.pdf`)}
        className="text-sm text-[var(--accent)] border border-[var(--accent)]/30 rounded-md px-3 py-1.5 hover:bg-[var(--accent)]/10 transition-colors"
      >
        ↓ Download PDF
      </button>
    );
  }

  if (type === "csv_export" && result.csv_base64) {
    return (
      <button
        onClick={() => downloadBlob(base64ToBlob(result.csv_base64, "text/csv"), "export.csv")}
        className="text-sm text-[var(--accent)] border border-[var(--accent)]/30 rounded-md px-3 py-1.5 hover:bg-[var(--accent)]/10 transition-colors"
      >
        ↓ Download CSV ({result.row_count} rows)
      </button>
    );
  }

  if (type === "resize_image" && result.resized_image_base64) {
    return (
      <img
        src={`data:image/png;base64,${result.resized_image_base64}`}
        alt="resized"
        className="rounded-md border border-[var(--border)] max-w-full"
      />
    );
  }

  return (
    <pre className="font-mono text-xs bg-[var(--bg)] border border-[var(--border)] rounded-md p-3 overflow-x-auto whitespace-pre-wrap break-words">
      {JSON.stringify(result, null, 2)}
    </pre>
  );
}

export default function JobDetail({ jobId, onClose, onChanged }) {
  const [job, setJob] = useState(null);
  const [events, setEvents] = useState([]);
  const [cancelling, setCancelling] = useState(false);

useEffect(() => {
    if (!jobId) return;

    async function fetchDetail() {
      try {
        const [jobData, eventsData] = await Promise.all([
          getJob(jobId),
          getJobEvents(jobId),
        ]);
        setJob(jobData);
        setEvents(eventsData || []);
      } catch (err) {
        console.error("Failed to fetch job detail:", err);
      }
    }

    fetchDetail();
    const interval = setInterval(fetchDetail, 3000);

    return () => clearInterval(interval);
  }, [jobId]);


async function handleCancel() {
    if (!jobId) return;
    setCancelling(true);
    try {
      await cancelJob(jobId);
      
      const [jobData, eventsData] = await Promise.all([
        getJob(jobId),
        getJobEvents(jobId),
      ]);
      setJob(jobData);
      setEvents(eventsData || []);

      onChanged?.();
    } catch (err) {
      alert("Failed to cancel: " + err.message);
    } finally {
      setCancelling(false);
    }
  }

  if (!jobId) {
    return (
      <div className="rounded-lg border border-[var(--border)] bg-[var(--surface)] p-8 text-center text-sm text-[var(--text-muted)]">
        Select a job to see details
      </div>
    );
  }

  if (!job) {
    return (
      <div className="rounded-lg border border-[var(--border)] bg-[var(--surface)] p-8 text-center text-sm text-[var(--text-muted)]">
        Loading…
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-[var(--border)] bg-[var(--surface)] p-5 space-y-5">
      <div className="flex items-start justify-between">
        <div>
          <div className="text-sm font-medium">{job.Type}</div>
          <div className="font-mono text-xs text-[var(--text-muted)] mt-1">{job.ID}</div>
        </div>
        <div className="flex items-center gap-3">
          <StatusDot status={job.Status} />
          <button onClick={onClose} className="text-[var(--text-muted)] hover:text-[var(--text)] text-sm">
            ✕
          </button>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 text-sm">
        <div>
          <div className="text-xs text-[var(--text-muted)]">Attempts</div>
          <div className="font-mono mt-0.5">{job.Attempts} / {job.MaxAttempts}</div>
        </div>
        <div>
          <div className="text-xs text-[var(--text-muted)]">Created</div>
          <div className="font-mono mt-0.5">{formatDateTime(job.CreatedAt)}</div>
        </div>
      </div>

      {job.LastError && (
        <div>
          <div className="text-xs text-[var(--text-muted)] mb-1">Last error</div>
          <div className="font-mono text-xs text-[#F87171] bg-[#F87171]/10 rounded-md p-3 break-words">
            {job.LastError}
          </div>
        </div>
      )}

      {job.Result && (
        <div>
            <div className="text-xs text-[var(--text-muted)] mb-1">Result</div>
            <ResultView type={job.Type} result={job.Result} />
        </div>
        )}

      <div>
        <div className="text-xs text-[var(--text-muted)] mb-2">History</div>
        <div className="space-y-2">
          {events.map((ev) => (
            <div key={ev.ID} className="flex items-center gap-3 text-xs">
              <span className="font-mono text-[var(--text-muted)] w-32 shrink-0">
                {formatDateTime(ev.CreatedAt)}
              </span>
              <span className="capitalize">{ev.Event}</span>
              {ev.Detail && (
                <span className="text-[var(--text-muted)] truncate">— {ev.Detail}</span>
              )}
            </div>
          ))}
        </div>
      </div>

      {job.Status === "pending" && (
        <button
          onClick={handleCancel}
          disabled={cancelling}
          className="text-sm text-[#F87171] border border-[#F87171]/30 rounded-md px-3 py-1.5 hover:bg-[#F87171]/10 transition-colors disabled:opacity-40"
        >
          {cancelling ? "Cancelling…" : "Cancel job"}
        </button>
      )}
    </div>
  );
}