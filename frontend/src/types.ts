/// <reference types="vite/client" />

export type Matrix = number[][]

export interface StatisticsData {
  maximum: number
  minimum: number
  sum: number
  average: number
  hasDiagonalMatrix: boolean
}

export interface QrResponse {
  q: Matrix
  r: Matrix
  statistics: StatisticsData
}
