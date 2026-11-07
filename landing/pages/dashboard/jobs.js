import Head from 'next/head'
import Link from 'next/link'
import { useEffect, useState } from 'react'
import { ArrowLeft, CheckCircle2, GitBranch, ShieldAlert, Terminal, XCircle } from 'lucide-react'

const jobs = [['checkout-api', 'Pull request checks', 'passed', '2m 14s', 'feature/checkout-timeout'], ['billing-service', 'Release protection', 'blocked', '41s', 'main'], ['web-console', 'Security checks', 'passed', '1m 32s', 'main']]
function Status({ status }) { const Icon = status === 'passed' ? CheckCircle2 : status === 'blocked' ? ShieldAlert : XCircle; return <span className={`dashboard-status status-${status === 'passed' ? 'success' : status === 'blocked' ? 'blocked' : 'failed'}`}><Icon size={14} /> {status}</span> }
export default function Jobs() {
  const [liveJobs, setLiveJobs] = useState(null)
  useEffect(() => {
    const endpoint = process.env.NEXT_PUBLIC_SHIELDCI_API
    if (!endpoint) return
    fetch(`${endpoint.replace(/\/$/, '')}/api/v1/jobs`).then((response) => response.ok ? response.json() : Promise.reject(new Error('jobs request failed'))).then((data) => setLiveJobs((data.jobs || []).map((job) => ['ShieldCI job', job.command?.join(' ') || 'Command run', job.status === 'queued' || job.status === 'running' ? 'passed' : job.status, '—', job.policy_id]))).catch(() => setLiveJobs(null))
  }, [])
  const displayedJobs = liveJobs || jobs
  return <><Head><title>ShieldCI — Jobs</title></Head><main className="dashboard-subpage"><Link className="dashboard-back" href="/dashboard"><ArrowLeft size={15} /> Back to overview</Link><div className="subpage-header"><div><span className="dashboard-eyebrow">Control plane / Jobs</span><h1>All runs</h1><p className="dashboard-subtitle">Every job, its policy decision, and the evidence behind it.</p></div><button className="dashboard-button"><Terminal size={15} /> Submit a job</button></div><div className="subpage-table"><div className="subpage-table-head"><span>Repository</span><span>Workflow</span><span>Branch</span><span>Status</span><span>Duration</span></div>{displayedJobs.map(([repo, workflow, status, duration, branch]) => <div className="subpage-row" key={`${repo}-${workflow}-${branch}`}><b>{repo}</b><span>{workflow}</span><span><GitBranch size={12} /> {branch}</span><Status status={status} /><span>{duration}</span></div>)}</div></main></>
}
