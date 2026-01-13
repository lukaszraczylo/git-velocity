// Number formatting thresholds
const ONE_MILLION = 1_000_000
const ONE_THOUSAND = 1_000

// Time conversion constants
const MINUTES_PER_HOUR = 60
const HOURS_PER_DAY = 24

/**
 * Format a number with K/M suffixes for large values
 */
export function formatNumber(n) {
  if (n === null || n === undefined) return '0'
  if (n >= ONE_MILLION) {
    return (n / ONE_MILLION).toFixed(1) + 'M'
  }
  if (n >= ONE_THOUSAND) {
    return (n / ONE_THOUSAND).toFixed(1) + 'K'
  }
  return String(n)
}

/**
 * Format hours as a human-readable duration
 */
export function formatDuration(hours) {
  if (hours === null || hours === undefined || hours <= 0) return '-'
  if (hours < 1) {
    return Math.round(hours * MINUTES_PER_HOUR) + 'm'
  }
  if (hours < HOURS_PER_DAY) {
    return hours.toFixed(1) + 'h'
  }
  return (hours / HOURS_PER_DAY).toFixed(1) + 'd'
}

/**
 * Format a date string or Date object
 */
export function formatDate(dateInput) {
  if (!dateInput) return ''
  const date = new Date(dateInput)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

/**
 * Format a number as a percentage
 */
export function formatPercent(value) {
  if (value === null || value === undefined) return '0%'
  return value.toFixed(1) + '%'
}

/**
 * Convert a string to a URL-friendly slug
 */
export function slugify(str) {
  if (!str) return ''
  return str
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(/_/g, '-')
    .replace(/[^a-z0-9-]/g, '')
}
