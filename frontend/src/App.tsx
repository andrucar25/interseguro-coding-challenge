import { useRef, useState } from 'react'
import { calculateQr } from './api/calculateQr'
import { MatrixInput } from './components/MatrixInput'
import { MatrixTable } from './components/MatrixTable'
import { Statistics } from './components/Statistics'
import { parseMatrix } from './lib/parseMatrix'
import type { QrResponse } from './types'

const exampleMatrix = '12 -51 4\n6 167 -68\n-4 24 -41'

function formatNumber(value: number): string {
  const magnitude = Math.abs(value)
  if (magnitude < 0.0000005) {
    return '0'
  }

  if (magnitude >= 1_000_000_000_000 || magnitude < 0.000001) {
    return value
      .toExponential(6)
      .replace(/(\.\d*?[1-9])0+e/, '$1e')
      .replace(/\.0+e/, 'e')
  }

  return value.toFixed(6).replace(/\.?0+$/, '')
}

export default function App() {
  const [input, setInput] = useState('')
  const [result, setResult] = useState<QrResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const isRequestInFlight = useRef(false)

  async function handleSubmit() {
    if (isSubmitting || isRequestInFlight.current) {
      return
    }

    setError(null)
    const parsed = parseMatrix(input)
    if ('error' in parsed) {
      setError(parsed.error)
      return
    }

    isRequestInFlight.current = true
    setIsSubmitting(true)
    try {
      const nextResult = await calculateQr(parsed.matrix)
      setResult(nextResult)
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'The request could not be completed. Please try again.')
    } finally {
      isRequestInFlight.current = false
      setIsSubmitting(false)
    }
  }

  return (
    <main className="page-shell">
      <section className="calculator-card" aria-labelledby="page-title">
        <header className="page-header">
          <p className="eyebrow">Matrix calculation</p>
          <h1 id="page-title">QR Matrix Calculator</h1>
          <p>Enter a rectangular matrix to calculate its economic QR factorization and summary statistics.</p>
        </header>

        <MatrixInput
          value={input}
          error={error}
          isSubmitting={isSubmitting}
          onChange={setInput}
          onSubmit={handleSubmit}
          onLoadExample={() => {
            setInput(exampleMatrix)
            setError(null)
          }}
        />

        {result ? (
          <section className="results-section" aria-labelledby="results-title">
            <h2 id="results-title">Results</h2>
            <div className="result-block">
              <h3>Q Matrix</h3>
              <MatrixTable label="Q" matrix={result.q} formatNumber={formatNumber} />
            </div>
            <div className="result-block">
              <h3>R Matrix</h3>
              <MatrixTable label="R" matrix={result.r} formatNumber={formatNumber} />
            </div>
            <div className="result-block">
              <h3>Statistics</h3>
              <Statistics statistics={result.statistics} formatNumber={formatNumber} />
            </div>
          </section>
        ) : null}
      </section>
      <footer className="challenge-footer">Coding Challenge Interseguro · Developed by Andrés De la Barra Vásquez</footer>
    </main>
  )
}
