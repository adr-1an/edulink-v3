"use client"

import {useRouter} from "next/navigation"
import {ArrowLeft} from "lucide-react"
import {Button} from "@/components/ui/button"
import {useLocale} from "@/i18n/provider"

export default function BackButton() {
    const router = useRouter()
    const {t} = useLocale()
    return <Button variant="ghost" className="-ml-2" onClick={() => router.back()}><ArrowLeft /> {t("common.back")}</Button>
}
