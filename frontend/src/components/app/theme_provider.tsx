"use client"

import Cookies from "js-cookie"
import React, {createContext, useCallback, useContext, useMemo, useState} from "react"

export type Theme = "light" | "dark"

interface ThemeContextValue {
    theme: Theme
    setTheme: (theme: Theme) => void
    toggleTheme: () => void
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

export function ThemeProvider({initialTheme, children}: {
    initialTheme: Theme
    children: React.ReactNode
}) {
    const [theme, setThemeState] = useState<Theme>(initialTheme)

    const setTheme = useCallback((nextTheme: Theme) => {
        document.documentElement.classList.toggle("dark", nextTheme === "dark")
        Cookies.set("theme", nextTheme, {expires: 365, path: "/", sameSite: "lax"})
        setThemeState(nextTheme)
    }, [])

    const toggleTheme = useCallback(() => {
        setTheme(theme === "dark" ? "light" : "dark")
    }, [setTheme, theme])

    const value = useMemo(() => ({theme, setTheme, toggleTheme}), [setTheme, theme, toggleTheme])

    return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTheme() {
    const context = useContext(ThemeContext)
    if (!context) throw new Error("useTheme must be used within ThemeProvider")
    return context
}
