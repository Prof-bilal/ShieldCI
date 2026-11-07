import Head from 'next/head'
import { useEffect, useState } from 'react'
import { Activity, ArrowUpRight, CheckCircle2, ChevronDown, GitBranch, LockKeyhole, MoreHorizontal, ShieldAlert, Terminal, XCircle } from 'lucide-react'

const runs = [
  { repo: 'checkout-api', branch: 'feature/checkout-timeout', workflow: 'Pull request checks', status: 'passed', time: '2m 14s', commit: '5e8c1a' },
  { repo: 'billing-service', branch: 'main', workflow: 'Release protection', status: 'blocked', time: '41s', commit: 'a4d021' },
  { repo: 'web-console', branch: 'main', workflow: 'Security checks', status: 'passed', time: '1m 32s', commit: '91f72c' },
  { repo: 'checkout-api', branch: 'main', workflow: 'Dependency review', status: 'failed', time: '58s', commit: '7b2e10' },
]

function BrandMark() { return <span className="brand-mark" aria-hidden="true"><span /><span /></span> }

function Status({ value }) {
  const config = {
    passed: ['status-success', CheckCircle2, 'Passed'],
    blocked: ['status-blocked', ShieldAlert, 'Blocked'],
    failed: ['status-failed', XCircle, 'Failed'],
  }[value]
  const [className, Icon, label] = config
  return <span className={`dashboard-status ${className}`}><Icon size={14} /> {label}</span>
}

export default function Dashboard() {
  const [liveRuns, setLiveRuns] = useState(null)
  useEffect(() => {
    const endpoint = process.env.NEXT_PUBLIC_SHIELDCI_API
    if (!endpoint) return
    fetch(`${endpoint.replace(/\/$/, '')}/api/v1/jobs`)
      .then((response) => response.ok ? response.json() : Promise.reject(new Error('jobs request failed')))
      .then((data) => setLiveRuns((data.jobs || []).map((job) => ({ repo: 'ShieldCI job', branch: job.policy_id, workflow: job.command?.join(' ') || 'Command run', status: ['queued', 'running'].includes(job.status) ? 'passed' : job.status, time: '—', commit: job.id.slice(-6) }))))
      .catch(() => setLiveRuns(null))
  }, [])
  const displayedRuns = liveRuns || runs
  return <div className="dashboard-shell">
    <Head><title>ShieldCI — Dashboard</title><meta name="description" content="ShieldCI security control plane dashboard." /></Head>
    <aside className="dashboard-sidebar">
      <a className="brand dashboard-brand" href="/"><BrandMark /><span>Shield<span className="brand-accent">CI</span></span></a>
      <div className="workspace-switcher"><span className="workspace-avatar">A</span><span><b>Acme security</b><small>1 workspace</small></span><ChevronDown size={15} /></div>
      <nav className="dashboard-nav" aria-label="Dashboard navigation">
        <a className="active" href="/dashboard"><Activity size={17} /> Overview</a><a href="/dashboard/jobs"><Terminal size={17} /> Runs <span>12</span></a><a href="/dashboard/policies"><LockKeyhole size={17} /> Policies</a><a href="/dashboard/findings"><ShieldAlert size={17} /> Findings <span className="nav-alert">3</span></a>
      </nav>
      <div className="sidebar-footer"><a href="#settings">Settings</a><a href="/">Back to site</a><span>ShieldCI beta</span></div>
    </aside>
    <main className="dashboard-main">
      <header className="dashboard-header"><div><p className="dashboard-eyebrow">Control plane / Overview</p><h1>Good morning, Bilal.</h1><p className="dashboard-subtitle">Here’s what your pipelines are telling you today.</p></div><button className="dashboard-button"><GitBranch size={15} /> Connect repository <ArrowUpRight size={15} /></button></header>
      <section className="metric-grid" aria-label="Workspace metrics"><article><span className="metric-label">Runs this week</span><strong>128</strong><small className="metric-positive">+18% from last week</small></article><article><span className="metric-label">Policy pass rate</span><strong>94.5%</strong><small className="metric-positive">+2.1% from last week</small></article><article><span className="metric-label">Blocked actions</span><strong>07</strong><small className="metric-warning">3 need review</small></article><article><span className="metric-label">Open findings</span><strong>03</strong><small className="metric-warning">1 high severity</small></article></section>
      <section className="dashboard-grid"><div className="dashboard-panel runs-panel" id="runs"><div className="panel-heading"><div><span className="panel-kicker">{liveRuns ? 'Live API activity' : 'Latest activity'}</span><h2>Recent runs</h2></div><a href="#all-runs">View all <ArrowUpRight size={14} /></a></div><div className="run-table"><div className="run-table-heading"><span>Repository</span><span>Workflow</span><span>Status</span><span>Duration</span><span /></div>{displayedRuns.map((run) => <div className="run-table-row" key={`${run.repo}-${run.commit}`}><div className="repo-cell"><span className="repo-avatar">{run.repo[0].toUpperCase()}</span><span><b>{run.repo}</b><small><GitBranch size={11} /> {run.branch}</small></span></div><div><b>{run.workflow}</b><small>commit {run.commit}</small></div><Status value={run.status} /><span className="run-duration">{run.time}</span><button className="icon-button" aria-label={`More options for ${run.repo}`}><MoreHorizontal size={17} /></button></div>)}</div></div><div className="dashboard-panel activity-panel"><div className="panel-heading"><div><span className="panel-kicker">Policy events</span><h2>What was stopped</h2></div><a href="#audit">Audit log <ArrowUpRight size={14} /></a></div><div className="event-list"><div className="dashboard-event"><span className="event-marker warning"><ShieldAlert size={15} /></span><div><b>Outbound request blocked</b><p>billing-service · registry.example.com</p></div><time>8m ago</time></div><div className="dashboard-event"><span className="event-marker success"><CheckCircle2 size={15} /></span><div><b>Policy updated</b><p>checkout-api · production-policy</p></div><time>42m ago</time></div><div className="dashboard-event"><span className="event-marker warning"><ShieldAlert size={15} /></span><div><b>Secret access denied</b><p>web-console · AWS_SECRET_KEY</p></div><time>1h ago</time></div></div></div></section>
      <section className="dashboard-panel dashboard-callout" id="policies"><div className="callout-icon"><LockKeyhole size={19} /></div><div><span className="panel-kicker">Policy coverage</span><h2>3 repositories are protected</h2><p>Your next step is to connect a runner for <b>billing-service</b>. Jobs stay blocked until a policy is attached.</p></div><button className="dashboard-button secondary">Review policies <ArrowUpRight size={15} /></button></section>
    </main>
  </div>
}
