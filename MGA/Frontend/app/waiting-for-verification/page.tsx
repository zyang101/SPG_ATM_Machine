"use client";
import { useSearchParams } from "next/navigation";
import WaitingForVerification from "../../components/waiting-for-verification";

export default function Page() {
  const params = useSearchParams();
  const token = params.get("token") || "";
  const returnTo = params.get("returnTo") || "/";
  return <WaitingForVerification token={token} returnTo={returnTo} />;
}