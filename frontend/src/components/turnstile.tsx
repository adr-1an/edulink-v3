"use client"

import Script from "next/script"
import {useEffect, useRef, useState} from "react"

interface TurnstileAPI {
    render: (
        container: HTMLElement,
        options: {
            sitekey: string
            action?: string
            callback?: (token: string) => void
            "expired-callback"?: () => void
            "error-callback"?: () => void
        },
    ) => string
    remove: (widgetID: string) => void
}

declare global {
    interface Window {
        turnstile?: TurnstileAPI
    }
}

export const turnstileSiteKey = process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY ?? ""

export function useTurnstile() {
    const [token, setToken] = useState("")
    const [widgetKey, setWidgetKey] = useState(0)

    function reset() {
        setToken("")
        setWidgetKey((current) => current + 1)
    }

    return {
        configured: Boolean(turnstileSiteKey),
        reset,
        token,
        widgetKey,
        setToken,
    }
}

export function TurnstileWidget({action, siteKey = turnstileSiteKey, onTokenChange}: {
    action: string
    siteKey?: string
    onTokenChange: (token: string) => void
}) {
    const containerRef = useRef<HTMLDivElement | null>(null)
    const widgetIDRef = useRef<string | null>(null)
    const [scriptReady, setScriptReady] = useState(false)

    useEffect(() => {
        if (!scriptReady || !siteKey || !containerRef.current || !window.turnstile || widgetIDRef.current) return

        widgetIDRef.current = window.turnstile.render(containerRef.current, {
            sitekey: siteKey,
            action,
            callback: onTokenChange,
            "expired-callback": () => onTokenChange(""),
            "error-callback": () => onTokenChange(""),
        })

        return () => {
            if (widgetIDRef.current && window.turnstile) {
                window.turnstile.remove(widgetIDRef.current)
                widgetIDRef.current = null
            }
        }
    }, [action, onTokenChange, scriptReady, siteKey])

    return (
        <>
            <Script
                src="https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit"
                strategy="afterInteractive"
                onReady={() => setScriptReady(true)}
            />
            <div ref={containerRef} className="min-h-16" />
        </>
    )
}
