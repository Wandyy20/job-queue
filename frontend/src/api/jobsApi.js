const BASE_URL = "http://localhost:8080"

export async function createJob(type, payload) {
    const res = await fetch(`${BASE_URL}/jobs`, {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({ type, payload }),
    });
    if (!res.ok) throw new Error("Failed to create job")
        return res.json()
}

export async function getJob(id) {
    const res = await fetch(`${BASE_URL}/jobs/${id}`);
    if (!res.ok) throw new Error("Failed to get job");
    return res.json();
}

export async function getJobEvents(id) {
  const res = await fetch(`${BASE_URL}/jobs/${id}/events`);
  if (!res.ok) throw new Error("Failed to get job events");
  return res.json();
}

export async function listJobs(status) {
  const url = status ? `${BASE_URL}/jobs?status=${status}` : `${BASE_URL}/jobs`;
  const res = await fetch(url);
  if (!res.ok) throw new Error("Failed to list jobs");
  return res.json();
}

export async function cancelJob(id) {
    const res = await fetch(`${BASE_URL}/jobs/${id}`, {method: "DELETE"});
    if(!res.ok) throw new Error("Failed to cancel job")
}

