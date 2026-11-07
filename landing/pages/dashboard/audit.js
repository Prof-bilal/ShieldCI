import Head from 'next/head'
import Link from 'next/link'
import { useEffect, useState } from 'react'
import { ArrowLeft, CheckCircle2, ShieldAlert } from 'lucide-react'
const events = [['Outbound request blocked', 'billing-service · registry.example.com', '8m ago', 'blocked'], ['Policy updated', 'checkout-api · production-policy', '42m ago', 'allowed'], ['Secret access denied', 'web-console · AWS_SECRET_KEY', '1h ago', 'blocked']]
export default function Audit() {
  const [liveEvents, setLiveEvents] = useState(null)
  useEffect(() => {
    const endpoint = process.env.NEXT_PUBLIC_SHIELDCI_API
    if (!endpoint) return
    const base = endpoint.replace(/\/$/, '')
    fetch(`${base}/api/v1/jobs`).then((response) => response.ok ? response.json() : Promise.reject(new Error('jobs request failed'))).then((data) => data.jobs?.[0]?.id ? fetch(`${base}/api/v1/jobs/${data.jobs[0].id}/audit`) : { ok: true, json: () => Promise.resolve({ events: [] }) }).then((response) => response.ok ? response.json() : Promise.reject(new Error('audit request failed'))).then((data) => setLiveEvents((data.events || []).map((event) => [event.type, event.reason || event.decision, new Date(event.created_at).toLocaleString(), event.decision === 'allowed' ? 'allowed' : 'blocked']))).catch(() => setLiveEvents(null))
  }, [])
  const displayedEvents = liveEvents || events
  return <><Head><title>ShieldCI — Audit log</title></Head><main className="dashboard-subpage"><Link className="dashboard-back" href="/dashboard"><ArrowLeft size={15} /> Back to overview</Link><div className="subpage-header"><div><span className="dashboard-eyebrow">Control plane / Audit</span><h1>Audit log</h1><p className="dashboard-subtitle">A reviewable record of every policy decision.</p></div></div><div className="audit-feed">{displayedEvents.map(([title, detail, time, decision]) => <article className="audit-feed-row" key={`${title}-${time}`}><span className={`event-marker ${decision === 'blocked' ? 'warning' : 'success'}`}>{decision === 'blocked' ? <ShieldAlert size={15} /> : <CheckCircle2 size={15} />}</span><div><b>{title}</b><p>{detail}</p></div><time>{time}</time></article>)}</div></main></>
}
