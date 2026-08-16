"use client"

import {type LucideIcon} from "lucide-react"
import {Card, CardContent, CardDescription, CardHeader, CardTitle} from "@/components/ui/card"
import {cn} from "@/lib/utils"
import {useLocale} from "@/i18n/provider"

export default function AuthCard({
    icon: Icon,
    eyebrow,
    title,
    description,
    children,
    className,
    contentClassName,
}: {
    icon: LucideIcon
    eyebrow?: string
    title: string
    description: string
    children: React.ReactNode
    className?: string
    contentClassName?: string
}) {
    const {t} = useLocale()
    return (
        <Card className={cn("w-full max-w-lg overflow-hidden shadow-xl/10", className)}>
            <CardHeader className="gap-4 border-b bg-muted/20 px-5 py-6 sm:px-7 sm:py-7">
                <div className="flex size-11 items-center justify-center rounded-2xl bg-primary text-primary-foreground shadow-sm">
                    <Icon className="size-5" />
                </div>
                <div className="space-y-2">
                    <p className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">{eyebrow ?? t("auth.staffAccount")}</p>
                    <CardTitle className="text-2xl tracking-tight sm:text-3xl">{title}</CardTitle>
                    <CardDescription className="max-w-md text-sm leading-6">{description}</CardDescription>
                </div>
            </CardHeader>
            <CardContent className={cn("px-5 py-6 sm:px-7 sm:py-7", contentClassName)}>
                {children}
            </CardContent>
        </Card>
    )
}
