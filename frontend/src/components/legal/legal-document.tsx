import type {ReactNode} from "react"

export interface LegalSection {
    id: string
    title: string
    content: ReactNode
}

export default function LegalDocument({title, summary, effectiveDate, sections}: {
    title: string
    summary: string
    effectiveDate: string
    sections: LegalSection[]
}) {
    return (
        <div className="mx-auto grid w-full max-w-6xl gap-8 px-4 py-10 sm:px-6 sm:py-14 lg:grid-cols-[15rem_minmax(0,1fr)] lg:px-8">
            <aside className="hidden lg:block">
                <nav className="sticky top-8 rounded-2xl border bg-card p-3 shadow-xs" aria-label={`${title} contents`}>
                    <p className="px-3 pb-2 pt-1 text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">On this page</p>
                    <ul className="space-y-0.5">
                        {sections.map((section) => (
                            <li key={section.id}>
                                <a className="block rounded-lg px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground" href={`#${section.id}`}>
                                    {section.title}
                                </a>
                            </li>
                        ))}
                    </ul>
                </nav>
            </aside>

            <main className="min-w-0">
                <header className="border-b pb-8">
                    <p className="text-sm font-medium text-primary">Legal</p>
                    <h1 className="mt-2 text-4xl font-bold tracking-tight sm:text-5xl">{title}</h1>
                    <p className="mt-4 max-w-3xl text-base leading-7 text-muted-foreground sm:text-lg">{summary}</p>
                    <p className="mt-5 text-sm text-muted-foreground"><span className="font-medium text-foreground">Effective:</span> {effectiveDate}</p>
                </header>

                <article className="divide-y">
                    {sections.map((section, index) => (
                        <section className="scroll-mt-8 py-8" id={section.id} key={section.id}>
                            <div className="flex gap-4">
                                <span className="mt-1 hidden text-sm font-medium tabular-nums text-muted-foreground sm:block">{String(index + 1).padStart(2, "0")}</span>
                                <div className="min-w-0 flex-1">
                                    <h2 className="text-2xl font-semibold tracking-tight">{section.title}</h2>
                                    <div className="mt-4 space-y-4 text-[0.95rem] leading-7 text-muted-foreground [&_a]:font-medium [&_a]:text-foreground [&_a]:underline [&_a]:underline-offset-4 [&_h3]:pt-2 [&_h3]:text-base [&_h3]:font-semibold [&_h3]:text-foreground [&_li]:pl-1 [&_strong]:font-semibold [&_strong]:text-foreground [&_ul]:ml-5 [&_ul]:list-disc [&_ul]:space-y-2">
                                        {section.content}
                                    </div>
                                </div>
                            </div>
                        </section>
                    ))}
                </article>
            </main>
        </div>
    )
}
