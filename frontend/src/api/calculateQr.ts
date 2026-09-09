import type { Matrix, QrResponse, StatisticsData } from '../types'

const networkErrorMessage = 'The request could not be completed. Check your connection and try again.'
const unexpectedResponseMessage = 'The service returned an unexpected response. Please try again.'
const knownServerMessages = new Set([
  'unable to obtain statistics',
  'statistics service timed out',
  'unable to factorize matrix',
])

function getQrUrl(): string {
  const configuredUrl = import.meta.env.VITE_API_URL?.trim()

  if (!configuredUrl) {
    throw new Error('The API URL is not configured. Set VITE_API_URL and restart the app.')
  }

  try {
    const url = new URL(configuredUrl)
    if ((url.protocol !== 'http:' && url.protocol !== 'https:') || url.search || url.hash) {
      throw new Error('Invalid API URL')
    }
  } catch {
    throw new Error('The API URL must be an absolute HTTP(S) URL without a query or fragment. Check VITE_API_URL and restart the app.')
  }

  return `${configuredUrl.replace(/\/+$/, '')}/qr`
}

function isFiniteMatrix(value: unknown): value is Matrix {
  return (
    Array.isArray(value) &&
    value.length > 0 &&
    value.every(
      (row) =>
        Array.isArray(row) &&
        row.length > 0 &&
        row.every((cell) => typeof cell === 'number' && Number.isFinite(cell)) &&
        row.length === value[0].length,
    )
  )
}

function isStatistics(value: unknown): value is StatisticsData {
  if (!value || typeof value !== 'object') {
    return false
  }

  const statistics = value as Record<string, unknown>
  return (
    typeof statistics.maximum === 'number' &&
    Number.isFinite(statistics.maximum) &&
    typeof statistics.minimum === 'number' &&
    Number.isFinite(statistics.minimum) &&
    typeof statistics.sum === 'number' &&
    Number.isFinite(statistics.sum) &&
    typeof statistics.average === 'number' &&
    Number.isFinite(statistics.average) &&
    typeof statistics.hasDiagonalMatrix === 'boolean'
  )
}

function isQrResponse(value: unknown): value is QrResponse {
  if (!value || typeof value !== 'object') {
    return false
  }

  const response = value as Record<string, unknown>
  return isFiniteMatrix(response.q) && isFiniteMatrix(response.r) && isStatistics(response.statistics)
}

function getMessage(payload: unknown): string | null {
  if (!payload || typeof payload !== 'object') {
    return null
  }

  const message = (payload as Record<string, unknown>).message
  return typeof message === 'string' ? message : null
}

export async function calculateQr(matrix: Matrix): Promise<QrResponse> {
  const qrUrl = getQrUrl()
  let response: Response

  try {
    response = await fetch(qrUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ matrix }),
    })
  } catch {
    throw new Error(networkErrorMessage)
  }

  let payload: unknown
  try {
    payload = await response.json()
  } catch {
    if (response.ok) {
      throw new Error(unexpectedResponseMessage)
    }

    payload = null
  }

  if (!response.ok) {
    const message = getMessage(payload)
    if (response.status === 400 && message) {
      throw new Error(message)
    }

    if ((response.status === 500 || response.status === 502 || response.status === 504) && message && knownServerMessages.has(message)) {
      throw new Error(message)
    }

    if (response.status >= 400 && response.status < 500) {
      throw new Error('The request was rejected. Check the matrix and try again.')
    }

    throw new Error('The service could not complete the request. Please try again.')
  }

  if (!isQrResponse(payload)) {
    throw new Error(unexpectedResponseMessage)
  }

  return payload
}
