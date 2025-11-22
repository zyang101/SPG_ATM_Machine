"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

export default function WaitingForVerification({ token, returnTo }: { token: string; returnTo?: string }) {
  const [status, setStatus] = useState("pending");
  const router = useRouter();

  useEffect(() => {
    if (!token) return;
    let stopped = false;
    const interval = setInterval(async () => {
      try {
        const url = `${process.env.NEXT_PUBLIC_API_URL || ""}/api/guest/status?token=${token}`;
        const resp = await fetch(url);
        if (!resp.ok) return;
        const j = await resp.json();
        if (j.status === "approved" || j.status === "consumed") {
          if (j.authToken) {
            localStorage.setItem("authToken", j.authToken);
          }
          clearInterval(interval);
          if (!stopped) router.push(returnTo || "/");
        } else if (j.status === "denied") {
          clearInterval(interval);
          if (!stopped) router.push("/sign-in");
        } else {
          setStatus(j.status);
        }
      } catch (e) {
        // ignore transient errors
      }
    }, 3000);
    return () => {
      stopped = true;
      clearInterval(interval);
    };
  }, [token, router, returnTo]);

  return (
    <div style={{ padding: 20 }}>
      <h2>Waiting for homeowner approval</h2>
      <p>Status: {status}</p>
      <p>An email has been sent to the homeowner — please wait or cancel to return to sign-in.</p>
      <button onClick={() => router.push("/sign-in")}>Cancel</button>
    </div>
  );
}