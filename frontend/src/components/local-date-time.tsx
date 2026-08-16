"use client"

import {useCallback, useSyncExternalStore} from "react"
import {useLocale} from "@/i18n/provider"

type DateTimePrecision = "date" | "minute" | "second"

const subscribe = () => () => {}

function useBrowserFormatting() {
    return useSyncExternalStore(subscribe, () => true, () => false)
}

export function useBrowserTimeZone() {
    return useBrowserFormatting()
}

function utcFallback(date: Date, precision: DateTimePrecision, locale: string) {
    const englishMonths = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]
    const polishMonths = ["sty", "lut", "mar", "kwi", "maj", "cze", "lip", "sie", "wrz", "paź", "lis", "gru"]
    const month = (locale === "pl" ? polishMonths : englishMonths)[date.getUTCMonth()]
    const day = date.getUTCDate()
    const year = date.getUTCFullYear()
    const datePart = locale === "pl" ? `${day} ${month} ${year}` : `${month} ${day}, ${year}`

    if (precision === "date") {
        return datePart
    }

    const pad = (value: number) => String(value).padStart(2, "0")
    const minutes = pad(date.getUTCMinutes())
    const seconds = precision === "second" ? `:${pad(date.getUTCSeconds())}` : ""

    if (locale === "pl") {
        return `${datePart}, ${pad(date.getUTCHours())}:${minutes}${seconds} UTC`
    }

    const hour = date.getUTCHours()
    const hour12 = hour % 12 || 12
    const period = hour < 12 ? "AM" : "PM"
    return `${datePart}, ${hour12}:${minutes}${seconds} ${period} UTC`
}

export function useLocalDateTimeFormatter() {
    const useBrowser = useBrowserFormatting()
    const {locale, t} = useLocale()

    return useCallback((value: string, precision: DateTimePrecision = "minute") => {
        const date = new Date(value)
        if (Number.isNaN(date.getTime())) return t("common.unknownDate")
        if (!useBrowser) return utcFallback(date, precision, locale)

        if (precision === "date") {
            return new Intl.DateTimeFormat(locale === "pl" ? "pl-PL" : "en-US", {
                day: "numeric",
                month: "short",
                year: "numeric",
            }).format(date)
        }

        return new Intl.DateTimeFormat(locale === "pl" ? "pl-PL" : "en-US", {
            day: "numeric",
            month: "short",
            year: "numeric",
            hour: "numeric",
            minute: "2-digit",
            second: precision === "second" ? "2-digit" : undefined,
            timeZoneName: "short",
        }).format(date)
    }, [locale, t, useBrowser])
}

export function useLocalDateKey(serverDate: string) {
    const useBrowser = useBrowserFormatting()
    if (!useBrowser) return serverDate

    const now = new Date()
    const pad = (value: number) => String(value).padStart(2, "0")
    return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
}

export default function LocalDateTime({
    value,
    precision = "minute",
    className,
}: {
    value: string
    precision?: DateTimePrecision
    className?: string
}) {
    const format = useLocalDateTimeFormatter()
    return <time className={className} dateTime={value}>{format(value, precision)}</time>
}
