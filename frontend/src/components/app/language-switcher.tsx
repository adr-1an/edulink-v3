"use client"

import {useState} from "react"
import {Languages} from "lucide-react"
import {useRouter} from "next/navigation"
import {toast} from "sonner"
import {setLocale} from "@/i18n/actions"
import {type Locale} from "@/i18n/config"
import {useLocale} from "@/i18n/provider"
import {Select, SelectItem, SelectPopup, SelectTrigger, SelectValue} from "@/components/ui/select"
import {cn} from "@/lib/utils"

export default function LanguageSwitcher({className, compact = false}: {
    className?: string
    compact?: boolean
}) {
    const router = useRouter()
    const {locale, t} = useLocale()
    const [pending, setPending] = useState(false)
    const items = [
        {label: t("language.english"), value: "en"},
        {label: t("language.polish"), value: "pl"},
    ]

    async function changeLocale(value: string | null) {
        if (value !== "en" && value !== "pl" || value === locale || pending) return
        setPending(true)
        const result = await setLocale(value as Locale)
        if (!result.ok) {
            setPending(false)
            toast.error(t("language.error"))
            return
        }
        router.refresh()
    }

    return (
        <Select items={items} value={locale} onValueChange={changeLocale} disabled={pending}>
            <SelectTrigger
                className={cn(compact ? "min-w-28 max-sm:min-w-9 max-sm:w-9 max-sm:px-2" : "w-full", className)}
                aria-label={t("language.change")}
            >
                <Languages />
                <SelectValue className={compact ? "max-sm:sr-only" : undefined} />
            </SelectTrigger>
            <SelectPopup>
                {items.map((item) => (
                    <SelectItem value={item.value} key={item.value}>{item.label}</SelectItem>
                ))}
            </SelectPopup>
        </Select>
    )
}
