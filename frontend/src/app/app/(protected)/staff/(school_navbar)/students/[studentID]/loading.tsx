import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="mx-auto w-full max-w-6xl space-y-6">
            <Skeleton className="h-9 w-32" />
            <div className="flex items-center gap-4">
                <Skeleton className="size-16 rounded-full" />
                <div className="space-y-2">
                    <Skeleton className="h-9 w-64 max-w-full" />
                    <Skeleton className="h-5 w-48 max-w-full" />
                </div>
            </div>
            <div className="grid gap-5 lg:grid-cols-[minmax(0,1.35fr)_minmax(18rem,0.65fr)]">
                <div className="space-y-5">
                    <Skeleton className="h-64 rounded-2xl" />
                    <Skeleton className="h-48 rounded-2xl" />
                </div>
                <Skeleton className="h-52 rounded-2xl" />
            </div>
        </div>
    )
}
