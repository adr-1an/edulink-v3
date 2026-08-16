"use server"

import {cookies} from "next/headers"
import {isLocale, localeCookieName, type Locale} from "./config"

export async function setLocale(locale: Locale) {
    if (!isLocale(locale)) return {ok: false as const}

    const cookieStore = await cookies()
    cookieStore.set(localeCookieName, locale, {
        httpOnly: true,
        maxAge: 60 * 60 * 24 * 365,
        path: "/",
        sameSite: "lax",
        secure: process.env.NODE_ENV === "production",
    })

    return {ok: true as const}
}
