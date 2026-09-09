import type { Matrix } from '../types'

export type ParseMatrixResult = { matrix: Matrix } | { error: string }

const numberToken = '[+-]?(?:(?:\\d+(?:\\.\\d*)?)|(?:\\.\\d+))(?:[eE][+-]?\\d+)?'
const rowSyntax = new RegExp(`^${numberToken}(?:(?:\\s*,\\s*|\\s+)${numberToken})*$`)

export function parseMatrix(input: string): ParseMatrixResult {
  const trimmedInput = input.trim()

  if (trimmedInput.length === 0) {
    return { error: 'Enter at least one matrix row.' }
  }

  const rows: Matrix = []
  const lines = trimmedInput.split(/\r?\n/)

  for (const [index, line] of lines.entries()) {
    const rowNumber = index + 1
    const trimmedLine = line.trim()

    if (trimmedLine.length === 0) {
      return { error: `Row ${rowNumber} is empty.` }
    }

    if (!rowSyntax.test(trimmedLine)) {
      return { error: `Row ${rowNumber} contains an invalid number.` }
    }

    const row = trimmedLine.split(/\s*,\s*|\s+/).map(Number)
    if (row.some((value) => !Number.isFinite(value))) {
      return { error: `Row ${rowNumber} contains an invalid number.` }
    }

    rows.push(row)
  }

  const columnCount = rows[0].length
  if (rows.some((row) => row.length !== columnCount)) {
    return { error: 'All rows must contain the same number of values.' }
  }

  if (columnCount > rows.length) {
    return { error: 'Enter at least as many rows as columns.' }
  }

  return { matrix: rows }
}
