import { useState } from "react";
import JobForm from "./components/JobForm";
import JobList from "./components/JobList";
import JobDetail from "./components/JobDetail";

function App() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [selectedJobId, setSelectedJobId] = useState(null);

  function bump() {
    setRefreshTrigger((k) => k + 1);
  }

  return (
    <div className="min-h-screen bg-[var(--bg)]">
      <header className="border-b border-[var(--border)] px-6 py-8 text-center relative overflow-hidden">
        <div
          className="absolute inset-0 opacity-[0.07]"
          style={{
            backgroundImage:
              "radial-gradient(circle at 50% 0%, var(--accent) 0%, transparent 60%)",
          }}
        />
        <div className="relative">
          <div className="flex items-center justify-center gap-2 mb-2">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[var(--accent)] opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-[var(--accent)]" />
            </span>
            <span className="text-xs font-mono text-[var(--text-muted)] uppercase tracking-wider">
              Live
            </span>
          </div>
          <h1 className="text-3xl font-semibold tracking-tight">Job Queue</h1>
          <p className="text-sm text-[var(--text-muted)] mt-1">
            Concurrency-safe background job processing
          </p>
        </div>
      </header>

      <main className="max-w-6xl mx-auto p-6 grid grid-cols-1 lg:grid-cols-[1fr_380px] gap-6 items-start">
        <div className="space-y-6">
          <JobForm onJobCreated={bump} />
          <JobList
            refreshTrigger={refreshTrigger}
            selectedJobId={selectedJobId}
            onSelectJob={setSelectedJobId}
          />
        </div>
        <div className="lg:sticky lg:top-6">
          <JobDetail
            jobId={selectedJobId}
            onClose={() => setSelectedJobId(null)}
            onChanged={bump}
          />
        </div>
      </main>
    </div>
  );
}

export default App;