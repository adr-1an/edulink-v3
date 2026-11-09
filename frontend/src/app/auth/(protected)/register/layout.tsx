import type {Metadata} from "next"
import {formatPageTitle, PAGE_TITLES} from "@/lib/page_metadata"

export const metadata: Metadata = {
    title: {
        default: PAGE_TITLES.staffRegistration,
        template: formatPageTitle("%s"),
    },
}

export default function Layout({children}: {children: React.ReactNode}) {
    return children
}
