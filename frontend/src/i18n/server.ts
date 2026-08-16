import "server-only"

import {cookies} from "next/headers"
import {defaultLocale, isLocale, localeCookieName, type Locale} from "./config"
import {translate} from "./messages"

export async function getLocale(): Promise<Locale> {
    const savedLocale = (await cookies()).get(localeCookieName)?.value
    return isLocale(savedLocale) ? savedLocale : defaultLocale
}

export async function getTranslations() {
    const locale = await getLocale()
    return {
        locale,
        t: (key: Parameters<typeof translate>[1], values?: Parameters<typeof translate>[2]) =>
            translate(locale, key, values),
    }
}
