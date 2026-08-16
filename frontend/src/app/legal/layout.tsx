import type {ReactNode} from "react"
import Link from "next/link"
import {ArrowLeft} from "lucide-react"
import {Button} from "@/components/ui/button"

export default function Layout({children}: {children: ReactNode}) {
    return (
        <div className="min-h-screen bg-background text-foreground">
            <header className="border-b bg-background/95">
                <div className="mx-auto flex h-16 max-w-6xl items-center justify-between gap-4 px-4 sm:px-6 lg:px-8">
                    <Button className="-ml-3" size="sm" variant="ghost" render={<Link href="/" />}>
                        <ArrowLeft /> Back to Edulink
                    </Button>
                    <nav className="flex items-center gap-1" aria-label="Legal pages">
                        <Button size="sm" variant="ghost" render={<Link href="/legal/terms" />}>Terms</Button>
                        <Button size="sm" variant="ghost" render={<Link href="/legal/privacy" />}>Privacy</Button>
                    </nav>
                </div>
            </header>
            {children}
            <footer className="border-t">
                <div className="mx-auto flex max-w-6xl flex-col gap-2 px-4 py-8 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between sm:px-6 lg:px-8">
                    <p>© 2026 Vertex. Edulink is operated from Warsaw, Poland.</p>
                    <a className="hover:text-foreground" href="mailto:support@vertexapp.net">support@vertexapp.net</a>
                </div>
            </footer>
        </div>
    )
}
