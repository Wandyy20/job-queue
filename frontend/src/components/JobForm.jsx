import { useState } from "react";
import { createJob } from "../api/jobsApi";
import useToast from "../hooks/useToast";

const JOB_TYPES = [
  { value: "test_job", icon: "○", fields: ["text"] },
  { value: "send_webhook", icon: "↗", fields: ["url", "data"] },
  { value: "summarize_text", icon: "≡", fields: ["text"] },
  { value: "classify_sentiment", icon: "◐", fields: ["text"] },
  { value: "translate_text", icon: "⇄", fields: ["text", "language"] },
  { value: "resize_image", icon: "▢", fields: ["image_base64", "width", "height"] },
  { value: "generate_pdf_report", icon: "▤", fields: ["title", "items"] },
  { value: "csv_export", icon: "▦", fields: ["headers", "rows"] },
  { value: "classify_toxic_comment", icon: "⚠", fields: ["text"] },
];

const FIELD_LABELS = {
  text: "Text",
  url: "Webhook URL",
  data: "Data (JSON)",
  language: "Target language",
  image_base64: "Image (base64)",
  width: "Width (px)",
  height: "Height (px)",
  title: "Title",
  items: "Items (one per line)",
  headers: "Headers (comma separated)",
  rows: "Rows (one per line, comma separated)",
};

export default function JobForm({ onJobCreated }) {
  const [type, setType] = useState("summarize_text");
  const [values, setValues] = useState({});
  const [submitting, setSubmitting] = useState(false);
  const showToast = useToast();

  const activeType = JOB_TYPES.find((t) => t.value === type);

  function setField(field, value) {
    setValues((v) => ({ ...v, [field]: value }));
  }

  function buildPayload() {
    switch (type) {
      case "send_webhook":
        return { url: values.url || "", data: safeJsonParse(values.data) };
      case "translate_text":
        return { text: values.text || "", language: values.language || "" };
      case "resize_image":
        return {
          image_base64: values.image_base64 || "",
          width: Number(values.width) || 0,
          height: Number(values.height) || 0,
        };
      case "generate_pdf_report":
        return {
          title: values.title || "",
          items: (values.items || "").split("\n").filter(Boolean),
        };
      case "csv_export": {
        const headers = (values.headers || "").split(",").map((s) => s.trim()).filter(Boolean);
        const rows = (values.rows || "")
          .split("\n")
          .filter(Boolean)
          .map((line) => line.split(",").map((s) => s.trim()));
        return { headers, rows };
      }
      default:
        return { text: values.text || "" };
    }
  }

  function safeJsonParse(str) {
    try {
      return JSON.parse(str || "{}");
    } catch {
      return {};
    }
  }

  async function handleSubmit(e) {
    e.preventDefault();

    const emptyField = activeType.fields.find((f) => !values[f] || !values[f].trim());
    if (emptyField) {
      showToast(`${FIELD_LABELS[emptyField] || emptyField} is required`, "error");
      return;
    }

    setSubmitting(true);
    try {
      await createJob(type, buildPayload());
      setValues({});
      showToast("Job submitted successfully");
      onJobCreated?.();
    } catch (err) {
      showToast("Failed to submit job: " + err.message, "error");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="rounded-lg border border-[var(--border)] bg-[var(--surface)] p-5 space-y-4"
    >
      <h2 className="text-sm font-medium text-[var(--text-muted)] uppercase tracking-wide">
        New Job
      </h2>

      <div>
        <label className="text-xs text-[var(--text-muted)] block mb-2">Job type</label>
        <div className="flex flex-wrap gap-1.5">
          {JOB_TYPES.map((t) => (
            <button
              key={t.value}
              type="button"
              onClick={() => {
                setType(t.value);
                setValues({});
              }}
              className={`px-3 py-1.5 rounded-md text-xs font-mono flex items-center gap-1.5 border transition-colors ${
                type === t.value
                  ? "border-[var(--accent)] bg-[var(--accent)]/10 text-[var(--accent)]"
                  : "border-[var(--border)] text-[var(--text-muted)] hover:border-[var(--text-muted)]"
              }`}
            >
              <span>{t.icon}</span>
              {t.value}
            </button>
          ))}
        </div>
      </div>

      <div className="space-y-3">
        {activeType.fields.map((field) => (
          <div key={field}>
            <label className="text-xs text-[var(--text-muted)] block mb-1.5">
              {FIELD_LABELS[field] || field}
            </label>
            {["text", "data", "items", "rows"].includes(field) ? (
              <textarea
                value={values[field] || ""}
                onChange={(e) => setField(field, e.target.value)}
                rows={3}
                className="w-full rounded-md border border-[var(--border)] bg-[var(--bg)] p-3 text-sm font-mono placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] resize-none"
              />
            ) : (
              <input
                type={["width", "height"].includes(field) ? "number" : "text"}
                value={values[field] || ""}
                onChange={(e) => setField(field, e.target.value)}
                className="w-full rounded-md border border-[var(--border)] bg-[var(--bg)] p-2.5 text-sm font-mono placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)]"
              />
            )}
          </div>
        ))}
      </div>

      <button
        type="submit"
        disabled={submitting}
        className="bg-[var(--accent)] text-[#0A0A0F] font-medium text-sm px-4 py-2 rounded-md hover:opacity-90 disabled:opacity-40 transition-opacity"
      >
        {submitting ? "Submitting…" : "Submit job"}
      </button>
    </form>
  );
}