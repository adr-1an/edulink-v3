import * as React from "react"

export default function Label({
    className,
    children,
    required = false,
    htmlFor,
}: {
    className?: string
    children?: React.ReactNode
    required?: boolean
    htmlFor?: string
}) {
    return (
        <label
            className={`
font-medium tracking-wide
mb-1.5

text-muted-foreground
transition-all duration-200

peer-focus:text-foreground
peer-disabled:text-muted-foreground/60

group-data-[disabled=true]:pointer-events-none
group-data-[disabled=true]:opacity-50

${className}
`}
            aria-required={required}
            htmlFor={htmlFor}
        >
            {children}
            {required && (
                <span className="text-red-500">*</span>
            )}
        </label>
    )
}
