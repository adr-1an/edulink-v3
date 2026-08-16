import {Skeleton} from "@/components/ui/skeleton"

export default function Loading() {
    return (
        <div className="mx-auto w-full max-w-6xl space-y-6">
            <Skeleton className="h-9 w-40" />
            <div className="space-y-3">
                <Skeleton className="h-6 w-28 rounded-full" />
                <Skeleton className="h-10 w-96 max-w-full" />
                <Skeleton className="h-5 w-72 max-w-full" />
            </div>
            <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_20rem]">
                <div className="space-y-5">
                    <Skeleton className="h-56 w-full rounded-2xl" />
                    <Skeleton className="h-72 w-full rounded-2xl" />
                </div>
                <Skeleton className="h-64 w-full rounded-2xl" />
            </div>
        </div>
    )
}
