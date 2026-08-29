export function formatYuan(fen: number): string {
  return `¥${(fen / 100).toFixed(2)}`
}

export function remainingFen(lifetimePaidFen: number, thresholdFen: number): number {
  return Math.max(0, thresholdFen - lifetimePaidFen)
}