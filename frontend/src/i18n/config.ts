export const supportedLocales = ["en", "pl"] as const

export type Locale = (typeof supportedLocales)[number]

export const defaultLocale: Locale = "en"
export const localeCookieName = "edulink_locale"

export function isLocale(value: unknown): value is Locale {
    return typeof value === "string" && supportedLocales.includes(value as Locale)
}

export function pluralCategory(locale: Locale, count: number) {
    if (!Number.isInteger(count)) return "other"
    if (locale === "en") return count === 1 ? "one" : "other"
    if (count === 1) return "one"

    const absolute = Math.abs(count)
    const lastDigit = absolute % 10
    const lastTwoDigits = absolute % 100
    return lastDigit >= 2 && lastDigit <= 4 && (lastTwoDigits < 12 || lastTwoDigits > 14)
        ? "few"
        : "many"
}
