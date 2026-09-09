import type { StatisticsData } from '../types'

interface StatisticsProps {
  statistics: StatisticsData
  formatNumber: (value: number) => string
}

export function Statistics({ statistics, formatNumber }: StatisticsProps) {
  const entries = [
    ['Maximum', formatNumber(statistics.maximum)],
    ['Minimum', formatNumber(statistics.minimum)],
    ['Average', formatNumber(statistics.average)],
    ['Total sum', formatNumber(statistics.sum)],
    ['Diagonal matrix', statistics.hasDiagonalMatrix ? 'Yes' : 'No'],
  ]

  return (
    <dl className="statistics-grid">
      {entries.map(([label, value]) => (
        <div className="statistic-card" key={label}>
          <dt>{label}</dt>
          <dd>{value}</dd>
        </div>
      ))}
    </dl>
  )
}
