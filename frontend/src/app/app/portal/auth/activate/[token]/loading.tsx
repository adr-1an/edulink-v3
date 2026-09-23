import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <main className="flex min-h-screen items-center justify-center bg-background px-4 py-8 sm:px-6">
            <div className="w-full max-w-lg overflow-hidden rounded-2xl border bg-card">
                <div className="space-y-4 border-b px-5 py-6 sm:px-7 sm:py-7">
                    <Skeleton className="size-11 rounded-2xl" />
                    <Skeleton className="h-3 w-28" />
                    <Skeleton className="h-8 w-64 max-w-full" />
                    <Skeleton className="h-4 w-full" />
                </div>
                <div className="space-y-5 px-5 py-6 sm:px-7 sm:py-7">
                    <Skeleton className="h-16 w-full rounded-lg" />
                    <Skeleton className="h-16 w-full rounded-lg" />
                    <Skeleton className="h-11 w-full rounded-lg" />
                </div>
            </div>
        </main>
    )
}
