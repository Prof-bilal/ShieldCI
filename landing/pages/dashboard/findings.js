import Head from 'next/head'
import Link from 'next/link'
import { useEffect, useState } from 'react'
import { ArrowLeft, ShieldAlert } from 'lucide-react'

const findings = [['aws-access-key', 'High', 'billing-service/.env.example', 'line 12', '8m ago'], ['python.lang.security', 'Medium', 'checkout-api/src/checkout.py', 'line 84', '42m ago'], ['private-key', 'Critical', 'web-console/scripts/deploy.sh', 'line 4', '1h ago']]

export default function Findings() {
  const [liveFindings, setLiveFindings] = useState(null)
  useEffect(() => {
    const endpoint = process.env.NEXT_PUBLIC_SHIELDCI_API
    if (!endpoint) return
    fetch(`${endpoint.replace(/\/$/, '')}/api/v1/findings`).then((response) => response.ok ? response.json() : Promise.reject(new Error('findings request failed'))).then((data) => setLiveFindings((data.findings || []).map((finding) => [finding.rule, finding.severity, finding.path, `line ${finding.line}`, 'live']))).catch(() => setLiveFindings(null))
  }, [])
  const displayedFindings = liveFindings || findings
  return <><Head><title>ShieldCI — Findings</title></Head><main className="dashboard-subpage"><Link className="dashboard-back" href="/dashboard"><ArrowLeft size={15} /> Back to overview</Link><div className="subpage-header"><div><span className="dashboard-eyebrow">Control plane / Findings</span><h1>Security findings</h1><p className="dashboard-subtitle">Scanner evidence linked back to the run that discovered it.</p></div><button className="dashboard-button"><ShieldAlert size={15} /> Run a scan</button></div><div className="subpage-table"><div className="subpage-table-head"><span>Rule</span><span>Severity</span><span>Location</span><span>Detected</span></div>{displayedFindings.map(([rule, severity, path, line, detected]) => <div className="subpage-row" key={`${rule}-${path}`}><b>{rule}</b><span className={`finding-severity ${severity.toLowerCase()}`}>{severity}</span><span>{path} · {line}</span><span>{detected}</span></div>)}</div></main></>
}
