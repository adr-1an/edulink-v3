"use client"

import {useRouter} from "next/navigation"
import {ArrowLeft} from "lucide-react"
import {Button} from "@/components/ui/button"

export default function BackButton() {
    const router = useRouter()
    return <Button variant="ghost" className="-ml-2" onClick={() => router.back()}><ArrowLeft /> Back</Button>
}
