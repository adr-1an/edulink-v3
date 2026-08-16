"use client"

import {ArrowLeft, RefreshCw} from "lucide-react"
import {useRouter} from "next/navigation"
import {Button} from "@/components/ui/button"
import {useLocale} from "@/i18n/provider"

export default function ErrorActions() {
    const router = useRouter()
    const {t} = useLocale()

    return (
        <>
            <Button variant="outline" onClick={() => router.back()}>
                <ArrowLeft />
                {t("error.goBack")}
            </Button>
            <Button onClick={() => router.refresh()}>
                <RefreshCw />
                {t("error.tryAgain")}
            </Button>
        </>
    )
}
