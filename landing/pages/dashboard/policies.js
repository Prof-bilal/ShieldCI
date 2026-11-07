import Head from 'next/head'
import Link from 'next/link'
import { useEffect, useState } from 'react'
import { ArrowLeft, CheckCircle2, LockKeyhole } from 'lucide-react'
const policies = [['production-policy', '3 repositories', 'Updated today', 'enforce'], ['observe-default', '1 repository', 'Updated yesterday', 'observe'], ['release-protection', '2 repositories', 'Updated Oct 31', 'enforce']]
export default function Policies() {
  const [livePolicies, setLivePolicies] = useState(null)
  useEffect(() => {
    const endpoint = process.env.NEXT_PUBLIC_SHIELDCI_API
    if (!endpoint) return
    fetch(`${endpoint.replace(/\/$/, '')}/api/v1/policies`).then((response) => response.ok ? response.json() : Promise.reject(new Error('policies request failed'))).then((data) => setLivePolicies((data.policies || []).map((record) => [record.policy.name, 'Connected policy', new Date(record.created_at).toLocaleDateString(), 'enforce']))).catch(() => setLivePolicies(null))
  }, [])
  const displayedPolicies = livePolicies || policies
  return <><Head><title>ShieldCI — Policies</title></Head><main className="dashboard-subpage"><Link className="dashboard-back" href="/dashboard"><ArrowLeft size={15} /> Back to overview</Link><div className="subpage-header"><div><span className="dashboard-eyebrow">Control plane / Policies</span><h1>Access policies</h1><p className="dashboard-subtitle">The explicit boundary each workflow must earn before it runs.</p></div><button className="dashboard-button"><LockKeyhole size={15} /> New policy</button></div><div className="subpage-table"><div className="subpage-table-head"><span>Policy</span><span>Coverage</span><span>Last changed</span><span>Mode</span></div>{displayedPolicies.map(([name, coverage, changed, mode]) => <div className="subpage-row" key={name}><b>{name}</b><span>{coverage}</span><span>{changed}</span><span className={`policy-mode ${mode}`}>{mode === 'enforce' && <CheckCircle2 size={13} />} {mode}</span></div>)}</div></main></>
}
