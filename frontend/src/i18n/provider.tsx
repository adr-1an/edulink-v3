"use client"

import {createContext, useContext, useMemo} from "react"
import {type Locale} from "./config"
import {translate, type MessageKey} from "./messages"

interface LocaleContextValue {
    locale: Locale
    t: (key: MessageKey, values?: Record<string, string | number>) => string
}

const LocaleContext = createContext<LocaleContextValue | null>(null)

export function LocaleProvider({children, locale}: {children: React.ReactNode; locale: Locale}) {
    const value = useMemo<LocaleContextValue>(() => ({
        locale,
        t: (key, values) => translate(locale, key, values),
    }), [locale])

    return <LocaleContext.Provider value={value}>{children}</LocaleContext.Provider>
}

export function useLocale() {
    const context = useContext(LocaleContext)
    if (!context) throw new Error("useLocale must be used within LocaleProvider")
    return context
}
