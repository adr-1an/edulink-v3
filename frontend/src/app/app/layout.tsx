import type {Metadata} from "next"
import {formatPageTitle, SITE_TITLE} from "@/lib/page_metadata"

export const metadata: Metadata = {
    title: {
        default: SITE_TITLE,
        template: formatPageTitle("%s"),
    },
}

export default function Layout({children}: {children: React.ReactNode}) {
    return children
}
