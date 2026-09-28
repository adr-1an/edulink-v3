"use client"

import { motion } from "framer-motion"
import React, {useEffect} from "react"
import {toast, Toaster} from "sonner"
import Link from "next/link"
import Cookies from "js-cookie"
import {usePathname, useRouter} from "next/navigation"
import {Card, CardFooter} from "@/components/ui/card";
import {Button} from "@/components/ui/button";
import {useTheme} from "@/components/app/theme_provider";
import {useLocale} from "@/i18n/provider";
import {usesImmersiveAppShell} from "@/lib/app_shell";


export default function ClientLayout({ children }: { children: React.ReactNode }) {
    const router = useRouter()
    const pathname = usePathname()
    const {theme} = useTheme()
    const {t} = useLocale()
    const immersive = usesImmersiveAppShell(pathname)

    useEffect(() => {
        const success = Cookies.get("success")
        const pendingRedirect = Cookies.get("pendingRedirect")

        if (success) {
            Cookies.remove("success")
            toast.success(success)
        }

        if (pendingRedirect) {
            Cookies.remove("pendingRedirect")
            router.replace(pendingRedirect)
        }
    }, [router])

    return (
        <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
        >
            <div className="flex min-h-screen w-full flex-col bg-background text-foreground">
                <div className={immersive ? "flex-1" : "flex-1 p-2 md:p-8"}>
                    {children}
                </div>

                {!immersive && <Card className="w-full rounded-none border-x-0 border-b-0 bg-transparent text-center">
                    <CardFooter className="flex flex-wrap justify-center gap-1 pb-6">
                        <Button variant="link">
                            <Link href={"/"} target="_blank">{t("common.homepage")}</Link>
                        </Button>
                        <Button variant="link">
                            <Link href={"/legal/terms"} target="_blank">{t("common.terms")}</Link>
                        </Button>
                        <Button variant="link">
                            <Link href={"/legal/privacy"} target="_blank">{t("common.privacy")}</Link>
                        </Button>
                        <Button variant="link">
                            <Link href={"/contact"} target="_blank">{t("common.contact")}</Link>
                        </Button>
                    </CardFooter>
                </Card>}
            </div>

            <Toaster richColors visibleToasts={5} position="top-right" theme={theme} />
        </motion.div>
    )
}
