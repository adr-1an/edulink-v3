import {type Locale} from "../config"
import {en, type MessageKey} from "./en"
import {pl} from "./pl"

export const messages: Record<Locale, Record<MessageKey, string>> = {en, pl}

export function translate(locale: Locale, key: MessageKey, values?: Record<string, string | number>) {
    const template = messages[locale][key]
    if (!values) return template

    return template.replace(/\{(\w+)\}/g, (placeholder, name: string) => {
        const value = values[name]
        return value === undefined ? placeholder : String(value)
    })
}

export type {MessageKey}
