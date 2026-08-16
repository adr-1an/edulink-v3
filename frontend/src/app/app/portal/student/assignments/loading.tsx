import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="space-y-6">
            <section className="overflow-hidden rounded-3xl border bg-card">
                <div className="flex items-start gap-4 px-5 py-7 sm:px-7">
                    <Skeleton className="size-12 rounded-2xl" />
                    <div className="space-y-3">
                        <Skeleton className="h-4 w-28" />
                        <Skeleton className="h-8 w-52" />
                        <Skeleton className="h-4 w-72 max-w-full" />
                    </div>
                </div>
                <div className="grid border-t sm:grid-cols-3">
                    {Array.from({length: 3}).map((_, index) => (
                        <div className="flex items-center gap-3 border-t px-5 py-4 first:border-t-0 sm:border-l sm:border-t-0 sm:first:border-l-0" key={index}>
                            <Skeleton className="size-4 rounded-full" />
                            <div className="space-y-2">
                                <Skeleton className="h-6 w-8" />
                                <Skeleton className="h-3 w-20" />
                            </div>
                        </div>
                    ))}
                </div>
            </section>
            <Skeleton className="h-12 w-full rounded-2xl" />
            <div className="grid gap-3 lg:grid-cols-2">
                {Array.from({length: 4}).map((_, index) => (
                    <Skeleton className="h-64 rounded-2xl" key={index} />
                ))}
            </div>
        </div>
    )
}
