import React from "react"

export default function PageTitle({
    children,
    centered = true
}: {
    children: React.ReactNode,
    centered?: boolean
}) {
    return (
        <h1 className={`text-3xl font-semibold tracking-[-0.035em] sm:text-4xl ${centered ? "text-center" : ""}`}>
            {children}
        </h1>
    )
}

export function Subtitle({
    children,
    centered = true
}: {
    children: React.ReactNode,
    centered?: boolean
}) {
    return (
        <h2 className={`text-xl font-semibold tracking-[-0.025em] sm:text-2xl ${centered ? "text-center" : ""}`}>
            {children}
        </h2>
    )
}
