import React from "react"

export default function PageTitle({
    children,
    centered = true
}: {
    children: React.ReactNode,
    centered?: boolean
}) {
    return (
        <h1 className={`text-3xl font-semibold tracking-wide ${centered && "text-center"}`}>
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
        <h1 className={`text-xl font-semibold tracking-wide text-center ${centered && "text-center"}`}>
            {children}
        </h1>
    )
}