import type { Matrix } from '../types'

interface MatrixTableProps {
  label: string
  matrix: Matrix
  formatNumber: (value: number) => string
}

export function MatrixTable({ label, matrix, formatNumber }: MatrixTableProps) {
  return (
    <div className="matrix-table-scroll" tabIndex={0} aria-label={`${label} matrix table`}>
      <table className="matrix-table">
        <caption>{label} matrix</caption>
        <tbody>
          {matrix.map((row, rowIndex) => (
            <tr key={rowIndex}>
              <th scope="row">{rowIndex + 1}</th>
              {row.map((value, columnIndex) => (
                <td key={columnIndex}>{formatNumber(value)}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
