export function scoreAccent(score: number) {
    const normalizedScore = Math.min(100, Math.max(0, score))
    const hue = Math.round((normalizedScore / 100) * 120)
    return `hsl(${hue} 72% 42%)`
}
