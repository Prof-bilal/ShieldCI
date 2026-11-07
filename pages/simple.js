import Head from 'next/head'
import { CheckCircle } from 'lucide-react'

export default function Simple() {
  return (
    <div>
      <Head>
        <title>Simple Test</title>
      </Head>
      <h1>Simple Page</h1>
      <CheckCircle className="w-8 h-8" />
    </div>
  )
}
