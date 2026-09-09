interface MatrixInputProps {
  value: string
  error: string | null
  isSubmitting: boolean
  onChange: (value: string) => void
  onLoadExample: () => void
  onSubmit: () => void
}

export function MatrixInput({ value, error, isSubmitting, onChange, onLoadExample, onSubmit }: MatrixInputProps) {
  const describedBy = error ? 'matrix-help matrix-error' : 'matrix-help'

  return (
    <form
      className="matrix-form"
      onSubmit={(event) => {
        event.preventDefault()
        onSubmit()
      }}
    >
      <label htmlFor="matrix">Matrix values</label>
      <p id="matrix-help" className="field-help">
        One row per line. Separate values with spaces or commas.
        <br />
        12 -51 4
        <br />
        6 167 -68
        <br />
        -4 24 -41
      </p>
      <textarea
        id="matrix"
        name="matrix"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        aria-describedby={describedBy}
        aria-invalid={Boolean(error)}
        placeholder="1 2\n3 4"
        disabled={isSubmitting}
        rows={8}
        spellCheck={false}
      />
      {error ? (
        <p id="matrix-error" className="form-error" role="alert">
          {error}
        </p>
      ) : null}
      <div className="form-actions">
        <button type="button" className="secondary-button" onClick={onLoadExample} disabled={isSubmitting}>
          Load example
        </button>
        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Calculating...' : 'Calculate QR'}
        </button>
      </div>
      {isSubmitting ? (
        <p className="loading-message" role="status" aria-live="polite">
          Calculating the QR factorization.
        </p>
      ) : null}
    </form>
  )
}
