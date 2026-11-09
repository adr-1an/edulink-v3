import type {Metadata} from "next"
import {PAGE_TITLES} from "@/lib/page_metadata"

export const metadata: Metadata = {title: PAGE_TITLES.verifyEmailChange}

export default function Layout({children}: {children: React.ReactNode}) {
    return children
}
