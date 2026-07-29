import { describe, it, expect } from 'vitest'
import {
  sanitizeHtml,
  isNumeric,
  toUnix,
  fromUnix,
  now,
  nowSubtract,
  humanTime,
  serviceLink,
  convertToChartData,
} from '../mixins'

describe('mixins', () => {
  describe('sanitizeHtml', () => {
    it('returns empty string for null/undefined', () => {
      expect(sanitizeHtml(null)).toBe('')
      expect(sanitizeHtml(undefined)).toBe('')
    })

    it('strips dangerous script tags', () => {
      const result = sanitizeHtml('<script>alert("xss")</script><p>Safe</p>')
      expect(result).not.toContain('script')
      expect(result).toContain('Safe')
    })

    it('allows safe HTML', () => {
      const result = sanitizeHtml('<p>Hello <strong>World</strong></p>')
      expect(result).toContain('Hello')
      expect(result).toContain('<strong>')
      expect(result).toContain('World')
    })
  })

  describe('isNumeric', () => {
    it('returns true for numbers', () => {
      expect(isNumeric(123)).toBe(true)
      expect(isNumeric('456')).toBe(true)
      expect(isNumeric(0)).toBe(true)
      expect(isNumeric('0')).toBe(true)
    })

    it('returns false for non-numbers', () => {
      expect(isNumeric('abc')).toBe(false)
      expect(isNumeric(null)).toBe(false)
      expect(isNumeric(undefined)).toBe(false)
      expect(isNumeric('')).toBe(false)
    })
  })

  describe('toUnix / fromUnix', () => {
    it('converts date to unix timestamp', () => {
      const date = new Date('2024-01-01T00:00:00Z')
      expect(toUnix(date)).toBe(1704067200)
    })

    it('converts unix timestamp to ISO string', () => {
      const result = fromUnix(1704067200)
      expect(result).toBe('2024-01-01T00:00:00.000Z')
    })
  })

  describe('now / nowSubtract', () => {
    it('returns current date', () => {
      const result = now()
      expect(result).toBeInstanceOf(Date)
    })

    it('subtracts seconds from now', () => {
      const before = Date.now()
      const result = nowSubtract(60)
      const after = Date.now()

      expect(result.getTime()).toBeGreaterThanOrEqual(before - 60000 - 100)
      expect(result.getTime()).toBeLessThanOrEqual(after - 60000 + 100)
    })
  })

  describe('humanTime', () => {
    it('returns 0ms for falsy values', () => {
      expect(humanTime(null)).toBe('0ms')
      expect(humanTime(0)).toBe('0ms')
    })

    it('formats milliseconds correctly', () => {
      expect(humanTime(500)).toBe('500 μs')
      expect(humanTime(15000)).toBe('15 ms')
    })
  })

  describe('serviceLink', () => {
    it('returns root for null service', () => {
      expect(serviceLink(null)).toBe('/')
    })

    it('uses permalink if available', () => {
      expect(serviceLink({ id: 1, permalink: 'my-service' })).toBe('/service/my-service')
    })

    it('falls back to id', () => {
      expect(serviceLink({ id: 42 })).toBe('/service/42')
    })
  })

  describe('convertToChartData', () => {
    it('returns empty data for invalid input', () => {
      expect(convertToChartData(null)).toEqual({ data: [] })
      expect(convertToChartData('invalid')).toEqual({ data: [] })
    })

    it('converts data points correctly', () => {
      const input = [
        { timeframe: '2024-01-01T00:00:00Z', amount: 100 },
        { timeframe: '2024-01-02T00:00:00Z', amount: 200 },
      ]
      const result = convertToChartData(input)

      expect(result.data).toHaveLength(2)
      expect(result.data[0].x).toBe(new Date('2024-01-01T00:00:00Z').getTime())
      expect(result.data[0].y).toBe(100)
    })
  })
})
